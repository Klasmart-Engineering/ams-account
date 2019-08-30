package handlers

import (
	"context"
	"log"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-emails/emailqueue"
	"bitbucket.org/calmisland/go-server-emails/emailtemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
)

type resendEmailVerificationRequestBody struct {
	AccountID string `json:"accountId"`
}

// HandleResendEmailVerification handles requests for resending email verifications.
func HandleResendEmailVerification(ctx context.Context, req *apirequests.Request, resp *apirequests.Response) error {
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

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	verificationInfo, err := accountDB.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil {
		log.Printf("[RESENDVERIFY] A resend verification request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorVerificationNotFound)
	}

	if accounts.IsAccountVerified(verificationInfo.Flags) {
		log.Printf("[RESENDVERIFY] A resend email verification request for account [%s] that was already verified from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorVerificationNotFound)
	}

	log.Printf("[RESENDVERIFY] A successful resend email verification request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	// Generate a new verification code
	verificationCode, err := securitycodes.GenerateSecurityCode(4)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Create the account verification in the database
	err = accountDB.CreateAccountVerification(accountID, accountdatabase.VerificationTypeEmail, verificationCode)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Re-send the verification email
	userEmail := verificationInfo.Email
	userLanguage := verificationInfo.Language
	emailMessage := &emailqueue.EmailMessage{
		RecipientEmail: userEmail,
		Language:       userLanguage,
		TemplateName:   emailtemplates.EmailVerificationTemplate,
		TemplateData: struct {
			Code string
		}{
			Code: verificationCode,
		},
	}
	err = globals.EmailSendQueue.EnqueueEmail(emailMessage)
	if err != nil {
		return resp.SetServerError(err)
	}

	return nil
}
