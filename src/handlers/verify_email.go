package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/src/handlers/handlers_common"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
)

type verifyEmailRequestBody struct {
	AccountID        string `json:"accountId"`
	VerificationCode string `json:"verificationCode"`
}

type verifyEmailResponseBody struct {
	Email string `json:"email"`
}

// HandleVerifyEmail handles account email verification requests.
func HandleVerifyEmail(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody verifyEmailRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 || len(reqBody.VerificationCode) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	verificationCode := reqBody.VerificationCode
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil {
		logger.LogFormat("[VERIFY] A verify email request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	} else if accounts.IsAccountEmailVerified(verificationInfo.Flags) {
		return resp.SetClientError(apierrors.ErrorAlreadyVerified)
	}

	if verificationInfo.VerificationCodes.Email == nil {
		logger.LogFormat("[VERIFY] An email verify request for account [%s] without pending email verification from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	} else if !securitycodes.ValidateSecurityCode(*verificationInfo.VerificationCodes.Email, verificationCode) {
		logger.LogFormat("[VERIFY] An email verify request for account [%s] with incorrect verification code from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	}

	err = globals.AccountDatabase.SetAccountFlags(accountID, accounts.IsAccountVerifiedFlag|accounts.IsAccountEmailVerifiedFlag)
	if err != nil {
		return resp.SetServerError(err)
	}

	logger.LogFormat("[VERIFY] A successful email verify request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	userEmail := verificationInfo.Email
	userLanguage := verificationInfo.Language
	if len(userLanguage) == 0 {
		userLanguage = handlers_common.DefaultLanguageCode
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
		return resp.SetServerError(err)
	}

	// Remove the verification code
	err = globals.AccountDatabase.RemoveAccountVerification(accountID, accountdatabase.VerificationTypeEmail)
	if err != nil {
		return resp.SetServerError(err)
	}

	response := verifyEmailResponseBody{
		Email: userEmail,
	}
	resp.SetBody(&response)
	return nil
}
