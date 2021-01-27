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
	"bitbucket.org/calmisland/go-server-utils/emailutils"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/phoneutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
)

const (
	forgotPasswordRegularVerificationCodeByteLength = 4
	forgotPasswordAdminVerificationCodeByteLength   = 10
)

type forgotPasswordRequestBody struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNr"`
	Language    string `json:"lang"`
}

// HandleForgotPassword handles forgot password requests.
func HandleForgotPassword(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody forgotPasswordRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	userEmail := reqBody.Email
	userPhoneNumber := reqBody.PhoneNumber
	userLanguage := textutils.SanitizeString(reqBody.Language)
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	var isUsingEmail bool
	if len(userEmail) > 0 {
		// Validate parameters
		if !emailutils.IsValidEmailAddressFormat(userEmail) {
			return resp.SetClientError(apierrors.ErrorInvalidEmailFormat)
		} else if !emailutils.IsValidEmailAddressHost(userEmail) {
			return resp.SetClientError(apierrors.ErrorInvalidEmailHost)
		}

		// There should not be an email and a phone number at the same time
		userPhoneNumber = ""
		isUsingEmail = true
	} else if len(userPhoneNumber) > 0 {
		userPhoneNumber, err = phoneutils.CleanPhoneNumber(userPhoneNumber)
		if err != nil {
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		} else if !phoneutils.IsValidPhoneNumber(userPhoneNumber) {
			logger.LogFormat("[FORGETPW] A request to recover from a forgotten password for account [%s] with invalid phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		}

		// There should not be an email and a phone number at the same time
		userEmail = ""
		isUsingEmail = false
	} else {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("email"))
	}

	// Sets the default language if none is set
	if !langutils.IsValidLanguageCode(userLanguage) {
		userLanguage = handlers_common.DefaultLanguageCode
	}

	var accountID string
	var foundAccount bool
	if isUsingEmail {
		// Get the account ID from the email
		accountID, foundAccount, err = globals.AccountDatabase.GetAccountIDFromEmail(userEmail)
		if err != nil {
			return resp.SetServerError(err)
		}
	} else {
		// Get the account ID from the phone number
		accountID, foundAccount, err = globals.AccountDatabase.GetAccountIDFromPhoneNumber(userPhoneNumber)
		if err != nil {
			return resp.SetServerError(err)
		}
	}

	// Then get the account information
	var accInfo *accountdatabase.AccountSignInInfo
	if foundAccount {
		accInfo, err = globals.AccountDatabase.GetAccountSignInInfoByID(accountID)
		if err != nil {
			return resp.SetServerError(err)
		} else if accInfo != nil && !accounts.IsAccountVerified(accInfo.Flags) {
			return resp.SetClientError(apierrors.ErrorEmailNotVerified)
		}
	} else {
		return resp.SetClientError(apierrors.ErrorAccountNotFound)
	}

	if accInfo != nil {
		logger.LogFormat("[FORGETPW] A request to recover from a forgotten password received for existing account [%s] from IP [%s] with UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

		verificationCodeByteCount := forgotPasswordRegularVerificationCodeByteLength
		if accInfo.AdminRole > 0 {
			verificationCodeByteCount = forgotPasswordAdminVerificationCodeByteLength
		}
		verificationCode, err := securitycodes.GenerateSecurityCode(verificationCodeByteCount)
		if err != nil {
			return resp.SetServerError(err)
		}

		err = globals.AccountDatabase.CreateAccountVerification(accInfo.ID, accountdatabase.VerificationTypePassword, verificationCode)
		if err != nil {
			return resp.SetServerError(err)
		}

		// Override the user language based on the database record
		userLanguage = accInfo.Language
		if len(userLanguage) == 0 {
			userLanguage = handlers_common.DefaultLanguageCode
		}

		template := &messagetemplates.PasswordResetTemplate{
			Code: verificationCode,
		}

		if isUsingEmail {
			err = sendForgotPasswordEmailFound(userEmail, userLanguage, template)
			if err != nil {
				return resp.SetServerError(err)
			}
		} else {
			userPhoneNumber = accInfo.PhoneNumber
			err = sendForgotPasswordSMSFound(userPhoneNumber, userLanguage, template)
			if err != nil {
				return resp.SetServerError(err)
			}
		}
	} else {
		if isUsingEmail {
			logger.LogFormat("[FORGETPW] A request to recover from a forgotten password received for non-existing account [%s] from IP [%s] with UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			err = sendForgotPasswordEmailNotFound(userEmail, userLanguage)
			if err != nil {
				return resp.SetServerError(err)
			}
		} else {
			// NOTE: We don't send anything to unknown phone numbers
			logger.LogFormat("[FORGETPW] A request to recover from a forgotten password received for non-existing account [%s] from IP [%s] with UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
		}
	}

	return nil
}

func sendForgotPasswordEmailFound(email, language string, template *messagetemplates.PasswordResetTemplate) error {
	emailMessage := &messages.Message{
		MessageType: messages.MessageTypeEmail,
		Priority:    messages.MessagePriorityEmailHigh,
		Recipient:   email,
		Language:    language,
		Template:    template,
	}
	return globals.MessageSendQueue.EnqueueMessage(emailMessage)
}

func sendForgotPasswordSMSFound(phoneNumber, language string, template *messagetemplates.PasswordResetTemplate) error {
	emailMessage := &messages.Message{
		MessageType: messages.MessageTypeSMS,
		Priority:    messages.MessagePrioritySMSTransactional,
		Recipient:   phoneNumber,
		Language:    language,
		Template:    template,
	}
	return globals.MessageSendQueue.EnqueueMessage(emailMessage)
}

func sendForgotPasswordEmailNotFound(email, language string) error {
	emailMessage := &messages.Message{
		MessageType: messages.MessageTypeEmail,
		Priority:    messages.MessagePriorityEmailNormal,
		Recipient:   email,
		Language:    language,
		Template:    &messagetemplates.PasswordResetFailTemplate{},
	}
	return globals.MessageSendQueue.EnqueueMessage(emailMessage)
}