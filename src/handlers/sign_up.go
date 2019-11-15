package handlers

import (
	"context"
	"fmt"
	"log"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-emails/emailqueue"
	"bitbucket.org/calmisland/go-server-emails/emailtemplates"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
	"bitbucket.org/calmisland/go-server-utils/emailutils"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
	"github.com/google/uuid"
)

type signUpRequestBody struct {
	User      string `json:"user"`
	Password  string `json:"pw"`
	Language  string `json:"lang"`
	PartnerID int32  `json:"partnerId"`
}

type signUpResponseBody struct {
	AccountID string `json:"accountId"`
}

const (
	defaultCountryCode  = "XX"
	defaultLanguageCode = "en_US"
)

// HandleSignUp handles sign-up requests.
func HandleSignUp(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody signUpRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	userEmail := reqBody.User
	userPassword := reqBody.Password
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
		log.Printf("[SIGNUP] A sign-up request for account [%s] with invalid email address from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidEmailFormat)
	} else if !emailutils.IsValidEmailAddressHost(userEmail) {
		log.Printf("[SIGNUP] A sign-up request for account [%s] with invalid email host from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidEmailHost)
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

	// Check if the email is already used by another account
	accountExists, err := accountDB.AccountExists(userEmail, partnerID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accountExists {
		log.Printf("[SIGNUP] A sign-up request for already existing account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorEmailAlreadyUsed)
	}

	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(userPassword, false)
	if err != nil {
		return resp.SetServerError(err)
	}

	verificationCode, err := securitycodes.GenerateSecurityCode(4)
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

	// Send the verification email
	emailMessage := &emailqueue.EmailMessage{
		RecipientEmail: userEmail,
		Language:       userLanguage,
		TemplateName:   emailtemplates.EmailVerificationTemplate,
		TemplateData: struct {
			Code string
			Link string
		}{
			Code: verificationCode,
			Link: fmt.Sprintf("http://localhost:8080/#/verify?accountId={%s}&code={%s}", accountID, verificationCode),
		},
	}
	err = globals.EmailSendQueue.EnqueueEmail(emailMessage)
	if err != nil {
		return resp.SetServerError(err)
	}

	err = accountDB.CreateAccount(&accountdatabase.CreateAccountInfo{
		ID:                    accountID,
		Email:                 userEmail,
		PasswordHash:          hashedPassword,
		Flags:                 0,
		EmailVerificationCode: verificationCode,
		Country:               countryCode,
		Language:              userLanguage,
	}, partnerID)
	if err != nil {
		return resp.SetServerError(err)
	}

	log.Printf("[SIGNUP] A successful sign-up request for account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

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
