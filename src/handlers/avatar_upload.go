package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"github.com/calmisland/go-errors"
	"github.com/google/uuid"
)

type avatarUploadRequestBody struct {
	ContentType   string `json:"contentType"`
	ContentSHA256 string `json:"contentSha256"`
	ContentLength int64  `json:"contentLength"`
}

type avatarUploadResponseBody struct {
	AttachmentID  string            `json:"attachmentId"`
	UploadURL     string            `json:"uploadUrl"`
	UploadMethod  string            `json:"uploadMethod"`
	UploadHeaders map[string]string `json:"uploadHeaders,omitempty"`
}

const (
	contentLengthHeaderName = "Content-Length"
	imageContentType        = "image/jpeg"

	attachmentMaxSize = 1 * 1024 * 1024 // 1 MB
	sha256ByteLength  = sha256.Size
	sha256HexLength   = sha256ByteLength * 2
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
	contentSHA256Str := reqBody.ContentSHA256
	contentSHA256, err := hex.DecodeString(contentSHA256Str)
	if err != nil || len(contentSHA256) != sha256ByteLength {
		return resp.SetClientError(apierrors.ErrorInvalidParameters.WithField("contentSha256"))
	}

	attachmentUUID, err := uuid.NewRandom()
	if err != nil {
		return resp.SetServerError(errors.Wrap(err, "Failed to generate an UUID"))
	}

	// Get the signed upload URL
	attachmentID := attachmentUUID.String()

	response := avatarUploadResponseBody{
		AttachmentID:  attachmentID,
		UploadURL:     "https://not.yet.badanamu.net/upload",
		UploadMethod:  "POST",
		UploadHeaders: nil,
	}
	resp.SetBody(&response)
	return nil
}
