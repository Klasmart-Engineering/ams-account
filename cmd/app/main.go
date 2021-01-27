// +build !lambda

package main

import (
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/handlers"
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/setup/globalsetup"
	"bitbucket.org/calmisland/go-server-configs/configs"
	"bitbucket.org/calmisland/go-server-requests/apiservers/httpserver"
)

func main() {
	err := configs.UpdateConfigDirectoryPath(configs.DefaultConfigFolderName)
	if err != nil {
		panic(err)
	}

	globalsetup.Setup()
	rootRouter := handlers.InitializeRoutes()

	restServer := &httpserver.Server{
		ListenAddress: ":8089",
		Handler:       rootRouter,
	}

	err = restServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
