package v1

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"bitbucket.org/calmisland/go-server-auth/authmiddlewares"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
)

// HandleSelfAccountAvatarDelete handles avatar image delete requests.
func HandleSelfAccountAvatarDelete(c echo.Context) error {
	cc := c.(*authmiddlewares.AuthContext)
	accountID := cc.Session.Data.AccountID

	hub := sentryecho.GetHubFromContext(c)
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID: accountID,
		})
	})

	err := globals.AvatarStorage.DeleteAvatarFile(accountID)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	return cc.NoContent(http.StatusOK)
}
