package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"github.com/labstack/echo/v4"
)

// HandleSelfAccountAvatarDelete handles avatar image delete requests.
func HandleSelfAccountAvatarDelete(c echo.Context) error {
	accountID := helpers.GetAccountID(c)
	err := globals.AvatarStorage.DeleteAvatarFile(accountID)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	return c.NoContent(http.StatusOK)
}
