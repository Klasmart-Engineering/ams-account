package handlers

import (
	"context"
	"time"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/go-server-account/accountdatabase"
	"bitbucket.org/calmisland/go-server-cloud/cloudstorage"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-utils/timeutils"
)

const (
	avatarDownloadURLExpireDuration = 30 * time.Minute
)

// HandleOtherAccountAvatarDownload handles other account avatar download requests.
func HandleOtherAccountAvatarDownload(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	accountID, _ := req.GetPathParam("accountId")
	if len(accountID) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters)
	}

	// Get the database
	accountDB, err := accountdatabase.GetDatabase()
	if err != nil {
		return resp.SetServerError(err)
	}

	// Then get the account information
	accInfo, err := accountDB.GetAccountInfo(accountID)
	if err != nil {
		return resp.SetServerError(err)
	} else if accInfo == nil {
		return resp.SetClientError(apierrors.ErrorItemNotFound)
	}

	// Get the download URL expiration time
	urlExpireTime := timeutils.EpochMSNow().Add(avatarDownloadURLExpireDuration)

	downloadURLResult, err := globals.AvatarStorage.GetAvatarFileDownloadURL(accountID, &cloudstorage.GetFileDownloadURLInput{
		Expires: urlExpireTime.Time(),
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	resp.Redirect(downloadURLResult.URL)
	return nil
}
