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

type resendEmailVerificationRequestBody struct {
	AccountID string `json:"accountId"`
}

// HandleResendEmailVerification handles requests for resending email verifications.
func HandleResendEmailVerification(c echo.Context) error {
	// Parse the request body
	reqBody := new(resendEmailVerificationRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return echoadapter.SetClientError(c, apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 {
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	req := c.Request()
	clientIP := net.ParseIP(c.RealIP())
	clientUserAgent := req.UserAgent()

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return err
	} else if verificationInfo == nil {
		logger.LogFormat("[RESENDVERIFY] A resend email verification request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorVerificationNotFound)
	}

	if accounts.IsAccountEmailVerified(verificationInfo.Flags) {
		logger.LogFormat("[RESENDVERIFY] A resend email verification request for account [%s] that was already verified from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorAlreadyVerified)
	}

	logger.LogFormat("[RESENDVERIFY] A successful resend email verification request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	// Generate a new verification code
	verificationCode, err := securitycodes.GenerateSecurityCode(defs.SignUpVerificationCodeByteLength)
	if err != nil {
		return err
	}

	// Create the account verification in the database
	err = globals.AccountDatabase.CreateAccountVerification(accountID, accountdatabase.VerificationTypeEmail, verificationCode)
	if err != nil {
		return err
	}

	// Re-send the verification email
	userEmail := verificationInfo.Email
	userLanguage := verificationInfo.Language
	verificationLink := globals.AccountVerificationService.GetVerificationLink(accountID, verificationCode, userLanguage)
	emailMessage := &messages.Message{
		MessageType: messages.MessageTypeEmail,
		Priority:    messages.MessagePriorityEmailHigh,
		Recipient:   userEmail,
		Language:    userLanguage,
		Template: &messagetemplates.EmailVerificationTemplate{
			Code: verificationCode,
			Link: verificationLink,
		},
	}
	err = globals.MessageSendQueue.EnqueueMessage(emailMessage)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
