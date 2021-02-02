package v1

import (
	"net/http"
	"time"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"bitbucket.org/calmisland/go-server-cloud/cloudstorage"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-utils/timeutils"
	"github.com/labstack/echo/v4"
)

// HandleSelfAccountAvatarDownload handles self account avatar download requests.
func HandleSelfAccountAvatarDownload(c echo.Context) error {
	accountID := helpers.GetAccountID(c)
	// Gets the If-Modified-Since header value, if there is one
	ifNoETagMatch, _ := apirequests.EchoGetHeaderIfNoneMatch(c)

	// Gets the If-Modified-Since header value, if there is one
	var ifModifiedSinceTime *time.Time
	ifModifiedSinceTimeValue, hasIfModifiedSince, err := apirequests.EchoGetHeaderIfModifiedSince(c)
	if err != nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorInvalidParameters.WithField("If-Modified-Since"))
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
		return helpers.HandleInternalError(c, err)
	} else if downloadURLResult == nil {
		return apirequests.EchoSetClientError(c, apierrors.ErrorItemNotFound)
	}

	// Skip the redirection if the client can use the cached version
	if downloadURLResult.UseCachedVersion {
		return c.NoContent(http.StatusNotModified)
	}

	return c.Redirect(http.StatusTemporaryRedirect, downloadURLResult.DownloadOutput.URL)
}
