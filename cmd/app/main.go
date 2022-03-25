//go:build !lambda
// +build !lambda

package main

import (
	"bitbucket.org/calmisland/account-lambda-funcs/internal/routers"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/setup/globalsetup"
	"bitbucket.org/calmisland/go-server-configs/configs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func customHTTPErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	c.Echo().DefaultHTTPErrorHandler(err, c)
}

func main() {
	err := configs.UpdateConfigDirectoryPath(configs.DefaultConfigFolderName)
	if err != nil {
		panic(err)
	}

	globalsetup.Setup()

	echo := routers.SetupRouter()
	echo.Logger.SetLevel(log.DEBUG)
	echo.HTTPErrorHandler = customHTTPErrorHandler
	// Start server
	echo.Logger.Fatal(echo.Start(":8089"))
}
