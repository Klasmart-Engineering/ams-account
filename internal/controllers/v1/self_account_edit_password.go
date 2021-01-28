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
	"github.com/labstack/echo/v4"
)

type editSelfAccountPasswordRequestBody struct {
	CurrentPassword string `json:"currPass"`
	NewPassword     string `json:"newPass"`
}

// HandleEditSelfAccountPassword handles requests of editing the password of the signed in account.
func HandleEditSelfAccountPassword(c echo.Context) error {
	// Parse the request body
	reqBody := new(editSelfAccountPasswordRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return echoadapter.SetClientError(c, apierrors.ErrorBadRequestBody)
	}

	cc := c.(*echoadapter.AuthContext)
	accountID := cc.Session.Data.AccountID
	req := c.Request()
	clientIP := net.ParseIP(c.RealIP())
	clientUserAgent := req.UserAgent()

	currentPassword := reqBody.CurrentPassword
	newPassword := reqBody.NewPassword
	if len(currentPassword) == 0 || len(newPassword) == 0 {
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidParameters)
	}

	// Validate the password
	err = globals.PasswordPolicyValidator.ValidatePassword(newPassword)
	if err != nil {
		return echoadapter.HandlePasswordValidatorError(c, err)
	}

	// Get the account information
	accInfo, err := globals.AccountDatabase.GetAccountSignInInfoByID(accountID)
	if err != nil {
		return err
	} else if accInfo == nil {
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidLogin)
	}

	// Verify that the current password is correct
	if !globals.PasswordHasher.VerifyPasswordHash(currentPassword, accInfo.PasswordHash) { // Verifies the password
		logger.LogFormat("[EDITACCOUNTPW] An edit password request for account [%s] with the incorrect current password from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidPassword)
	}

	// Generate the password hash
	extraSecure := (accInfo.AdminRole > 0)
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(newPassword, extraSecure)
	if err != nil {
		return err
	}

	// Change the password in the database
	err = globals.AccountDatabase.EditAccount(accountID, &accountdatabase.AccountEditInfo{
		PasswordHash: &hashedPassword,
	})
	if err != nil {
		return err
	}

	logger.LogFormat("[EDITACCOUNTPW] A successful edit account password request for account [%s]\n", accountID)

	// Resets the flag that this account must set a new password
	if accounts.AccountMustSetPassword(accInfo.Flags) {
		err = globals.AccountDatabase.RemoveAccountFlags(accountID, accounts.MustSetPasswordFlag)
		if err != nil {
			return err
		}
	}

	userEmail := accInfo.Email
	userLanguage := accInfo.Language

	// Sets the default language if none is set
	if len(userLanguage) == 0 {
		userLanguage = defs.DefaultLanguageCode
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
			return err
		}
	}

	return cc.NoContent(http.StatusOK)
}