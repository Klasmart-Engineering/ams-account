package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"github.com/labstack/echo/v4"
)

type accountEmailVerifiedResponseBody struct {
	Verified bool `json:"verified"`
}

// HandleAccountEmailVerified handles account email verified requests.
func HandleAccountEmailVerified(c echo.Context) error {
	accountID := c.QueryParam("accountId")
	if len(accountID) == 0 {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters)
	}

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	} else if verificationInfo == nil {
		// Instead of returning an error, return false so it avoids accountId guessing

		return c.JSON(http.StatusOK, &accountEmailVerifiedResponseBody{
			Verified: false,
		})
	} else {
		return c.JSON(http.StatusOK, &accountEmailVerifiedResponseBody{
			Verified: accounts.IsAccountEmailVerified(verificationInfo.Flags),
		})
	}
}
