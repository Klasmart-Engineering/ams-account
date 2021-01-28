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
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
	"github.com/labstack/echo/v4"
)

type verifyPhoneNumberRequestBody struct {
	AccountID        string `json:"accountId"`
	VerificationCode string `json:"verificationCode"`
}

type verifyPhoneNumberResponseBody struct {
	PhoneNumber string `json:"phoneNr"`
}

// HandleVerifyPhoneNumber handles account phone number verification requests.
func HandleVerifyPhoneNumber(c echo.Context) error {
	// Parse the request body
	reqBody := new(verifyPhoneNumberRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return echoadapter.SetClientError(c, apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 || len(reqBody.VerificationCode) == 0 {
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	verificationCode := reqBody.VerificationCode
	req := c.Request()
	clientIP := net.ParseIP(c.RealIP())
	clientUserAgent := req.UserAgent()

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return err
	} else if verificationInfo == nil {
		logger.LogFormat("[VERIFY] A verify phone number request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidVerificationCode)
	} else if accounts.IsAccountPhoneNumberVerified(verificationInfo.Flags) {
		return echoadapter.SetClientError(c, apierrors.ErrorAlreadyVerified)
	}

	if verificationInfo.VerificationCodes.PhoneNumber == nil {
		logger.LogFormat("[VERIFY] A phone number verify request for account [%s] without pending phone number verification from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidVerificationCode)
	} else if !securitycodes.ValidateSecurityCode(*verificationInfo.VerificationCodes.PhoneNumber, verificationCode) {
		logger.LogFormat("[VERIFY] A phone number verify request for account [%s] with incorrect verification code from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return echoadapter.SetClientError(c, apierrors.ErrorInvalidVerificationCode)
	}

	err = globals.AccountDatabase.SetAccountFlags(accountID, accounts.IsAccountVerifiedFlag|accounts.IsAccountPhoneNumberVerifiedFlag)
	if err != nil {
		return err
	}

	logger.LogFormat("[VERIFY] A successful phone number verify request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	userPhoneNumber := verificationInfo.PhoneNumber
	userLanguage := verificationInfo.Language
	if len(userLanguage) == 0 {
		userLanguage = defs.DefaultLanguageCode
	}

	// TODO: Do we want to send a welcome SMS?

	// Remove the verification code
	err = globals.AccountDatabase.RemoveAccountVerification(accountID, accountdatabase.VerificationTypePhoneNumber)
	if err != nil {
		return err
	}

	response := verifyPhoneNumberResponseBody{
		PhoneNumber: userPhoneNumber,
	}
	return c.JSON(http.StatusOK, response)
}
