package v2

import (
	"net"
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/defs"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/services/account_jwt_service"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/messages"
	"bitbucket.org/calmisland/go-server-messages/messagetemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-utils/emailutils"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/phoneutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
)

type signUpByTokenRequestBody struct {
	VerificationToken string `json:"verificationToken"`
	VerificationCode  string `json:"verificationCode"`
}

type signUpByTokenResponseBody struct {
	AccountID string `json:"accountId"`
}

// HandleSignUp handles sign-up requests.
func HandleSignUpConfirm(c echo.Context) error {
	// Parse the request body
	reqBody := new(signUpByTokenRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorBadRequestBody)
	}

	verificationToken := reqBody.VerificationToken
	verificationCode := reqBody.VerificationCode

	claims, errVerify := account_jwt_service.VerifyToken(verificationToken)

	if errVerify != nil {
		errStr := errVerify.Error()
		if errStr == "Token is expired." { // jwt module returns this text
			return apirequests.EchoSetClientError(c, apierrors.ErrorExpiredVerificationToken)
		} else if errStr == "signature is invalid" {
			return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidSignature)
		}
		return errVerify
	}

	// Verify that the current password is correct
	if defs.EnsureTestVerificationCode(verificationCode) == true {
		// pass it
		logger.LogFormat("[SIGNUP CONFIRM] USE_TEST_VERIFICATION_CODE and verification code matches %s\n", defs.TEST_VERIFICATION_CODE)
	} else if !globals.PasswordHasher.VerifyPasswordHash(verificationCode, claims.VerificationCode) { // Verifies the password
		logger.LogFormat("[SIGNUP CONFIRM] Verification Code [%s] does not match\n", verificationCode)
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidVerificationCode)
	}

	errClaim := claims.Valid()
	if errClaim != nil {
		return errClaim
	}

	userEmail := claims.Email
	userPhoneNumber := claims.PhoneNumber
	userPassword := claims.Password
	userLanguage := textutils.SanitizeString(claims.Language)
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

	var flags int32 = 0
	if isUsingEmail {
		// Check if the email is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithEmail(userEmail)
		if err != nil {
			return err
		} else if accountExists {
			logger.LogFormat("[SIGNUP] A sign-up request for already existing account [%s] email from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorEmailAlreadyUsed)
		}
		flags = int32(accounts.IsAccountVerifiedFlag | accounts.IsAccountEmailVerifiedFlag)
	} else {
		// Check if the phone number is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithPhoneNumber(userPhoneNumber)
		if err != nil {
			return err
		} else if accountExists {
			logger.LogFormat("[SIGNUP] A sign-up request for already existing account [%s] phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorPhoneNumberAlreadyUsed)
		}
		flags = int32(accounts.IsAccountVerifiedFlag | accounts.IsAccountPhoneNumberVerifiedFlag)
	}

	accountUUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	geoIPResult, err := globals.GeoIPService.GetCountryFromIP(clientIP)
	if err != nil {
		return err
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

	err = globals.AccountDatabase.CreateAccount(&accountdatabase.CreateAccountInfo{
		ID:           accountID,
		Email:        userEmail,
		PhoneNumber:  userPhoneNumber,
		PasswordHash: userPassword,
		Flags:        flags,
		Country:      countryCode,
		Language:     userLanguage,
	})
	if err != nil {
		return err
	}

	logger.LogFormat("[SIGNUP] A successful sign-up request for account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

	// send welcome email
	if len(userEmail) > 0 {
		emailMessage := &messages.Message{
			MessageType: messages.MessageTypeEmail,
			Priority:    messages.MessagePriorityEmailNormal,
			Recipient:   userEmail,
			Language:    userLanguage,
			Template:    &messagetemplates.WelcomeTemplate{},
		}
		err = globals.MessageSendQueue.EnqueueMessage(emailMessage)
		if err != nil {
			return err
		}
	}

	response := signUpByTokenResponseBody{
		AccountID: accountID,
	}

	return c.JSON(http.StatusOK, response)
}
