package handlers_v2

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

type verifyEmailReqBody struct {
	Email string `json:"email"`
}

type verifyEmailRespBody struct {
	Result bool `json:"result"`
}

// HandleSignUp handles sign-up requests.
func HandleVerifyEmail(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Parse the request body
	var reqBody verifyEmailReqBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	email := reqBody.Email

	_, ok, err := globals.AccountDatabase.GetAccountIDFromEmail(email)
	if err != nil {
		return resp.SetServerError(err)
	}

	response := verifyEmailRespBody{
		Result: ok,
	}

	resp.SetBody(&response)
	return nil
}
