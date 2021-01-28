package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/echoadapter"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"github.com/labstack/echo/v4"
)

// HandleSelfAccountAvatarDelete handles avatar image delete requests.
func HandleSelfAccountAvatarDelete(c echo.Context) error {
	cc := c.(*echoadapter.AuthContext)
	accountID := cc.Session.Data.AccountID
	err := globals.AvatarStorage.DeleteAvatarFile(accountID)
	if err != nil {
		return err
	}

	return cc.NoContent(http.StatusOK)
}
