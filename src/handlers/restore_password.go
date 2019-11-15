package handlers

import (
	"context"
	"log"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
)

type restorePasswordRequestBody struct {
	AccountID          string `json:"accountId"`
	AccountEmail       string `json:"accountEmail"`
	AccountPhoneNumber string `json:"accountPhoneNr"`
	VerificationCode   string `json:"verificationCode"`
	Password           string `json:"pw"`
	PartnerID          int32  `json:"partnerId"`
}

// HandleRestorePassword handles password restore requests.
func HandleRestorePassword(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody restorePasswordRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 && len(reqBody.AccountEmail) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	} else if len(reqBody.VerificationCode) == 0 || len(reqBody.Password) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	verificationCode := reqBody.VerificationCode
	password := reqBody.Password
	partnerID := accounts.GetPartnerID(reqBody.PartnerID)
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	// Make sure that the partner ID is valid
	if !accounts.IsPartnerValid(partnerID) {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	if len(accountID) == 0 {
		if len(reqBody.AccountEmail) > 0 {
			// Get the account ID from the email
			accountEmail := reqBody.AccountEmail
			accountIDResult, foundAccount, err := accountDB.GetAccountIDFromEmail(accountEmail, partnerID)
			if err != nil {
				return resp.SetServerError(err)
			} else if !foundAccount {
				log.Printf("[RESTOREPW] A restore password request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountEmail, clientIP, clientUserAgent)
				return resp.SetClientError(apierrors.ErrorItemNotFound)
			}

			accountID = accountIDResult
		} else if len(reqBody.AccountPhoneNumber) > 0 {
			// Get the account ID from the email
			accountPhoneNumber := reqBody.AccountPhoneNumber
			accountIDResult, foundAccount, err := accountDB.GetAccountIDFromPhoneNumber(accountPhoneNumber)
			if err != nil {
				return resp.SetServerError(err)
			} else if !foundAccount {
				log.Printf("[RESTOREPW] A restore password request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountPhoneNumber, clientIP, clientUserAgent)
				return resp.SetClientError(apierrors.ErrorItemNotFound)
			}

			accountID = accountIDResult
		} else {
			return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("accountEmail"))
		}
	}

	verificationInfo, err := accountDB.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil || verificationInfo.VerificationCodes.Password == nil {
		log.Printf("[RESTOREPW] A restore password request for account [%s] without a forgot password request from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorItemNotFound)
	} else if !securitycodes.ValidateSecurityCode(*verificationInfo.VerificationCodes.Password, verificationCode) {
		log.Printf("[RESTOREPW] A restore password request for account [%s] with incorrect password verification code from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	}

	// Validate the password
	err = globals.PasswordPolicyValidator.ValidatePassword(password)
	if err != nil {
		return handlePasswordValidatorError(resp, err)
	}

	log.Printf("[RESTOREPW] A successful restore password request for account [%s] using a forgot password request from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	// Generate the password hash
	extraSecure := (verificationInfo.AdminRole > 0)
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(password, extraSecure)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Change the password
	err = accountDB.EditAccount(accountID, &accountdatabase.AccountEditInfo{
		PasswordHash: &hashedPassword,
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	// Remove the verification code
	err = accountDB.RemoveAccountVerification(accountID, accountdatabase.VerificationTypePassword)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Resets the flag that this account must set a new password
	if accounts.AccountMustSetPassword(verificationInfo.Flags) {
		err = accountDB.RemoveAccountFlags(accountID, accounts.MustSetPasswordFlag)
		if err != nil {
			return resp.SetServerError(err)
		}
	}

	userEmail := verificationInfo.Email
	userLanguage := verificationInfo.Language

	// Sets the default language if none is set
	if len(userLanguage) == 0 {
		userLanguage = defaultLanguageCode
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
			return resp.SetServerError(err)
		}
	}

	return nil
}
