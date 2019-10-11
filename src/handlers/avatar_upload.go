package handlers

import (
	"context"
	"encoding/hex"

	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
)

type avatarUploadRequestBody struct {
	ContentType   string `json:"contentType"`
	ContentSHA256 string `json:"contentSha256"`
	ContentLength int64  `json:"contentLength"`
}

type avatarUploadResponseBody struct {
	UploadURL    string `json:"uploadUrl"`
	UploadMethod string `json:"uploadMethod"`
}

const (
	contentLengthHeaderName = "Content-Length"

	imageContentType = "image/jpeg"

	attachmentMaxSize = 10 * 1024 * 1024 // 10 MB

	sha256ByteLength = 256 / 8
	sha256HexLength  = sha256ByteLength * 2
)

// HandleAvatarUpload handles avtar upload requests.
func HandleAvatarUpload(ctx context.Context, req *apirequests.Request, resp *apirequests.Response) error {
	var reqBody avatarUploadRequestBody
	err := req.UnmarshalBody(&reqBody)
	if err != nil {
		return resp.SetServerError(err)
	}

	if len(reqBody.ContentType) == 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentType"))
	} else if len(reqBody.ContentSHA256) != sha256HexLength {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentSha256"))
	} else if reqBody.ContentLength <= 0 {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentLength"))
	} else if reqBody.ContentLength > attachmentMaxSize {
		return resp.SetClientError(apierrors.ErrorInputTooLong.WithField("contentLength").WithValue(attachmentMaxSize))
	}

	// Validate the content type
	if reqBody.ContentType != imageContentType {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentType"))
	}

	// Validate the content hash
	contentHash, err := hex.DecodeString(reqBody.ContentSHA256)
	if err != nil || len(contentHash) != sha256ByteLength {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentSha256"))
	}

	response := avatarUploadResponseBody{
		UploadURL:    "https://not.yet.badanamu.net/upload",
		UploadMethod: "POST",
	}
	resp.SetBody(&response)
	return nil
}
