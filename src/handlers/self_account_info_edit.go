package handlers

import (
	"context"
	"log"

	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
)

type editSelfAccountInfoRequestBody struct {
	Language *string       `json:"lang"`
	Names    *editNameInfo `json:"names"`
}

type editNameInfo struct {
	FullName  string `json:"fullName"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

const (
	maxFullNameLength = 64
	maxPartNameLength = 32
)

// HandleEditSelfAccountInfo handles requests for editing account information for the signed in account.
func HandleEditSelfAccountInfo(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody editSelfAccountInfoRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	language := reqBody.Language
	if language != nil {
		languageValue := textutils.SanitizeString(*language)
		if !langutils.IsValidLanguageCode(languageValue) {
			return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("lang"))
		}

		language = &languageValue
	}

	var editNameInfo *accountdatabase.AccountNameInfo
	names := reqBody.Names
	if names != nil {
		fullName := textutils.SanitizeString(names.FullName)
		firstName := textutils.SanitizeString(names.FirstName)
		lastName := textutils.SanitizeString(names.LastName)
		if len(fullName) == 0 && len(firstName) == 0 && len(lastName) == 0 { // At least one name must be specified
			return resp.SetClientError(apierrors.ErrorInvalidParameters)
		} else if len(fullName) > maxFullNameLength {
			return resp.SetClientError(apierrors.ErrorInputTooLong.WithField("names.fullName").WithValue(maxFullNameLength))
		} else if len(firstName) > maxPartNameLength {
			return resp.SetClientError(apierrors.ErrorInputTooLong.WithField("names.firstName").WithValue(maxPartNameLength))
		} else if len(lastName) > maxPartNameLength {
			return resp.SetClientError(apierrors.ErrorInputTooLong.WithField("names.lastName").WithValue(maxPartNameLength))
		}

		editNameInfo = &accountdatabase.AccountNameInfo{
			FullName:  &fullName,
			FirstName: &firstName,
			LastName:  &lastName,
		}
	}

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	accountID := req.Session.Data.AccountID
	err = accountDB.EditAccount(accountID, &accountdatabase.AccountEditInfo{
		Language: language,
		Names:    editNameInfo,
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	log.Printf("[EDITACCOUNTINFO] A successful edit account request for account [%s]\n", accountID)
	return nil
}
