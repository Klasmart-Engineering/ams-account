package echoadapter

import (
	"net/http"
	"strings"
	"time"

	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"github.com/calmisland/go-errors"

	"github.com/labstack/echo/v4"
)

func SetClientError(c echo.Context, apiErr *apierrors.APIError) error {
	if apiErr == nil {
		return errors.New("The client API error cannot be nil")
	}

	// Log the client error
	logger.LogFormat("Client Error: %s", apiErr.Error())

	c.JSON(apiErr.StatusCode, apiErr)

	return nil
}

func HandlePasswordValidatorError(c echo.Context, err error) error {
	switch err.(type) {
	case *passwords.PasswordTooShortError:
		passwordErr := err.(*passwords.PasswordTooShortError)
		return SetClientError(c, apierrors.ErrorPasswordTooShort.WithValue(int64(passwordErr.MinimumLength)))
	case *passwords.PasswordTooLongError:
		passwordErr := err.(*passwords.PasswordTooLongError)
		return SetClientError(c, apierrors.ErrorPasswordTooLong.WithValue(int64(passwordErr.MaximumLength)))
	case *passwords.PasswordLowerCaseMissingError:
		passwordErr := err.(*passwords.PasswordLowerCaseMissingError)
		return SetClientError(c, apierrors.ErrorPasswordLowerCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordUpperCaseMissingError:
		passwordErr := err.(*passwords.PasswordUpperCaseMissingError)
		return SetClientError(c, apierrors.ErrorPasswordUpperCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordNumberMissingError:
		passwordErr := err.(*passwords.PasswordNumberMissingError)
		return SetClientError(c, apierrors.ErrorPasswordNumberMissing.WithValue(int64(passwordErr.MinimumCount)))
	default:
		return err
	}
}

const (
	headerIfModifiedSince   = "If-Modified-Since"
	headerIfUnmodifiedSince = "If-Unmodified-Since"
	headerIfNoneMatch       = "If-None-Match"
	headerIfMatch           = "If-Match"

	headerLastModified = "Last-Modified"
	headerETag         = "ETag "
)

func GetHeaderIfNoneMatch(c echo.Context) ([]string, bool) {
	return getHeaderEtags(c, headerIfNoneMatch)
}

func GetHeaderIfModifiedSince(c echo.Context) (time.Time, bool, error) {
	return GetHeaderDate(c, headerIfModifiedSince)
}

func GetHeaderDate(c echo.Context, headerName string) (time.Time, bool, error) {
	timeValueText := c.Request().Header.Get(headerName)
	if len(timeValueText) == 0 {
		return time.Time{}, false, nil
	}

	timeValue, err := http.ParseTime(timeValueText)
	return timeValue, true, errors.Wrapf(err, "Failed to parse the date timestamp: %s", timeValueText)
}

func getHeaderEtags(c echo.Context, headerName string) ([]string, bool) {
	etagsText := c.Request().Header.Get(headerName)
	if len(etagsText) == 0 {
		return nil, false
	}

	etags := strings.Split(etagsText, ",")
	for i, etag := range etags {
		etags[i] = strings.TrimSpace(etag)
	}
	return etags, true
}
