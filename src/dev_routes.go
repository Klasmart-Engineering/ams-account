// +build !lambda

package main

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-requests/apirouter"
)

func initLambdaDevFunctions(rootRouter *apirouter.Router) {
	devRouter := apirouter.NewRouter()
	devRouter.AddMethodHandler("GET", "createtables", createTablesRequest)
	rootRouter.AddRouter("dev", devRouter)
}

func createTablesRequest(_ context.Context, _ *apirequests.Request, resp *apirequests.Response) error {
	err := globals.AccountDatabase.CreateDatabaseTables()
	if err != nil {
		return resp.SetServerError(err)
	}

	resp.SetBodyDirect("text/plain", []byte("OK"), false)
	return nil
}
