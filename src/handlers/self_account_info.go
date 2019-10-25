package handlers

import (
	"context"

	"bitbucket.org/calmisland/go-server-account/accountdatabase"
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
	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	accountID := req.Session.Data.AccountID

	// Then get the account information
	accInfo, err := accountDB.GetAccountInfo(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accInfo == nil {
		return resp.SetClientError(apierrors.ErrorInvalidLogin)
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
