package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"github.com/labstack/echo/v4"
)

type otherAccountInfoResponseBody struct {
	FullName  string `json:"fullName,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email"`
}

// HandleGetOtherAccountInfo handles retrieving the other account information requests.
func HandleGetOtherAccountInfo(c echo.Context) error {
	accountID := c.Param("accountId")
	if len(accountID) == 0 {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters)
	}

	// Then get the account information
	accInfo, err := globals.AccountDatabase.GetAccountInfo(accountID)
	if err != nil {
		return err
	} else if accInfo == nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorItemNotFound)
	}

	response := otherAccountInfoResponseBody{
		FullName:  accInfo.FullName,
		FirstName: accInfo.FirstName,
		LastName:  accInfo.LastName,
		Email:     accInfo.Email,
	}

	return c.JSON(http.StatusOK, response)
}
