package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/pkg/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/handlers/handlers_common"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
)

type resendEmailVerificationRequestBody struct {
	AccountID string `json:"accountId"`
}

// HandleResendEmailVerification handles requests for resending email verifications.
func HandleResendEmailVerification(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody resendEmailVerificationRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil {
		logger.LogFormat("[RESENDVERIFY] A resend email verification request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorVerificationNotFound)
	}

	if accounts.IsAccountEmailVerified(verificationInfo.Flags) {
		logger.LogFormat("[RESENDVERIFY] A resend email verification request for account [%s] that was already verified from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorAlreadyVerified)
	}

	logger.LogFormat("[RESENDVERIFY] A successful resend email verification request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	// Generate a new verification code
	verificationCode, err := securitycodes.GenerateSecurityCode(handlers_common.SignUpVerificationCodeByteLength)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Create the account verification in the database
	err = globals.AccountDatabase.CreateAccountVerification(accountID, accountdatabase.VerificationTypeEmail, verificationCode)
	if err != nil {
		return resp.SetServerError(err)
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
		return resp.SetServerError(err)
	}

	return nil
}
