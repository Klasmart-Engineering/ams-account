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

type resendPhoneNumberVerificationRequestBody struct {
	AccountID string `json:"accountId"`
}

// HandleResendPhoneNumberVerification handles requests for resending phone number verifications.
func HandleResendPhoneNumberVerification(c echo.Context) error {
	// Parse the request body
	reqBody := new(resendPhoneNumberVerificationRequestBody)
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
		logger.LogFormat("[RESENDVERIFY] A resend phone number verification request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorVerificationNotFound)
	}

	if accounts.IsAccountPhoneNumberVerified(verificationInfo.Flags) {
		logger.LogFormat("[RESENDVERIFY] A resend phone number verification request for account [%s] that was already verified from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorAlreadyVerified)
	}

	logger.LogFormat("[RESENDVERIFY] A successful resend phone number verification request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	// Generate a new verification code
	verificationCode, err := securitycodes.GenerateSecurityCode(defs.SignUpVerificationCodeByteLength)
	if err != nil {
		return err
	}

	// Create the account verification in the database
	err = globals.AccountDatabase.CreateAccountVerification(accountID, accountdatabase.VerificationTypePhoneNumber, verificationCode)
	if err != nil {
		return err
	}

	// Re-send the verification SMS
	userPhoneNumber := verificationInfo.PhoneNumber
	userLanguage := verificationInfo.Language
	smsMessage := &messages.Message{
		MessageType: messages.MessageTypeSMS,
		Priority:    messages.MessagePrioritySMSTransactional,
		Recipient:   userPhoneNumber,
		Language:    userLanguage,
		Template: &messagetemplates.PhoneVerificationTemplate{
			Code: verificationCode,
		},
	}
	err = globals.MessageSendQueue.EnqueueMessage(smsMessage)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
