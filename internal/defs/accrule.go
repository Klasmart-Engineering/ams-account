package defs

import (
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"github.com/labstack/echo/v4"
)

const (
	DefaultCountryCode  = "XX"
	DefaultLanguageCode = "en_US"

	SignUpVerificationCodeByteLength = 4
)

func HandlePasswordValidatorError(c echo.Context, err error) error {
	switch err.(type) {
	case *passwords.PasswordTooShortError:
		passwordErr := err.(*passwords.PasswordTooShortError)
		return apirequests.EchoSetClientError(c, apierrors.ErrorPasswordTooShort.WithValue(int64(passwordErr.MinimumLength)))
	case *passwords.PasswordTooLongError:
		passwordErr := err.(*passwords.PasswordTooLongError)
		return apirequests.EchoSetClientError(c, apierrors.ErrorPasswordTooLong.WithValue(int64(passwordErr.MaximumLength)))
	case *passwords.PasswordLowerCaseMissingError:
		passwordErr := err.(*passwords.PasswordLowerCaseMissingError)
		return apirequests.EchoSetClientError(c, apierrors.ErrorPasswordLowerCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordUpperCaseMissingError:
		passwordErr := err.(*passwords.PasswordUpperCaseMissingError)
		return apirequests.EchoSetClientError(c, apierrors.ErrorPasswordUpperCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordNumberMissingError:
		passwordErr := err.(*passwords.PasswordNumberMissingError)
		return apirequests.EchoSetClientError(c, apierrors.ErrorPasswordNumberMissing.WithValue(int64(passwordErr.MinimumCount)))
	default:
		return err
	}
}
