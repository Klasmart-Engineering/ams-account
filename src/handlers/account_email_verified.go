package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

type accountEmailVerifiedResponseBody struct {
	Verified bool `json:"verified"`
}

// HandleAccountEmailVerified handles account email verified requests.
func HandleAccountEmailVerified(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	accountID, _ := req.GetQueryParam("accountId")
	if len(accountID) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil {
		// Instead of returning an error, return false so it avoids accountId guessing
		resp.SetBody(&accountEmailVerifiedResponseBody{
			Verified: false,
		})
		return nil
	} else {
		resp.SetBody(&accountEmailVerifiedResponseBody{
			Verified: accounts.IsAccountEmailVerified(verificationInfo.Flags),
		})
		return nil
	}
}
