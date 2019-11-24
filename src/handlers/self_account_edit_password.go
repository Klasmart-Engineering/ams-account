package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

type editSelfAccountPasswordRequestBody struct {
	CurrentPassword string `json:"currPass"`
	NewPassword     string `json:"newPass"`
}

// HandleEditSelfAccountPassword handles requests of editing the password of the signed in account.
func HandleEditSelfAccountPassword(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
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

	// Validate the password
	err = globals.PasswordPolicyValidator.ValidatePassword(newPassword)
	if err != nil {
		return handlePasswordValidatorError(resp, err)
	}

	// Get the account information
	accountID := req.Session.Data.AccountID
	accInfo, err := globals.AccountDatabase.GetAccountSignInInfoByID(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accInfo == nil {
		return resp.SetClientError(apierrors.ErrorInvalidLogin)
	}

	// Verify that the current password is correct
	if !globals.PasswordHasher.VerifyPasswordHash(currentPassword, accInfo.PasswordHash) { // Verifies the password
		logger.LogFormat("[EDITACCOUNTPW] An edit password request for account [%s] with the incorrect current password from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidPassword)
	}

	// Generate the password hash
	extraSecure := (accInfo.AdminRole > 0)
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(newPassword, extraSecure)
	if err != nil {
		return resp.SetServerError(err)
	}

	// Change the password in the database
	err = globals.AccountDatabase.EditAccount(accountID, &accountdatabase.AccountEditInfo{
		PasswordHash: &hashedPassword,
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	logger.LogFormat("[EDITACCOUNTPW] A successful edit account password request for account [%s]\n", accountID)

	// Resets the flag that this account must set a new password
	if accounts.AccountMustSetPassword(accInfo.Flags) {
		err = globals.AccountDatabase.RemoveAccountFlags(accountID, accounts.MustSetPasswordFlag)
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
