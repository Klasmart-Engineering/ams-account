package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/go-server-auth/authmiddlewares"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"github.com/labstack/echo/v4"
)

type selfAccountInfoResponseBody struct {
	FullName  string `json:"fullName,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email"`
	Country   string `json:"country"`
	Language  string `json:"lang"`
}

// HandleGetSelfAccountInfo handles retrieving the signed in account information requests.
func HandleGetSelfAccountInfo(c echo.Context) error {
	// Then get the account information
	cc := c.(*authmiddlewares.AuthContext)

	accountID := cc.Session.Data.AccountID
	accInfo, err := globals.AccountDatabase.GetAccountInfo(accountID)
	if err != nil {
		return err
	} else if accInfo == nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorItemNotFound)
	}

	response := selfAccountInfoResponseBody{
		FullName:  accInfo.FullName,
		FirstName: accInfo.FirstName,
		LastName:  accInfo.LastName,
		Email:     accInfo.Email,
		Country:   accInfo.Country,
		Language:  accInfo.Language,
	}

	return cc.JSON(http.StatusOK, response)
}
