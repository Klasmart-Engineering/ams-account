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

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	verificationInfo, err := accountDB.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil {
		log.Printf("[VERIFY] A verify request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	}

	if verificationInfo.VerificationCodes.Email == nil {
		log.Printf("[VERIFY] An email verify request for account [%s] without pending email verification from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	} else if !securitycodes.ValidateSecurityCode(*verificationInfo.VerificationCodes.Email, verificationCode) {
		log.Printf("[VERIFY] An email verify request for account [%s] with incorrect verification code from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	}

	err = accountDB.SetAccountFlags(accountID, accounts.IsAccountVerifiedFlag)
	if err != nil {
		return resp.SetServerError(err)
	}

	log.Printf("[VERIFY] A successful email verify request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	userEmail := verificationInfo.Email
	userLanguage := verificationInfo.Language
	if len(userLanguage) == 0 {
		userLanguage = defaultLanguageCode
	}

	// Send the welcome email
	emailMessage := &emailqueue.EmailMessage{
		RecipientEmail: userEmail,
		Language:       userLanguage,
		TemplateName:   emailtemplates.WelcomeTemplate,
	}
	err = globals.EmailSendQueue.EnqueueEmail(emailMessage)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Remove the verification code
	err = accountDB.RemoveAccountVerification(accountID, accountdatabase.VerificationTypeEmail)
	if err != nil {
		return resp.SetServerError(err)
	}

	response := verifyEmailResponseBody{
		Email: userEmail,
	}
	resp.SetBody(&response)
	return nil
}
