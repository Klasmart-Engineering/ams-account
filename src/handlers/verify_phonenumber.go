package handlers

import (
	"context"

	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/securitycodes"
)

type verifyPhoneNumberRequestBody struct {
	AccountID        string `json:"accountId"`
	VerificationCode string `json:"verificationCode"`
}

type verifyPhoneNumberResponseBody struct {
	PhoneNumber string `json:"phoneNr"`
}

// HandleVerifyPhoneNumber handles account phone number verification requests.
func HandleVerifyPhoneNumber(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody verifyPhoneNumberRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.AccountID) == 0 || len(reqBody.VerificationCode) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	accountID := reqBody.AccountID
	verificationCode := reqBody.VerificationCode
	clientIP := req.SourceIP
	clientUserAgent := req.UserAgent

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	verificationInfo, err := accountDB.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil {
		logger.LogFormat("[VERIFY] A verify phone number request for non-existing account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	} else if accounts.IsAccountPhoneNumberVerified(verificationInfo.Flags) {
		return resp.SetClientError(apierrors.ErrorAlreadyVerified)
	}

	if verificationInfo.VerificationCodes.PhoneNumber == nil {
		logger.LogFormat("[VERIFY] A phone number verify request for account [%s] without pending phone number verification from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	} else if !securitycodes.ValidateSecurityCode(*verificationInfo.VerificationCodes.PhoneNumber, verificationCode) {
		logger.LogFormat("[VERIFY] A phone number verify request for account [%s] with incorrect verification code from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)
		return resp.SetClientError(apierrors.ErrorInvalidVerificationCode)
	}

	err = accountDB.SetAccountFlags(accountID, accounts.IsAccountVerifiedFlag|accounts.IsAccountPhoneNumberVerifiedFlag)
	if err != nil {
		return resp.SetServerError(err)
	}

	logger.LogFormat("[VERIFY] A successful phone number verify request for account [%s] from IP [%s] UserAgent [%s]\n", accountID, clientIP, clientUserAgent)

	userPhoneNumber := verificationInfo.PhoneNumber
	userLanguage := verificationInfo.Language
	if len(userLanguage) == 0 {
		userLanguage = defaultLanguageCode
	}

	// TODO: Do we want to send a welcome SMS?

	// Remove the verification code
	err = accountDB.RemoveAccountVerification(accountID, accountdatabase.VerificationTypePhoneNumber)
	if err != nil {
		return resp.SetServerError(err)
	}

	response := verifyPhoneNumberResponseBody{
		PhoneNumber: userPhoneNumber,
	}
	resp.SetBody(&response)
	return nil
}
