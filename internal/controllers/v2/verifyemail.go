package v2

import (
	"net/http"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"github.com/labstack/echo/v4"
)

type verifyEmailReqBody struct {
	Email string `json:"email"`
}

type verifyEmailRespBody struct {
	Result bool `json:"result"`
}

// HandleSignUp handles sign-up requests.
func HandleVerifyEmail(c echo.Context) error {
	// Parse the request body
	reqBody := new(verifyEmailReqBody)
	err := c.Bind(reqBody)

	if err != nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorBadRequestBody)
	}

	email := reqBody.Email

	_, ok, err := globals.AccountDatabase.GetAccountIDFromEmail(email)
	if err != nil {
		return helpers.HandleInternalError(c, err)
	}

	response := verifyEmailRespBody{
		Result: ok,
	}

	return c.JSON(http.StatusOK, response)
}
