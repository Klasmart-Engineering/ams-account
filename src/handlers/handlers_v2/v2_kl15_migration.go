package handlers_v2

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/src/handlers/handlers_common"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"bitbucket.org/calmisland/go-server-utils/emailutils"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/phoneutils"
	"github.com/google/uuid"
)

type kl15MigrationRequestBody struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNr"`
	Password    string `json:"pw"`
	Language    string `json:"lang"`
}

type kl15MigrationResponseBody struct {
	Status string `json:"status"`
}

// HandleSignUp handles sign-up requests.
func HandleKl15Migration(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody kl15MigrationRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	userEmail := reqBody.Email
	userPhoneNumber := reqBody.PhoneNumber
	userPassword := reqBody.Password
	userLanguage := reqBody.Language
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	var isUsingEmail bool
	if len(userEmail) > 0 {
		// Validate parameters
		if !emailutils.IsValidEmailAddressFormat(userEmail) {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for account [%s] with invalid email address from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("email"))
		} else if !emailutils.IsValidEmailAddressHost(userEmail) {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for account [%s] with invalid email host from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
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
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for account [%s] with invalid phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		}

		// There should not be an email and a phone number at the same time
		userEmail = ""
		isUsingEmail = false
	} else {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("email"))
	}

	kl15AccInfo, found, err := globals.AccountDatabase.GetAccountsMigrationKl1dot5InfoFromEmail(userEmail)
	if !found || err != nil {
		return resp.SetClientError(apierrors.ErrorAccountNotFound)
	}

	// Check Migration status.
	if kl15AccInfo.MigrationStatus == accountdatabase.AccountsKl1dot5MigrationStatusDone {
		resp.SetBody(&kl15MigrationResponseBody{
			Status: "ok",
		})
		return nil
	}
	// Check password
	isPassed := passwords.ValidateSha3_512_Password(userPassword, &passwords.Sha3_512_Hash{
		Hash:   kl15AccInfo.PwHash,
		Secret: kl15AccInfo.PwHashSecret,
	})
	if !isPassed {
		return resp.SetClientError(apierrors.ErrorInvalidPassword)
	}

	// generate AMS hashed password
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(userPassword, false)
	if err != nil {
		return resp.SetServerError(err)
	}

	// NOTE: Password validator skipped because KL 1.5 password policy is unknown

	var flags int32 = 0
	if isUsingEmail {
		// Check if the email is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithEmail(userEmail)
		if err != nil {
			return resp.SetServerError(err)
		} else if accountExists {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for already existing account [%s] email from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorEmailAlreadyUsed)
		}
		flags = int32(accounts.IsAccountVerifiedFlag | accounts.IsAccountEmailVerifiedFlag)
	} else {
		// Check if the phone number is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithPhoneNumber(userPhoneNumber)
		if err != nil {
			return resp.SetServerError(err)
		} else if accountExists {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for already existing account [%s] phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return resp.SetClientError(apierrors.ErrorPhoneNumberAlreadyUsed)
		}
		flags = int32(accounts.IsAccountVerifiedFlag | accounts.IsAccountPhoneNumberVerifiedFlag)
	}

	accountUUID, err := uuid.NewRandom()
	if err != nil {
		return resp.SetServerError(err)
	}

	geoIPResult, err := globals.GeoIPService.GetCountryFromIP(clientIP)
	if err != nil {
		return resp.SetServerError(err)
	}

	countryCode := handlers_common.DefaultCountryCode
	if geoIPResult != nil && len(geoIPResult.CountryCode) > 0 {
		countryCode = geoIPResult.CountryCode
	}

	// Sets the default language if none is set
	if !langutils.IsValidLanguageCode(userLanguage) {
		userLanguage = handlers_common.DefaultLanguageCode
	}

	accountID := accountUUID.String()

	err = globals.AccountDatabase.CreateAccount(&accountdatabase.CreateAccountInfo{
		ID:           accountID,
		Email:        userEmail,
		PhoneNumber:  userPhoneNumber,
		PasswordHash: hashedPassword,
		Flags:        flags,
		Country:      countryCode,
		Language:     userLanguage,
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	logger.LogFormat("[KL1.5-MIGRATION] A successful sign-up request for account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

	err = globals.AccountDatabase.SetAccountsMigrationKl1dot5MigrationStatus(userEmail, accountdatabase.AccountsKl1dot5MigrationStatusDone)
	if err != nil {
		return resp.SetServerError(err)
	}

	response := kl15MigrationResponseBody{
		Status: "ok",
	}
	resp.SetBody(&response)
	return nil
}
