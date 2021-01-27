package accountverificationservicemock

import (
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/services/accountverificationservice"
	"github.com/calmisland/go-testify/mock"
)

// MockService is the mocked account verification service.
type MockService struct {
	accountverificationservice.Service
	mock.Mock
}

// GetVerificationLink returns a verification link.
func (service *MockService) GetVerificationLink(accountID, verificationCode, language string) string {
	args := service.Called(accountID, verificationCode, language)
	return args.String(0)
}
