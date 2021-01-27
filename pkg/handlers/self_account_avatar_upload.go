package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"bitbucket.org/calmisland/account-lambda-funcs/pkg/globals"
	"bitbucket.org/calmisland/go-server-cloud/cloudstorage"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-utils/timeutils"
)

type avatarUploadRequestBody struct {
	ContentType   string `json:"contentType"`
	ContentSHA256 string `json:"contentSha256"`
	ContentLength int64  `json:"contentLength"`
}

type avatarUploadResponseBody struct {
	UploadURL     string            `json:"uploadUrl"`
	UploadMethod  string            `json:"uploadMethod"`
	UploadHeaders map[string]string `json:"uploadHeaders,omitempty"`
}

const (
	contentLengthHeaderName = "Content-Length"
	imageContentType        = "image/jpeg"

	avatarMaxSize                 = 1 * 1024 * 1024 // 1 MB
	avatarExpireDuration          = 24 * time.Hour
	avatarUploadURLExpireDuration = 30 * time.Minute

	sha256ByteLength = sha256.Size
	sha256HexLength  = sha256ByteLength * 2
)

// HandleSelfAvatarUpload handles self avtar upload requests.
func HandleSelfAvatarUpload(_ context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	var reqBody avatarUploadRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetClientError(apierrors.ErrorBadRequestBody)
	}

	if len(reqBody.ContentType) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentType"))
	} else if len(reqBody.ContentSHA256) != sha256HexLength {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentSha256"))
	} else if reqBody.ContentLength <= 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentLength"))
	} else if reqBody.ContentLength > avatarMaxSize {
		return resp.SetClientError(apierrors.ErrorInputTooLong.WithField("contentLength").WithValue(avatarMaxSize))
	}

	// Validate the content type
	if reqBody.ContentType != imageContentType {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentType"))
	}

	accountID := req.Session.Data.AccountID

	// Validate the content hash
	contentSHA256Str := reqBody.ContentSHA256
	contentSHA256, err := hex.DecodeString(contentSHA256Str)
	if err != nil || len(contentSHA256) != sha256ByteLength {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentSha256"))
	}

	// Get the avatar expiration time
	avatarExpireTimeMs := timeutils.EpochMSNow().Add(avatarExpireDuration)
	avatarExpireTime := avatarExpireTimeMs.Time()

	// Get the upload URL expiration time
	urlExpireTime := timeutils.EpochMSNow().Add(avatarUploadURLExpireDuration)

	uploadURLResult, err := globals.AvatarStorage.GetAvatarFileUploadURL(accountID, &cloudstorage.GetFileUploadURLInput{
		ContentLength:  &reqBody.ContentLength,
		ContentType:    &reqBody.ContentType,
		ContentExpires: &avatarExpireTime,
		ContentSHA256:  contentSHA256,
		Expires:        urlExpireTime.Time(),
	})
	if err != nil {
		return resp.SetServerError(err)
	}

	response := avatarUploadResponseBody{
		UploadURL:     uploadURLResult.URL,
		UploadMethod:  uploadURLResult.Method,
		UploadHeaders: uploadURLResult.Headers,
	}
	resp.SetBody(&response)
	return nil
}
