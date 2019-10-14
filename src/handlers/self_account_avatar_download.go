package handlers

import (
	"context"

	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

// HandleSelfAccountAvatarDownload handles self account avatar download requests.
func HandleSelfAccountAvatarDownload(ctx context.Context, req *apirequests.Request, resp *apirequests.Response) error {

	// TODO: Redirect to a signed avatar URL through CDN
	resp.Redirect("http://www.calmid.com/wp-content/uploads/2019/03/logo.jpg")
	return nil
}
