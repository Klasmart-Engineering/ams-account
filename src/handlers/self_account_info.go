package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

type selfAccountInfoResponseBody struct {
	FullName  string `json:"fullName,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email"`
	Country   string `json:"country"`
	Language  string `json:"lang"`
}

// HandleGetSelfAccountInfo handles retrieving the signed in account information requests.
func HandleGetSelfAccountInfo(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Then get the account information
	accountID := req.Session.Data.AccountID
	accInfo, err := globals.AccountDatabase.GetAccountInfo(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accInfo == nil {
		return resp.SetClientError(apierrors.ErrorItemNotFound)
	}

	response := selfAccountInfoResponseBody{
		FullName:  accInfo.FullName,
		FirstName: accInfo.FirstName,
		LastName:  accInfo.LastName,
		Email:     accInfo.Email,
		Country:   accInfo.Country,
		Language:  accInfo.Language,
	}
	resp.SetBody(&response)
	return nil
}
