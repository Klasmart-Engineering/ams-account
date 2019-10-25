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
	"bitbucket.org/calmisland/go-server-utils/emailutils"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
)

type forgotPasswordRequestBody struct {
	User      string `json:"user"`
	Language  string `json:"lang"`
	PartnerID int32  `json:"partnerId"`
}

// HandleForgotPassword handles forgot password requests.
func HandleForgotPassword(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody forgotPasswordRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	userEmail := reqBody.User
	userLanguage := textutils.SanitizeString(reqBody.Language)
	partnerID := accounts.GetPartnerID(reqBody.PartnerID)
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	// Make sure that the partner ID is valid
	if !accounts.IsPartnerValid(partnerID) {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	// Validate parameters
	if !emailutils.IsValidEmailAddressFormat(userEmail) {
		return resp.SetClientError(apierrors.ErrorInvalidEmailFormat)
	} else if !emailutils.IsValidEmailAddressHost(userEmail) {
		return resp.SetClientError(apierrors.ErrorInvalidEmailHost)
	}

	// Sets the default language if none is set
	if !langutils.IsValidLanguageCode(userLanguage) {
		userLanguage = defaultLanguageCode
	}

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	// First get the account ID from the email
	accountID, err := accountDB.GetAccountID(userEmail, partnerID)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Then get the account information
	var accInfo *accountdatabase.AccountSignInInfo
	if accountID != nil {
		accInfo, err = accountDB.GetAccountSignInInfoByID(*accountID)
		if err != nil {
			return resp.SetServerError(err)
		} else if accInfo != nil && !accounts.IsAccountVerified(accInfo.Flags) {
			return resp.SetClientError(apierrors.ErrorEmailNotVerified)
		}
	}

	if accInfo != nil {
		log.Printf("[FORGETPW] A request to recover from a forgotten password received for existing account [%s] from IP [%s] with UserAgent [%s]\n", *accountID, clientIP, clientUserAgent)

		verificationCodeByteCount := 4
		if accInfo.AdminRole > 0 {
			verificationCodeByteCount = 8
		}
		verificationCode, err := securitycodes.GenerateSecurityCode(verificationCodeByteCount)
		if err != nil {
			return resp.SetServerError(err)
		}

		err = accountDB.CreateAccountVerification(accInfo.ID, accountdatabase.VerificationTypePassword, verificationCode)
		if err != nil {
			return resp.SetServerError(err)
		}

		// Override the user language based on the database record
		userLanguage = accInfo.Language
		if len(userLanguage) == 0 {
			userLanguage = defaultLanguageCode
		}

		userEmail = accInfo.Email
		err = sendForgotPasswordEmailFound(userEmail, userLanguage, verificationCode)
	} else {
		log.Printf("[FORGETPW] A request to recover from a forgotten password received for non-existing account [%s] from IP [%s] with UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

		err = sendForgotPasswordEmailNotFound(userEmail, userLanguage)
	}

	if err != nil {
		return resp.SetServerError(err)
	}

	return nil
}

func sendForgotPasswordEmailFound(email, language, verificationCode string) error {
	emailMessage := &emailqueue.EmailMessage{
		RecipientEmail: email,
		Language:       language,
		TemplateName:   emailtemplates.PasswordResetTemplate,
		TemplateData: struct {
			Code string
		}{
			Code: verificationCode,
		},
	}
	return globals.EmailSendQueue.EnqueueEmail(emailMessage)
}

func sendForgotPasswordEmailNotFound(email, language string) error {
	emailMessage := &emailqueue.EmailMessage{
		RecipientEmail: email,
		Language:       language,
		TemplateName:   emailtemplates.PasswordResetFailTemplate,
	}
	return globals.EmailSendQueue.EnqueueEmail(emailMessage)
}
