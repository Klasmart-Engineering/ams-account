package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/echoadapter"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-utils/langutils"
	"bitbucket.org/calmisland/go-server-utils/textutils"
	"github.com/labstack/echo/v4"
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
func HandleEditSelfAccountInfo(c echo.Context) error {
	// Parse the request body
	reqBody := new(editSelfAccountInfoRequestBody)
	err := c.Bind(reqBody)

	if err != nil {
		return echoadapter.SetClientError(c, apierrors.ErrorBadRequestBody)
	}

	language := reqBody.Language
	if language != nil {
		languageValue := textutils.SanitizeString(*language)
		if !langutils.IsValidLanguageCode(languageValue) {
			return echoadapter.SetClientError(c, apierrors.ErrorInvalidParameters.WithField("lang"))
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
			return echoadapter.SetClientError(c, apierrors.ErrorInvalidParameters)
		} else if len(fullName) > maxFullNameLength {
			return echoadapter.SetClientError(c, apierrors.ErrorInputTooLong.WithField("names.fullName").WithValue(maxFullNameLength))
		} else if len(firstName) > maxPartNameLength {
			return echoadapter.SetClientError(c, apierrors.ErrorInputTooLong.WithField("names.firstName").WithValue(maxPartNameLength))
		} else if len(lastName) > maxPartNameLength {
			return echoadapter.SetClientError(c, apierrors.ErrorInputTooLong.WithField("names.lastName").WithValue(maxPartNameLength))
		}

		editNameInfo = &accountdatabase.AccountNameInfo{
			FullName:  &fullName,
			FirstName: &firstName,
			LastName:  &lastName,
		}
	}

	cc := c.(*echoadapter.AuthContext)
	accountID := cc.Session.Data.AccountID
	err = globals.AccountDatabase.EditAccount(accountID, &accountdatabase.AccountEditInfo{
		Language: language,
		Names:    editNameInfo,
	})

	if err != nil {
		return err
	}

	logger.LogFormat("[EDITACCOUNTINFO] A successful edit account request for account [%s]\n", accountID)
	return cc.NoContent(http.StatusOK)
}
