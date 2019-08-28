package main

import (
	"bitbucket.org/calmisland/go-server-auth/authmiddlewares"
	"bitbucket.org/calmisland/go-server-requests/apirouter"
	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
)

var (
	rootRouter *apirouter.Router
)

func initLambdaFunctions() {
	rootRouter = apirouter.NewRouter()
	routerV1 := createLambdaRouterV1()
	rootRouter.AddRouter("v1", routerV1)
}

func createLambdaRouterV1() *apirouter.Router {
	router := apirouter.NewRouter()
	router.AddMiddleware(authmiddlewares.ValidateSession(globals.AccessTokenValidator, true))

	return router
}
