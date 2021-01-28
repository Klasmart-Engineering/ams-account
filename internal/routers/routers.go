package routers

import (
	"net/http"

	apiControllerV1 "bitbucket.org/calmisland/account-lambda-funcs/internal/controllers/v1"
	apiControllerV2 "bitbucket.org/calmisland/account-lambda-funcs/internal/controllers/v2"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/echoadapter"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRouter is ...
func SetupRouter() *echo.Echo {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", hello)

	v1 := e.Group("/v1")

	v1.GET("/serverinfo", apiControllerV1.HandleServerInfo)

	v1.POST("/forgotpassword", apiControllerV1.HandleForgotPassword)
	v1.POST("/restorepassword", apiControllerV1.HandleRestorePassword)
	v1.POST("/signup", apiControllerV1.HandleSignUp)

	v1resend := v1.Group("/resend/verification")
	v1resend.POST("/email", apiControllerV1.HandleResendEmailVerification)
	v1resend.POST("/phonenumber", apiControllerV1.HandleResendPhoneNumberVerification)

	v1verify := v1.Group("/verify")
	v1verify.GET("/email", apiControllerV1.HandleAccountEmailVerified)
	v1verify.POST("/email", apiControllerV1.HandleVerifyEmail)
	v1verify.GET("/phonenumber", apiControllerV1.HandleAccountPhoneVerified)
	v1verify.POST("/phonenumber", apiControllerV1.HandleVerifyPhoneNumber)

	authMiddleware := echoadapter.AuthMiddleware(globals.AccessTokenValidator, true)

	v1self := v1.Group("/self")
	v1self.Use(authMiddleware)
	v1self.GET("/info", apiControllerV1.HandleGetSelfAccountInfo)
	v1self.POST("/info", apiControllerV1.HandleEditSelfAccountInfo)
	v1self.POST("/password", apiControllerV1.HandleEditSelfAccountPassword)
	v1self.GET("/avatar", apiControllerV1.HandleSelfAccountAvatarDownload)
	v1self.PUT("/avatar", apiControllerV1.HandleSelfAvatarUpload)
	v1self.DELETE("/avatar", apiControllerV1.HandleSelfAccountAvatarDelete)

	v1other := v1.Group("/other")
	v1other.Use(authMiddleware)
	v1other.GET("/:accountId/info", apiControllerV1.HandleGetOtherAccountInfo)
	v1other.GET("/:accountId/avatar", apiControllerV1.HandleOtherAccountAvatarDownload)

	v2 := e.Group("/v2")

	v2.POST("/signup/request", apiControllerV2.HandleSignupRequest)
	v2.POST("/signup/confirm", apiControllerV2.HandleSignUpConfirm)

	v2.POST("/verify/email", apiControllerV2.HandleVerifyEmail)
	v2.POST("/kl15/migrate", apiControllerV2.HandleKl15Migration)

	return e
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
