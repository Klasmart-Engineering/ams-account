// +build !lambda

package main

import (
	"bitbucket.org/calmisland/account-lambda-funcs/internal/routers"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/setup/globalsetup"
	"bitbucket.org/calmisland/go-server-configs/configs"
)

func main() {
	err := configs.UpdateConfigDirectoryPath(configs.DefaultConfigFolderName)
	if err != nil {
		panic(err)
	}

	globalsetup.Setup()

	echo := routers.SetupRouter()

	// Start server
	echo.Logger.Fatal(echo.Start(":8089"))
}
