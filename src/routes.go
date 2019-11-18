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
	requireAuthMiddleware := authmiddlewares.ValidateSession(globals.AccessTokenValidator, true)
	router := apirouter.NewRouter()
	router.AddMethodHandler("GET", "serverinfo", handlers.HandleServerInfo)

	accountRouter := apirouter.NewRouter()
	accountRouter.AddMethodHandler("POST", "forgotpassword", handlers.HandleForgotPassword)
	accountRouter.AddMethodHandler("POST", "restorepassword", handlers.HandleRestorePassword)
	accountRouter.AddMethodHandler("POST", "signup", handlers.HandleSignUp)
	router.AddRouter("account", accountRouter)

	accountResendRouter := apirouter.NewRouter()
	accountRouter.AddRouter("resend", accountResendRouter)

	accountResendVerifyRouter := apirouter.NewRouter()
	accountResendVerifyRouter.AddMethodHandler("POST", "email", handlers.HandleResendEmailVerification)
	accountResendVerifyRouter.AddMethodHandler("POST", "phonenumber", handlers.HandleResendPhoneNumberVerification)
	accountResendRouter.AddRouter("verification", accountResendVerifyRouter)

	accountVerifyRouter := apirouter.NewRouter()
	accountVerifyRouter.AddMethodHandler("GET", "email", handlers.HandleAccountEmailVerified)
	accountVerifyRouter.AddMethodHandler("POST", "email", handlers.HandleVerifyEmail)
	accountVerifyRouter.AddMethodHandler("GET", "phonenumber", handlers.HandleAccountPhoneVerified)
	accountVerifyRouter.AddMethodHandler("POST", "phonenumber", handlers.HandleVerifyPhoneNumber)
	accountRouter.AddRouter("verify", accountVerifyRouter)

	selfAccountRouter := apirouter.NewRouter()
	selfAccountRouter.AddMiddleware(requireAuthMiddleware)
	selfAccountRouter.AddMethodHandler("GET", "info", handlers.HandleGetSelfAccountInfo)
	selfAccountRouter.AddMethodHandler("POST", "info", handlers.HandleEditSelfAccountInfo)
	selfAccountRouter.AddMethodHandler("POST", "password", handlers.HandleEditSelfAccountPassword)
	selfAccountRouter.AddMethodHandler("GET", "avatar", handlers.HandleSelfAccountAvatarDownload)
	selfAccountRouter.AddMethodHandler("PUT", "avatar", handlers.HandleSelfAvatarUpload)
	selfAccountRouter.AddMethodHandler("DELETE", "avatar", handlers.HandleSelfAccountAvatarDelete)
	accountRouter.AddRouter("self", selfAccountRouter)

	otherAccountRouter := apirouter.NewRouter()
	otherAccountRouter.AddMiddleware(requireAuthMiddleware)
	accountRouter.AddRouter("other", otherAccountRouter)

	specificOtherAccountRouter := apirouter.NewRouter()
	specificOtherAccountRouter.AddMethodHandler("GET", "info", handlers.HandleGetOtherAccountInfo)
	specificOtherAccountRouter.AddMethodHandler("GET", "avatar", handlers.HandleOtherAccountAvatarDownload)
	otherAccountRouter.AddRouterWildcard("accountId", specificOtherAccountRouter)

	return router
}
