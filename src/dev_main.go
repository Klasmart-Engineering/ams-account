// +build !lambda

package main

import (
	"bitbucket.org/calmisland/go-server-configs/configs"
	"bitbucket.org/calmisland/go-server-requests/apiservers/httpserver"
	"bitbucket.org/calmisland/account-lambda-funcs/src/server"
)

func main() {
	err := configs.UpdateConfigDirectoryPath(configs.DefaultConfigFolderName)
	if err != nil {
		panic(err)
	}
	server.Setup()

	initLambdaFunctions()
	initLambdaDevFunctions()

	restServer := &httpserver.Server{
		ListenAddress: ":8089",
		Handler:       rootRouter,
	}

	err = restServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
