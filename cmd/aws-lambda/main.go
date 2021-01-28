// +build lambda

package main

import (
	"bitbucket.org/calmisland/account-lambda-funcs/internal/handlers"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/setup/globalsetup"
	"bitbucket.org/calmisland/go-server-aws/awslambda"
	"bitbucket.org/calmisland/go-server-configs/configs"
)

func main() {
	err := configs.UpdateConfigDirectoryPath("./" + configs.DefaultConfigFolderName)
	if err != nil {
		panic(err)
	}

	globalsetup.Setup()
	rootRouter := handlers.InitializeRoutes()

	err = awslambda.StartAPIHandler(rootRouter)
	if err != nil {
		panic(err)
	}
}
