package handlers

import (
	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-auth/authmiddlewares"
	"bitbucket.org/calmisland/go-server-requests/apirouter"
	"bitbucket.org/calmisland/go-server-requests/standardhandlers"
)

var (
	rootRouter *apirouter.Router
)

// InitializeRoutes initializes the routes.
func InitializeRoutes() *apirouter.Router {
	if rootRouter != nil {
		return rootRouter
	}

	rootRouter = apirouter.NewRouter()
	if globals.CORSOptions != nil {
		rootRouter.AddCORSMiddleware(globals.CORSOptions)
	}

	routerV1 := createLambdaRouterV1()
	rootRouter.AddRouter("v1", routerV1)
	return rootRouter
}

func createLambdaRouterV1() *apirouter.Router {
	requireAuthMiddleware := authmiddlewares.ValidateSession(globals.AccessTokenValidator, true)
	router := apirouter.NewRouter()
	router.AddMethodHandler("GET", "serverinfo", standardhandlers.HandleServerInfo)
	router.AddMethodHandler("POST", "forgotpassword", HandleForgotPassword)
	router.AddMethodHandler("POST", "restorepassword", HandleRestorePassword)
	router.AddMethodHandler("POST", "signup", HandleSignUp)

	accountResendRouter := apirouter.NewRouter()
	router.AddRouter("resend", accountResendRouter)

	accountResendVerifyRouter := apirouter.NewRouter()
	accountResendVerifyRouter.AddMethodHandler("POST", "email", HandleResendEmailVerification)
	accountResendVerifyRouter.AddMethodHandler("POST", "phonenumber", HandleResendPhoneNumberVerification)
	accountResendRouter.AddRouter("verification", accountResendVerifyRouter)

	accountVerifyRouter := apirouter.NewRouter()
	accountVerifyRouter.AddMethodHandler("GET", "email", HandleAccountEmailVerified)
	accountVerifyRouter.AddMethodHandler("POST", "email", HandleVerifyEmail)
	accountVerifyRouter.AddMethodHandler("GET", "phonenumber", HandleAccountPhoneVerified)
	accountVerifyRouter.AddMethodHandler("POST", "phonenumber", HandleVerifyPhoneNumber)
	router.AddRouter("verify", accountVerifyRouter)

	selfAccountRouter := apirouter.NewRouter()
	selfAccountRouter.AddMiddleware(requireAuthMiddleware)
	selfAccountRouter.AddMethodHandler("GET", "info", HandleGetSelfAccountInfo)
	selfAccountRouter.AddMethodHandler("POST", "info", HandleEditSelfAccountInfo)
	selfAccountRouter.AddMethodHandler("POST", "password", HandleEditSelfAccountPassword)
	selfAccountRouter.AddMethodHandler("GET", "avatar", HandleSelfAccountAvatarDownload)
	selfAccountRouter.AddMethodHandler("PUT", "avatar", HandleSelfAvatarUpload)
	selfAccountRouter.AddMethodHandler("DELETE", "avatar", HandleSelfAccountAvatarDelete)
	router.AddRouter("self", selfAccountRouter)

	otherAccountRouter := apirouter.NewRouter()
	otherAccountRouter.AddMiddleware(requireAuthMiddleware)
	router.AddRouter("other", otherAccountRouter)

	specificOtherAccountRouter := apirouter.NewRouter()
	specificOtherAccountRouter.AddMethodHandler("GET", "info", HandleGetOtherAccountInfo)
	specificOtherAccountRouter.AddMethodHandler("GET", "avatar", HandleOtherAccountAvatarDownload)
	otherAccountRouter.AddRouterWildcard("accountId", specificOtherAccountRouter)

	return router
}