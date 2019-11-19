package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
	"bitbucket.org/calmisland/go-server-utils/emailutils"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/phoneutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
	"github.com/google/uuid"
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

const (
	defaultCountryCode  = "XX"
	defaultLanguageCode = "en_US"

	signUpVerificationCodeByteLength = 4
)

// HandleSignUp handles sign-up requests.
func HandleSignUp(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody signUpRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	userEmail := reqBody.Email
	userPhoneNumber := phoneutils.CleanPhoneNumber(reqBody.PhoneNumber)
	userPassword := reqBody.Password
	userLanguage := textutils.SanitizeString(reqBody.Language)
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	var isUsingEmail bool
	if len(userEmail) > 0 {
		// Validate parameters
		if !emailutils.IsValidEmailAddressFormat(userEmail) {
			logger.LogFormat("[SIGNUP] A sign-up request for account [%s] with invalid email address from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("email"))
		} else if !emailutils.IsValidEmailAddressHost(userEmail) {
			logger.LogFormat("[SIGNUP] A sign-up request for account [%s] with invalid email host from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("email"))
		}

		// There should not be an email and a phone number at the same time
		userPhoneNumber = ""
		isUsingEmail = true
	} else if len(userPhoneNumber) > 0 {
		if !phoneutils.IsValidPhoneNumber(userPhoneNumber) {
			logger.LogFormat("[SIGNUP] A sign-up request for account [%s] with invalid phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		}

		// There should not be an email and a phone number at the same time
		userEmail = ""
		isUsingEmail = false
	} else {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("email"))
	}

	// Validate the password
	err = globals.PasswordPolicyValidator.ValidatePassword(userPassword)
	if err != nil {
		return handlePasswordValidatorError(resp, err)
	}

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	if isUsingEmail {
		// Check if the email is already used by another account
		accountExists, err := accountDB.AccountExistsWithEmail(userEmail)
		if err != nil {
			return resp.SetServerError(err)
		} else if accountExists {
			logger.LogFormat("[SIGNUP] A sign-up request for already existing account [%s] email from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorEmailAlreadyUsed)
		}
	} else {
		// Check if the phone number is already used by another account
		accountExists, err := accountDB.AccountExistsWithPhoneNumber(userPhoneNumber)
		if err != nil {
			return resp.SetServerError(err)
		} else if accountExists {
			logger.LogFormat("[SIGNUP] A sign-up request for already existing account [%s] phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorPhoneNumberAlreadyUsed)
		}
	}

	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(userPassword, false)
	if err != nil {
		return resp.SetServerError(err)
	}

	verificationCode, err := securitycodes.GenerateSecurityCode(signUpVerificationCodeByteLength)
	if err != nil {
		return resp.SetServerError(err)
	}

	accountUUID, err := uuid.NewRandom()
	if err != nil {
		return resp.SetServerError(err)
	}

	geoIPResult, err := globals.GeoIPService.GetCountryFromIP(clientIP)
	if err != nil {
		return resp.SetServerError(err)
	}

	countryCode := defaultCountryCode
	if geoIPResult != nil && len(geoIPResult.CountryCode) > 0 {
		countryCode = geoIPResult.CountryCode
	}

	// Sets the default language if none is set
	if !langutils.IsValidLanguageCode(userLanguage) {
		userLanguage = defaultLanguageCode
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
			Template: &messagetemplates.EmailVerificationTemplate{
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
		return resp.SetServerError(err)
	}

	var emailVerificationCode string
	var phoneNumberVerificationCode string
	if isUsingEmail {
		emailVerificationCode = verificationCode
	} else {
		phoneNumberVerificationCode = verificationCode
	}

	err = accountDB.CreateAccount(&accountdatabase.CreateAccountInfo{
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
		return resp.SetServerError(err)
	}

	logger.LogFormat("[SIGNUP] A successful sign-up request for account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

	response := signUpResponseBody{
		AccountID: accountID,
	}
	resp.SetBody(&response)
	return nil
}

func handlePasswordValidatorError(resp *apirequests.Response, err error) error {
	switch err.(type) {
	case *passwords.PasswordTooShortError:
		passwordErr := err.(*passwords.PasswordTooShortError)
		return resp.SetClientError(apierrors.ErrorPasswordTooShort.WithValue(int64(passwordErr.MinimumLength)))
	case *passwords.PasswordTooLongError:
		passwordErr := err.(*passwords.PasswordTooLongError)
		return resp.SetClientError(apierrors.ErrorPasswordTooLong.WithValue(int64(passwordErr.MaximumLength)))
	case *passwords.PasswordLowerCaseMissingError:
		passwordErr := err.(*passwords.PasswordLowerCaseMissingError)
		return resp.SetClientError(apierrors.ErrorPasswordLowerCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordUpperCaseMissingError:
		passwordErr := err.(*passwords.PasswordUpperCaseMissingError)
		return resp.SetClientError(apierrors.ErrorPasswordUpperCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordNumberMissingError:
		passwordErr := err.(*passwords.PasswordNumberMissingError)
		return resp.SetClientError(apierrors.ErrorPasswordNumberMissing.WithValue(int64(passwordErr.MinimumCount)))
	default:
		return resp.SetServerError(err)
	}
}
