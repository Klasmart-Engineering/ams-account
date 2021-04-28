package v2

import (
	"net"
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/defs"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/services/account_jwt_service"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/services/accountverificationservice"
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
	"github.com/labstack/echo/v4"
)

type verifyCodeRequestBody struct {
	Email        string `json:"email"`
	PhoneNumber  string `json:"phoneNr"`
	Password     string `json:"pw"`
	Language     string `json:"lang"`
	TemplateName string `json:"template,omitempty"`
}

type verifyCodeResponseBody struct {
	VerificationToken string `json:"verificationToken"`
}

// HandleSignUp handles sign-up requests.
func HandleSignupRequest(c echo.Context) error {
	reqBody := new(verifyCodeRequestBody)
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

	// Sets the default language if none is set
	if !langutils.IsValidLanguageCode(userLanguage) {
		userLanguage = defs.DefaultLanguageCode
	}

	token, errToken := account_jwt_service.CreateToken(&account_jwt_service.TokenMapClaims{
		Email:            userEmail,
		PhoneNumber:      userPhoneNumber,
		Password:         hashedPassword,
		Language:         userLanguage,
		VerificationCode: verificationCode,
	})

	verificationLink := accountverificationservice.GetVerificationLinkByToken(token, verificationCode, userLanguage)
	var message *messages.Message
	if isUsingEmail {
		var template messages.MessageTemplate

		if reqBody.TemplateName == "kidsloop" {
			template = &messagetemplates.EmailVerificationKidsloopTemplate{
				Code: verificationCode,
			}
		} else {
			template = &messagetemplates.EmailVerificationTemplate{
				Code: verificationCode,
				Link: verificationLink,
			}
		}

		message = &messages.Message{
			MessageType: messages.MessageTypeEmail,
			Priority:    messages.MessagePriorityEmailHigh,
			Recipient:   userEmail,
			Language:    userLanguage,
			Template:    template,
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

	logger.LogFormat("[VERIFICATION] A successful verification request from IP [%s] UserAgent [%s]\n", clientIP, clientUserAgent)
	logger.LogFormat("[VERIFICATION] Created Verification Code: %s\n", verificationCode)

	if errToken != nil {
		return helpers.HandleInternalError(c, err)
	}

	response := verifyCodeResponseBody{
		VerificationToken: token,
	}

	return c.JSON(http.StatusOK, response)
}
