package utils

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func EchoHandleHTTPError(errCode int, err error) *echo.HTTPError {
	if errCode == http.StatusInternalServerError {
		sentry.CaptureException(err)
	}
	return echo.NewHTTPError(errCode, err.Error())
}
