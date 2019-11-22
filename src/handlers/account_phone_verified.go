package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accounts"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

type accountPhoneVerifiedResponseBody struct {
	Verified bool `json:"verified"`
}

// HandleAccountPhoneVerified handles other account phone verified requests.
func HandleAccountPhoneVerified(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	accountID, _ := req.GetQueryParam("accountId")
	if len(accountID) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	verificationInfo, err := globals.AccountDatabase.GetAccountVerifications(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if verificationInfo == nil {
		// Instead of returning an error, return false so it avoids accountId guessing
		resp.SetBody(&accountPhoneVerifiedResponseBody{
			Verified: false,
		})
		return nil
	} else {
		resp.SetBody(&accountPhoneVerifiedResponseBody{
			Verified: accounts.IsAccountPhoneNumberVerified(verificationInfo.Flags),
		})
		return nil
	}
}
