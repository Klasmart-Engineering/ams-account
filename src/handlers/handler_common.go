package handlers

import (
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/apirequests"
	"bitbucket.org/calmisland/go-server-security/passwords"
)

const (
	defaultCountryCode  = "XX"
	defaultLanguageCode = "en_US"

	signUpVerificationCodeByteLength = 4
)

func handlePasswordValidatorError(resp *apirequests.Response, err error) error {
	switch err.(type) {
	case *passwords.PasswordTooShortError:
		passwordErr := err.(*passwords.PasswordTooShortError)
		return resp.SetClientError(apierrors.ErrorPasswordTooShort.WithValue(int64(passwordErr.MinimumLength)))
	case *passwords.PasswordTooLongError:
		passwordErr := err.(*passwords.PasswordTooLongError)
		return resp.SetClientError(apierrors.ErrorPasswordTooLong.WithValue(int64(passwordErr.MaximumLength)))
	case *passwords.PasswordLowerCaseMissingError:
		passwordErr := err.(*passwords.PasswordLowerCaseMissingError)
		return resp.SetClientError(apierrors.ErrorPasswordLowerCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordUpperCaseMissingError:
		passwordErr := err.(*passwords.PasswordUpperCaseMissingError)
		return resp.SetClientError(apierrors.ErrorPasswordUpperCaseMissing.WithValue(int64(passwordErr.MinimumCount)))
	case *passwords.PasswordNumberMissingError:
		passwordErr := err.(*passwords.PasswordNumberMissingError)
		return resp.SetClientError(apierrors.ErrorPasswordNumberMissing.WithValue(int64(passwordErr.MinimumCount)))
	default:
		return resp.SetServerError(err)
	}
}
