package handlers

import (
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/handlers/handlers_v2"
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

	routerV1 := CreateLambdaRouterV1()
	rootRouter.AddRouter("v1", routerV1)
	routerV2 := CreateRouterV2()
	rootRouter.AddRouter("v2", routerV2)
	return rootRouter
}

func CreateLambdaRouterV1() *apirouter.Router {
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

func CreateRouterV2() *apirouter.Router {
	router := apirouter.NewRouter()
	signupRouter := apirouter.NewRouter()
	signupRouter.AddMethodHandler("POST", "request", handlers_v2.HandleSignupRequest)
	signupRouter.AddMethodHandler("POST", "confirm", handlers_v2.HandleSignUpConfirm)
	router.AddRouter("signup", signupRouter)

	verifyRouter := apirouter.NewRouter()
	verifyRouter.AddMethodHandler("POST", "email", handlers_v2.HandleVerifyEmail)
	router.AddRouter("verify", verifyRouter)

	kl15MigrationRouter := apirouter.NewRouter()
	kl15MigrationRouter.AddMethodHandler("POST", "migrate", handlers_v2.HandleKl15Migration)
	router.AddRouter("kl15", kl15MigrationRouter)

	return router
}
