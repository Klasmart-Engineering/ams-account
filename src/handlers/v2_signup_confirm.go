package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/src/services/account_jwt_service"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-utils/emailutils"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/phoneutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
	"github.com/google/uuid"
)

type signUpByTokenRequestBody struct {
	VerificationToken string `json:"verificationToken"`
	VerificationCode  string `json:"verificationCode"`
}

type signUpByTokenResponseBody struct {
	AccountID string `json:"accountId"`
}

// HandleSignUp handles sign-up requests.
func HandleSignUpConfirm(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody signUpByTokenRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	verificationToken := reqBody.VerificationToken
	verificationCode := reqBody.VerificationCode
	claims, errVerify := account_jwt_service.VerifyToken(verificationToken)
	if errVerify != nil {
		return resp.SetServerError(errVerify)
	}

	// Verify that the current password is correct
	if !globals.PasswordHasher.VerifyPasswordHash(verificationCode, claims.VerificationCode) { // Verifies the password
		logger.LogFormat("[SIGNUP CONFIRM] Verification Code [%s] does not match\n", verificationCode)
		return resp.SetClientError(apierrors.ErrorInvalidPassword)
	}

	errClaim := claims.Valid()
	if errClaim != nil {
		return resp.SetServerError(errClaim)
	}

	userEmail := claims.Email
	userPhoneNumber := claims.PhoneNumber
	userPassword := claims.Password
	userLanguage := textutils.SanitizeString(claims.Language)
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
		userPhoneNumber, err = phoneutils.CleanPhoneNumber(userPhoneNumber)
		if err != nil {
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		} else if !phoneutils.IsValidPhoneNumber(userPhoneNumber) {
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

	if isUsingEmail {
		// Check if the email is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithEmail(userEmail)
		if err != nil {
			return resp.SetServerError(err)
		} else if accountExists {
			logger.LogFormat("[SIGNUP] A sign-up request for already existing account [%s] email from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorEmailAlreadyUsed)
		}
	} else {
		// Check if the phone number is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithPhoneNumber(userPhoneNumber)
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

	err = globals.AccountDatabase.CreateAccount(&accountdatabase.CreateAccountInfo{
		ID:           accountID,
		Email:        userEmail,
		PhoneNumber:  userPhoneNumber,
		PasswordHash: hashedPassword,
		Flags:        0,
		Country:      countryCode,
		Language:     userLanguage,
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	logger.LogFormat("[SIGNUP] A successful sign-up request for account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

	response := signUpByTokenResponseBody{
		AccountID: accountID,
	}
	resp.SetBody(&response)
	return nil
}
