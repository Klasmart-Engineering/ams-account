package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/go-server-auth/authmiddlewares"
	"github.com/labstack/echo/v4"
)

// HandleSelfAccountAvatarDelete handles avatar image delete requests.
func HandleSelfAccountAvatarDelete(c echo.Context) error {
	cc := c.(*authmiddlewares.AuthContext)
	accountID := cc.Session.Data.AccountID
	err := globals.AvatarStorage.DeleteAvatarFile(accountID)
	if err != nil {
		return err
	}

	return cc.NoContent(http.StatusOK)
}
