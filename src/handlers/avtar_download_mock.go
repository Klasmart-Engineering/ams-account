package handlers

import (
	"context"

	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

// HandleAvtarDownloadMock handles avtar download requests.
func HandleAvtarDownloadMock(ctx context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	// Get the database
	db, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	accountID := req.Session.Data.AccountID

	// Then get the account information
	accInfo, err := db.GetAccountInfo(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accInfo == nil {
		return resp.SetClientError(apierrors.ErrorInvalidLogin)
	}

	// TODO: Redirect to a signed avtar URL through CDN
	resp.Redirect("http://www.calmid.com/wp-content/uploads/2019/03/logo.jpg")
	return nil
}
