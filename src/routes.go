package main

import (
	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/src/handlers"
	"bitbucket.org/calmisland/go-server-auth/authmiddlewares"
	"bitbucket.org/calmisland/go-server-requests/apirouter"
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

	accountRouter := apirouter.NewRouter()
	router.AddRouter("account", accountRouter)

	selfAccountRouter := apirouter.NewRouter()
	selfAccountRouter.AddMethodHandler("GET", "info", handlers.HandleGetSelfAccountInfo)
	selfAccountRouter.AddMethodHandler("POST", "info", handlers.HandleEditSelfAccountInfo)
	accountRouter.AddRouter("self", selfAccountRouter)

	return router
}
