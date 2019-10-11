package handlers

import (
	"context"

	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

// HandleOtherAccountAvatarDownload handles other account avatar download requests.
func HandleOtherAccountAvatarDownload(ctx context.Context, req *apirequests.Request, resp *apirequests.Response) error {

	accountID, _ := req.GetPathParam("accountId")
	if len(accountID) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	// TODO: Redirect to a signed avatar URL through CDN
	resp.Redirect("http://www.calmid.com/wp-content/uploads/2019/03/logo.jpg")
	return nil
}
