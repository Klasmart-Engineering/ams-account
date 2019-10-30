// +build !lambda

package main

import (
	"context"

	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-requests/apirouter"
)

func initLambdaDevFunctions(rootRouter *apirouter.Router) {
	devRouter := apirouter.NewRouter()
	devRouter.AddMethodHandler("GET", "createtables", createTablesRequest)
	rootRouter.AddRouter("dev", devRouter)
}

func createTablesRequest(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	if req.HTTPMethod != "GET" {
		return resp.SetClientError(apierrors.ErrorBadRequestMethod)
	}

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	err = accountDB.CreateDatabaseTables()
	if err != nil {
		return resp.SetServerError(err)
	}

	resp.SetBodyDirect("text/plain", []byte("OK"), false)
	return nil
}
