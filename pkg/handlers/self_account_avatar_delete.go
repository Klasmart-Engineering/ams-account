package handlers

import (
	"context"

	"bitbucket.org/calmisland/account-lambda-funcs/pkg/globals"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

// HandleSelfAccountAvatarDelete handles avatar image delete requests.
func HandleSelfAccountAvatarDelete(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	accountID := req.Session.Data.AccountID
	err := globals.AvatarStorage.DeleteAvatarFile(accountID)
	if err != nil {
		return resp.SetServerError(err)
	}

	return nil
}
