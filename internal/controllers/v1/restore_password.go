package v1

import (
	"net"
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/defs"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
	"bitbucket.org/calmisland/go-server-utils/phoneutils"
	"github.com/labstack/echo/v4"
)

type restorePasswordRequestBody struct {
	AccountID          string `json:"accountId"`
	AccountEmail       string `json:"accountEmail"`
	AccountPhoneNumber string `json:"accountPhoneNr"`
	VerificationCode   string `json:"verificationCode"`
	Password           string `json:"pw"`
}

// HandleRestorePassword handles password restore requests.
func HandleRestorePassword(c echo.Context) error {
	// Parse the request body
	reqBody := new(restorePasswordRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 && len(reqBody.AccountEmail) == 0 && len(reqBody.AccountPhoneNumber) == 0 {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters)
	} else if len(reqBody.VerificationCode) == 0 || len(reqBody.Password) == 0 {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	verificationCode := reqBody.VerificationCode
	password := reqBody.Password
	req := c.Request()
	clientIP := net.ParseIP(c.RealIP())
	clientUserAgent := req.UserAgent()

	if len(accountID) == 0 {
		if len(reqBody.AccountEmail) > 0 {
			// Get the account ID from the email
			accountEmail := reqBody.AccountEmail
			accountIDResult, foundAccount, err := globals.AccountDatabase.GetAccountIDFromEmail(accountEmail)
			if err != nil {
				return err
			} else if !foundAccount {
				logger.LogFormat("[RESTOREPW] A restore password request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountEmail, clientIP, clientUserAgent)
				return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidVerificationCode)
			}

			accountID = accountIDResult
		} else if len(reqBody.AccountPhoneNumber) > 0 {
			accountPhoneNumber, err := phoneutils.CleanPhoneNumber(reqBody.AccountPhoneNumber)
			if err != nil {
				return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("accountPhoneNr"))
			} else if !phoneutils.IsValidPhoneNumber(accountPhoneNumber) {
				logger.LogFormat("[RESTOREPW] A restore password request for account [%s] with invalid phone number from IP [%s] UserAgent [%s]\n", accountPhoneNumber, clientIP, clientUserAgent)
				return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("accountPhoneNr"))
			}

			// Get the account ID from the email
			accountIDResult, foundAccount, err := globals.AccountDatabase.GetAccountIDFromPhoneNumber(accountPhoneNumber)
			if err != nil {
				return err
			} else if !foundAccount {
				logger.LogFormat("[RESTOREPW] A restore password request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountPhoneNumber, clientIP, clientUserAgent)
				return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidVerificationCode)
			}

			accountID = accountIDResult
		} else {
			return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters.WithField("accountEmail"))
		}
	}

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return err
	} else if verificationInfo == nil || verificationInfo.VerificationCodes.Password == nil {
		logger.LogFormat("[RESTOREPW] A restore password request for account [%s] without a forgot password request from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return apirequests.EchoSetClientError(c, apierrors.ErrorItemNotFound)
	} else if !securitycodes.ValidateSecurityCode(*verificationInfo.VerificationCodes.Password, verificationCode) {
		logger.LogFormat("[RESTOREPW] A restore password request for account [%s] with incorrect password verification code from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidVerificationCode)
	}

	// Validate the password
	err = globals.PasswordPolicyValidator.ValidatePassword(password)
	if err != nil {
		return defs.HandlePasswordValidatorError(c, err)
	}

	logger.LogFormat("[RESTOREPW] A successful restore password request for account [%s] using a forgot password request from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	// Generate the password hash
	extraSecure := (verificationInfo.AdminRole > 0)
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(password, extraSecure)
	if err != nil {
		return err
	}

	// Change the password
	err = globals.AccountDatabase.EditAccount(accountID, &accountdatabase.AccountEditInfo{
		PasswordHash: &hashedPassword,
	})
	if err != nil {
		return err
	}

	// Remove the verification code
	err = globals.AccountDatabase.RemoveAccountVerification(accountID, accountdatabase.VerificationTypePassword)
	if err != nil {
		return err
	}

	// Resets the flag that this account must set a new password
	if accounts.AccountMustSetPassword(verificationInfo.Flags) {
		err = globals.AccountDatabase.RemoveAccountFlags(accountID, accounts.MustSetPasswordFlag)
		if err != nil {
			return err
		}
	}

	userEmail := verificationInfo.Email
	userLanguage := verificationInfo.Language

	// Sets the default language if none is set
	if len(userLanguage) == 0 {
		userLanguage = defs.DefaultLanguageCode
	}

	// TODO: Do we want to send SMS for this if there is no available email address?
	if len(userEmail) > 0 {
		// Sends an email about the change
		emailMessage := &messages.Message{
			MessageType: messages.MessageTypeEmail,
			Priority:    messages.MessagePriorityEmailHigh,
			Recipient:   userEmail,
			Language:    userLanguage,
			Template:    &messagetemplates.ChangedPasswordTemplate{},
		}
		err = globals.MessageSendQueue.EnqueueMessage(emailMessage)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
