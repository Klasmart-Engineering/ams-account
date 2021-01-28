package accountverificationservice

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/calmisland/go-errors"
)

// Service is the account verification service.
type Service interface {
	// GetVerificationLink returns a verification link.
	GetVerificationLink(accountID, verificationCode, language string) string
}

// Config is the configuration for the account verification service.
type Config struct {
	// URL is the URL for the account verification.
	URL string `json:"url" env:"ACCOUNT_VERIFICATION_URL"`
}

type standardService struct {
	url              string
	hasLanguageParam bool
}

const (
	accountIDParamName         = "{accountId}"
	verificationCodeParamName  = "{code}"
	languageParamName          = "{language}"
	verificationTokenParamName = "{verificationToken}"
)

// New creates a new account verification service.
func New(config Config) (Service, error) {
	if len(config.URL) == 0 {
		return nil, errors.New("The URL cannot be empty")
	}

	url := config.URL
	if !strings.Contains(url, accountIDParamName) {
		return nil, errors.Errorf("The URL is missing the Account ID parameter: %s", accountIDParamName)
	} else if !strings.Contains(url, verificationCodeParamName) {
		return nil, errors.Errorf("The URL is missing the Verification Code parameter: %s", verificationCodeParamName)
	}

	hasLanguageParam := strings.Contains(url, languageParamName)
	return &standardService{
		url:              url,
		hasLanguageParam: hasLanguageParam,
	}, nil
}

// GetVerificationLink returns a verification link.
func (service *standardService) GetVerificationLink(accountID, verificationCode, language string) string {
	accountID = url.QueryEscape(accountID)
	verificationCode = url.QueryEscape(verificationCode)
	language = url.QueryEscape(language)

	linkURL := service.url
	linkURL = strings.ReplaceAll(linkURL, accountIDParamName, accountID)
	linkURL = strings.ReplaceAll(linkURL, verificationCodeParamName, verificationCode)

	if service.hasLanguageParam {
		linkURL = strings.ReplaceAll(linkURL, languageParamName, language)
	}

	return linkURL
}

func GetVerificationLinkByToken(verificationToken string, verificationCode string, language string) string {

	host := os.Getenv("HOST_PASS_FRONTAPP")
	path := fmt.Sprintf(`/#/verify_email_with_token?verificationToken=%s&code=%s&lang=%s`, verificationTokenParamName, verificationCodeParamName, languageParamName)

	verificationToken = url.QueryEscape(verificationToken)
	verificationCode = url.QueryEscape(verificationCode)
	language = url.QueryEscape(language)

	linkURL := host + path
	linkURL = strings.ReplaceAll(linkURL, verificationTokenParamName, verificationToken)
	linkURL = strings.ReplaceAll(linkURL, verificationCodeParamName, verificationCode)
	linkURL = strings.ReplaceAll(linkURL, languageParamName, language)

	return linkURL
}
