package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/pkg/globals"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

type otherAccountInfoResponseBody struct {
	FullName  string `json:"fullName,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email"`
}

// HandleGetOtherAccountInfo handles retrieving the other account information requests.
func HandleGetOtherAccountInfo(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	accountID, _ := req.GetPathParam("accountId")
	if len(accountID) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	// Then get the account information
	accInfo, err := globals.AccountDatabase.GetAccountInfo(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accInfo == nil {
		return resp.SetClientError(apierrors.ErrorItemNotFound)
	}

	response := otherAccountInfoResponseBody{
		FullName:  accInfo.FullName,
		FirstName: accInfo.FirstName,
		LastName:  accInfo.LastName,
		Email:     accInfo.Email,
	}
	resp.SetBody(&response)
	return nil
}