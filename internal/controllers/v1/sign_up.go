package v1

import (
	"net"
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/defs"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
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
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type signUpRequestBody struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNr"`
	Password    string `json:"pw"`
	Language    string `json:"lang"`
}

type signUpResponseBody struct {
	AccountID string `json:"accountId"`
}

// HandleSignUp handles sign-up requests.
func HandleSignUp(c echo.Context) error {
	// Parse the request body
	reqBody := new(signUpRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorBadRequestBody)
	}

	userEmail := reqBody.Email
	userPhoneNumber := reqBody.PhoneNumber
	userPassword := reqBody.Password
	userLanguage := textutils.SanitizeString(reqBody.Language)
	req := c.Request()
	clientIP := net.ParseIP(c.RealIP())
	clientUserAgent := req.UserAgent()

	var isUsingEmail bool
	if len(userEmail) > 0 {
		// Validate parameters
		if !emailutils.IsValidEmailAddressFormat(userEmail) {
			logger.LogFormat("[SIGNUP] A sign-up request for account [%s] with invalid email address from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("email"))
		} else if !emailutils.IsValidEmailAddressHost(userEmail) {
			logger.LogFormat("[SIGNUP] A sign-up request for account [%s] with invalid email host from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("email"))
		}

		// There should not be an email and a phone number at the same time
		userPhoneNumber = ""
		isUsingEmail = true
	} else if len(userPhoneNumber) > 0 {
		userPhoneNumber, err = phoneutils.CleanPhoneNumber(userPhoneNumber)
		if err != nil {
			return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		} else if !phoneutils.IsValidPhoneNumber(userPhoneNumber) {
			logger.LogFormat("[SIGNUP] A sign-up request for account [%s] with invalid phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		}

		// There should not be an email and a phone number at the same time
		userEmail = ""
		isUsingEmail = false
	} else {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters.WithField("email"))
	}

	// Validate the password
	err = globals.PasswordPolicyValidator.ValidatePassword(userPassword)
	if err != nil {
		return defs.HandlePasswordValidatorError(c, err)
	}

	if isUsingEmail {
		// Check if the email is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithEmail(userEmail)
		if err != nil {
			return helpers.HandleInternalError(c, err)
		} else if accountExists {
			logger.LogFormat("[SIGNUP] A sign-up request for already existing account [%s] email from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorEmailAlreadyUsed)
		}
	} else {
		// Check if the phone number is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithPhoneNumber(userPhoneNumber)
		if err != nil {
			return helpers.HandleInternalError(c, err)
		} else if accountExists {
			logger.LogFormat("[SIGNUP] A sign-up request for already existing account [%s] phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorPhoneNumberAlreadyUsed)
		}
	}

	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(userPassword, false)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	verificationCode, err := securitycodes.GenerateSecurityCode(defs.SignUpVerificationCodeByteLength)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	accountUUID, err := uuid.NewRandom()
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	geoIPResult, err := globals.GeoIPService.GetCountryFromIP(clientIP)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	countryCode := defs.DefaultCountryCode
	if geoIPResult != nil && len(geoIPResult.CountryCode) > 0 {
		countryCode = geoIPResult.CountryCode
	}

	// Sets the default language if none is set
	if !langutils.IsValidLanguageCode(userLanguage) {
		userLanguage = defs.DefaultLanguageCode
	}

	accountID := accountUUID.String()
	verificationLink := globals.AccountVerificationService.GetVerificationLink(accountID, verificationCode, userLanguage)
	var message *messages.Message
	if isUsingEmail {
		message = &messages.Message{
			MessageType: messages.MessageTypeEmail,
			Priority:    messages.MessagePriorityEmailHigh,
			Recipient:   userEmail,
			Language:    userLanguage,
			Template: &messagetemplates.EmailVerificationLnpTemplate{
				Code: verificationCode,
				Link: verificationLink,
			},
		}
	} else {
		message = &messages.Message{
			MessageType: messages.MessageTypeSMS,
			Priority:    messages.MessagePrioritySMSTransactional,
			Recipient:   userPhoneNumber,
			Language:    userLanguage,
			Template: &messagetemplates.PhoneVerificationTemplate{
				Code: verificationCode,
			},
		}
	}

	err = globals.MessageSendQueue.EnqueueMessage(message)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	var emailVerificationCode string
	var phoneNumberVerificationCode string
	if isUsingEmail {
		emailVerificationCode = verificationCode
	} else {
		phoneNumberVerificationCode = verificationCode
	}

	err = globals.AccountDatabase.CreateAccount(&accountdatabase.CreateAccountInfo{
		ID:                          accountID,
		Email:                       userEmail,
		PhoneNumber:                 userPhoneNumber,
		PasswordHash:                hashedPassword,
		Flags:                       0,
		EmailVerificationCode:       emailVerificationCode,
		PhoneNumberVerificationCode: phoneNumberVerificationCode,
		Country:                     countryCode,
		Language:                    userLanguage,
	})
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	logger.LogFormat("[SIGNUP] A successful sign-up request for account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

	response := signUpResponseBody{
		AccountID: accountID,
	}
	return c.JSON(http.StatusOK, response)
}
