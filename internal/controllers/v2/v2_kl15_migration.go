package v2

import (
	"net"
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/defs"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
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
	"github.com/labstack/echo/v4"
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

// HandleKl15Migration handles sign-up requests.
func HandleKl15Migration(c echo.Context) error {
	// Parse the request body
	reqBody := new(kl15MigrationRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorBadRequestBody)
	}

	userEmail := reqBody.Email
	userPhoneNumber := reqBody.PhoneNumber
	userPassword := reqBody.Password
	userLanguage := reqBody.Language
	req := c.Request()
	clientIP := net.ParseIP(c.RealIP())
	clientUserAgent := req.UserAgent()

	var isUsingEmail bool
	if len(userEmail) > 0 {
		// Validate parameters
		if !emailutils.IsValidEmailAddressFormat(userEmail) {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for account [%s] with invalid email address from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("email"))
		} else if !emailutils.IsValidEmailAddressHost(userEmail) {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for account [%s] with invalid email host from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
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
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for account [%s] with invalid phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorInputInvalidFormat.WithField("phoneNr"))
		}

		// There should not be an email and a phone number at the same time
		userEmail = ""
		isUsingEmail = false
	} else {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters.WithField("email"))
	}

	kl15AccInfo, found, err := globals.AccountDatabase.GetAccountsMigrationKl1dot5InfoFromEmail(userEmail)
	if !found || err != nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorAccountNotFound)
	}

	// Check Migration status.
	if kl15AccInfo.MigrationStatus == accountdatabase.AccountsKl1dot5MigrationStatusDone {
		return c.JSON(http.StatusOK, &kl15MigrationResponseBody{
			Status: "ok",
		})
	}

	// Check password
	isPassed := passwords.ValidateSha3_512_Password(userPassword, &passwords.Sha3_512_Hash{
		Hash:   kl15AccInfo.PwHash,
		Secret: kl15AccInfo.PwHashSecret,
	})

	// Check if this account is in AMS without migration status
	accountExists, err := globals.AccountDatabase.AccountExistsWithEmail(userEmail)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	// Override 1.5 password if account is in AMS AND Password is correct AND MigrationStatus is not done
	if accountExists && isPassed && kl15AccInfo.MigrationStatus != accountdatabase.AccountsKl1dot5MigrationStatusDone {
		return overridePasswordToExistingAccount(c, userEmail, userPassword)
	}

	if !isPassed {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidPassword)
	}

	// generate AMS hashed password
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(userPassword, false)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	// NOTE: Password validator skipped because KL 1.5 password policy is unknown

	var flags int32 = 0
	if isUsingEmail {
		// Check if the email is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithEmail(userEmail)
		if err != nil {
			return helpers.HandleInternalError(c, err)
		} else if accountExists {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for already existing account [%s] email from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorEmailAlreadyUsed)
		}
		flags = int32(accounts.IsAccountVerifiedFlag | accounts.IsAccountEmailVerifiedFlag)
	} else {
		// Check if the phone number is already used by another account
		accountExists, err := globals.AccountDatabase.AccountExistsWithPhoneNumber(userPhoneNumber)
		if err != nil {
			return helpers.HandleInternalError(c, err)
		} else if accountExists {
			logger.LogFormat("[KL1.5-MIGRATION] A sign-up request for already existing account [%s] phone number from IP [%s] UserAgent [%s]\n", userPhoneNumber, clientIP, clientUserAgent)
			return apirequests.EchoSetClientError(c, apierrors.ErrorPhoneNumberAlreadyUsed)
		}
		flags = int32(accounts.IsAccountVerifiedFlag | accounts.IsAccountPhoneNumberVerifiedFlag)
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
		return helpers.HandleInternalError(c, err)
	}

	logger.LogFormat("[KL1.5-MIGRATION] A successful sign-up request for account [%s] from IP [%s] UserAgent [%s]\n", userEmail, clientIP, clientUserAgent)

	err = globals.AccountDatabase.SetAccountsMigrationKl1dot5MigrationStatus(userEmail, accountdatabase.AccountsKl1dot5MigrationStatusDone)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	response := kl15MigrationResponseBody{
		Status: "ok",
	}

	return c.JSON(http.StatusOK, response)
}

func overridePasswordToExistingAccount(c echo.Context, userEmail string, userPassword string) error {
	//override and return
	accountIDResult, _, err := globals.AccountDatabase.GetAccountIDFromEmail(userEmail)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}
	// generate AMS hashed password
	hashedPassword, err := globals.PasswordHasher.GeneratePasswordHash(userPassword, false)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	// Change the password
	err = globals.AccountDatabase.EditAccount(accountIDResult, &accountdatabase.AccountEditInfo{
		PasswordHash: &hashedPassword,
	})
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	// update migration Status to done
	err = globals.AccountDatabase.SetAccountsMigrationKl1dot5MigrationStatus(userEmail, accountdatabase.AccountsKl1dot5MigrationStatusDone)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	response := kl15MigrationResponseBody{
		Status: "ok",
	}

	logger.LogFormat("[KL1.5-MIGRATION] Override password using 1.5 [%s] \n", userEmail)
	return c.JSON(http.StatusOK, response)
}
