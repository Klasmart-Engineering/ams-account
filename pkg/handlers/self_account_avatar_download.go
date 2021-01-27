package handlers

import (
	"context"
	"net/http"
	"time"

	"bitbucket.org/calmisland/account-lambda-funcs/pkg/globals"
	"bitbucket.org/calmisland/go-server-cloud/cloudstorage"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-utils/timeutils"
)

// HandleSelfAccountAvatarDownload handles self account avatar download requests.
func HandleSelfAccountAvatarDownload(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	accountID := req.Session.Data.AccountID

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
	} else if downloadURLResult == nil {
		return resp.SetClientError(apierrors.ErrorItemNotFound)
	}

	// Skip the redirection if the client can use the cached version
	if downloadURLResult.UseCachedVersion {
		resp.SetStatus(http.StatusNotModified)
		return nil
	}

	resp.Redirect(downloadURLResult.DownloadOutput.URL)
	return nil
}
