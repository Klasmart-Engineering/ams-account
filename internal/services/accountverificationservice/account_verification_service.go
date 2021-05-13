package accountverificationservice

import (
	"fmt"
	"net/url"

	"github.com/calmisland/go-errors"
)

// Service is the account verification service.
type Service interface {
	// GetVerificationLink returns a verification link.
	GetVerificationLink(accountID, verificationCode, language string) string
	GetVerificationLinkByToken(verificationToken, verificationCode, language string) string
}

// Config is the configuration for the account verification service.
type Config struct {
	// URL is the URL for the account verification.
	PassFrontendHost string `json:"url" env:"HOST_PASS_FRONTAPP"`
}

type standardService struct {
	passFrontendHost string
}

// New creates a new account verification service.
func New(config Config) (Service, error) {
	if len(config.PassFrontendHost) == 0 {
		return nil, errors.New("The URL cannot be empty")
	}

	return &standardService{
		passFrontendHost: config.PassFrontendHost,
	}, nil
}

// GetVerificationLink returns a verification link.
func (service *standardService) GetVerificationLink(accountID, verificationCode, language string) string {
	accountID = url.QueryEscape(accountID)
	verificationCode = url.QueryEscape(verificationCode)
	language = url.QueryEscape(language)

	return fmt.Sprintf("%s/#/verify_email?accountId=%s&code=%s&lang=%s", service.passFrontendHost, accountID, verificationCode, language)
}

func (service *standardService) GetVerificationLinkByToken(verificationToken string, verificationCode string, language string) string {
	verificationToken = url.QueryEscape(verificationToken)
	verificationCode = url.QueryEscape(verificationCode)
	language = url.QueryEscape(language)

	return fmt.Sprintf(`%s/#/verify_email_with_token?verificationToken=%s&code=%s&lang=%s`, service.passFrontendHost, verificationToken, verificationCode, language)
}
