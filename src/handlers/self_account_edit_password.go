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
)

type editSelfAccountPasswordRequestBody struct {
	CurrentPassword string `json:"currPass"`
	NewPassword     string `json:"newPass"`
}

// HandleEditSelfAccountPassword handles requests of editing the password of the signed in account.
func HandleEditSelfAccountPassword(ctx context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody editSelfAccountPasswordRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	currentPassword := reqBody.CurrentPassword
	newPassword := reqBody.NewPassword
	if len(currentPassword) == 0 || len(newPassword) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	// Validate the password
	err = globals.PasswordPolicyValidator.ValidatePassword(newPassword)
	if err != nil {
		return handlePasswordValidatorError(resp, err)
	}

	// Get the account information
	accountID := req.Session.Data.AccountID
	accInfo, err := accountDB.GetAccountSignInInfoByID(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accInfo == nil {
		return resp.SetClientError(apierrors.ErrorInvalidLogin)
	}

	// NOTE: If the account is marked to need a new password, we don't have to verify the old password
	needsNewPassword := accounts.AccountMustSetPassword(accInfo.Flags)
	if !needsNewPassword && !globals.PasswordHasher.VerifyPasswordHash(currentPassword, accInfo.PasswordHash) { // Verifies the password
		log.Printf("[EDITACCOUNTPW] An edit password request for account [%s] with the incorrect current password from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidPassword)
	}

	// Generate the password hash
	extraSecure := (accInfo.AdminRole > 0)
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(newPassword, extraSecure)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Change the password in the database
	err = accountDB.EditAccount(accountID, &accountdatabase.AccountEditInfo{
		PasswordHash: &hashedPassword,
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	log.Printf("[EDITACCOUNTPW] A successful edit account password request for account [%s]\n", accountID)

	// Resets the flag that this account must set a new password
	if needsNewPassword {
		err = accountDB.RemoveAccountFlags(accountID, accounts.MustSetPasswordFlag)
		if err != nil {
			return resp.SetServerError(err)
		}
	}

	userEmail := accInfo.Email
	userLanguage := accInfo.Language

	// Sets the default language if none is set
	if len(userLanguage) == 0 {
		userLanguage = defaultLanguageCode
	}

	// Sends an email about the change
	emailMessage := &emailqueue.EmailMessage{
		RecipientEmail: userEmail,
		Language:       userLanguage,
		TemplateName:   emailtemplates.ChangedPasswordTemplate,
	}
	err = globals.EmailSendQueue.EnqueueEmail(emailMessage)
	if err != nil {
		return resp.SetServerError(err)
	}

	return nil
}