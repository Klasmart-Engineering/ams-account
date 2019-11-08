package handlers

import (
	"context"
	"net/http"
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

	// Gets the If-Modified-Since header value, if there is one
	ifNoETagMatch, _ := req.GetHeaderIfNoneMatch()

	// Gets the If-Modified-Since header value, if there is one
	var ifModifiedSinceTime *time.Time
	ifModifiedSinceTimeValue, hasIfModifiedSince, err := req.GetHeaderIfModifiedSince()
	if err != nil {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("If-Modified-Since"))
	} else if hasIfModifiedSince {
		ifModifiedSinceTime = &ifModifiedSinceTimeValue
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

	downloadURLResult, err := globals.AvatarStorage.GetAvatarFileDownloadURL(accountID, &cloudstorage.GetFileDownloadURLUsingCacheInput{
		IfNoETagMatch:   ifNoETagMatch,
		IfModifiedSince: ifModifiedSinceTime,
		DownloadInput: &cloudstorage.GetFileDownloadURLInput{
			Expires: urlExpireTime.Time(),
		},
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	// Skip the redirection if the client can use the cached version
	if downloadURLResult.UseCachedVersion {
		resp.SetStatus(http.StatusNotModified)
		return nil
	} else if downloadURLResult == nil {
		return resp.SetClientError(apierrors.ErrorItemNotFound)
	}

	resp.Redirect(downloadURLResult.DownloadOutput.URL)
	return nil
}
