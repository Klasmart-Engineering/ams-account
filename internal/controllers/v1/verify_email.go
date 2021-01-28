package v1

import (
	"net"
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/defs"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/echoadapter"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
	"github.com/labstack/echo/v4"
)

type verifyEmailRequestBody struct {
	AccountID        string `json:"accountId"`
	VerificationCode string `json:"verificationCode"`
}

type verifyEmailResponseBody struct {
	Email string `json:"email"`
}

// HandleVerifyEmail handles account email verification requests.
func HandleVerifyEmail(c echo.Context) error {
	// Parse the request body
	reqBody := new(verifyEmailRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return echoadapter.SetClientError(c, apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 || len(reqBody.VerificationCode) == 0 {
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	verificationCode := reqBody.VerificationCode
	req := c.Request()
	clientIP := net.ParseIP(c.RealIP())
	clientUserAgent := req.UserAgent

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return err
	} else if verificationInfo == nil {
		logger.LogFormat("[VERIFY] A verify email request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidVerificationCode)
	} else if accounts.IsAccountEmailVerified(verificationInfo.Flags) {
		return echoadapter.SetClientError(c, apierrors.ErrorAlreadyVerified)
	}

	if verificationInfo.VerificationCodes.Email == nil {
		logger.LogFormat("[VERIFY] An email verify request for account [%s] without pending email verification from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidVerificationCode)
	} else if !securitycodes.ValidateSecurityCode(*verificationInfo.VerificationCodes.Email, verificationCode) {
		logger.LogFormat("[VERIFY] An email verify request for account [%s] with incorrect verification code from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidVerificationCode)
	}

	err = globals.AccountDatabase.SetAccountFlags(accountID, accounts.IsAccountVerifiedFlag|accounts.IsAccountEmailVerifiedFlag)
	if err != nil {
		return err
	}

	logger.LogFormat("[VERIFY] A successful email verify request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	userEmail := verificationInfo.Email
	userLanguage := verificationInfo.Language
	if len(userLanguage) == 0 {
		userLanguage = defs.DefaultLanguageCode
	}

	// Send the welcome email
	emailMessage := &messages.Message{
		MessageType: messages.MessageTypeEmail,
		Priority:    messages.MessagePriorityEmailNormal,
		Recipient:   userEmail,
		Language:    userLanguage,
		Template:    &messagetemplates.WelcomeTemplate{},
	}
	err = globals.MessageSendQueue.EnqueueMessage(emailMessage)
	if err != nil {
		return err
	}

	// Remove the verification code
	err = globals.AccountDatabase.RemoveAccountVerification(accountID, accountdatabase.VerificationTypeEmail)
	if err != nil {
		return err
	}

	response := verifyEmailResponseBody{
		Email: userEmail,
	}
	return c.JSON(http.StatusOK, response)
}
