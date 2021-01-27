package testsetup

import (
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/pkg/services/accountverificationservice/accountverificationservicemock"
	"bitbucket.org/calmisland/go-server-account/accountdatabase/accountmemorydb"
	"bitbucket.org/calmisland/go-server-account/avatars"
	"bitbucket.org/calmisland/go-server-cloud/cloudstorage/memorystorage"
	"bitbucket.org/calmisland/go-server-geoip/geoip"
	"bitbucket.org/calmisland/go-server-geoip/geoip/geoipmock"
	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-messages/sendmessagequeue/sendmessagequeuemock"
	"bitbucket.org/calmisland/go-server-requests/sessions"
	"bitbucket.org/calmisland/go-server-requests/tokens/accesstokens/accesstokensmock"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"github.com/calmisland/go-testify/mock"
)

// Setup Setup
func Setup() {
	// Disable the logging
	logger.SetLogger(nil)

	setupAccountDatabase()
	setupAccessTokenSystems()
	setupPasswordPolicyValidator()
	setupPasswordHasher()
	setupEmailQueue()
	setupGeoIP()
	setupAvatarStorage()
	setupAccountVerificationService()

	globals.Verify()
}

func setupAccountDatabase() {
	globals.AccountDatabase = accountmemorydb.New()
}

func setupAccessTokenSystems() {
	accessTokenValidator := &accesstokensmock.MockValidator{}
	accessTokenValidator.On("ValidateAccessToken", mock.Anything).Return(&sessions.SessionData{
		SessionID: "TEST-SESSION",
		AccountID: "TEST-ACCOUNT",
		DeviceID:  "TEST-DEVICE",
	}, nil)

	globals.AccessTokenValidator = accessTokenValidator
}

func setupPasswordPolicyValidator() {
	var err error
	globals.PasswordPolicyValidator, err = passwords.NewPasswordPolicyValidator(passwords.PasswordPolicyConfig{})
	if err != nil {
		panic(err)
	}
}

func setupPasswordHasher() {
	var err error
	globals.PasswordHasher, err = passwords.NewPasswordHasher(passwords.PasswordHashConfig{
		DefaultCost: 5,
		SecureCost:  5,
	})
	if err != nil {
		panic(err)
	}
}

func setupEmailQueue() {
	messageSendQueue := &sendmessagequeuemock.QueueMock{}
	messageSendQueue.On("EnqueueMessage", mock.Anything).Return(nil)
	globals.MessageSendQueue = messageSendQueue
}

func setupGeoIP() {
	geoIPService := &geoipmock.ServiceMock{}
	geoIPService.On("GetCountryFromIP", mock.Anything).Return(&geoip.CountryResult{
		CountryCode:   "XX",
		ContinentCode: "XX",
	})

	globals.GeoIPService = geoIPService
}

func setupAvatarStorage() {
	avatarMemoryStorage := memorystorage.NewStorage()
	avatarStorageConfig := avatars.StorageConfig{
		Storage:    avatarMemoryStorage,
		AvatarPath: "avatars/",
	}

	var err error
	globals.AvatarStorage, err = avatars.NewStorage(avatarStorageConfig)
	if err != nil {
		panic(err)
	}
}

func setupAccountVerificationService() {
	verificationService := &accountverificationservicemock.MockService{}
	verificationService.On("GetVerificationLink", mock.Anything, mock.Anything, mock.Anything).Return("http://localhost:9999/verify")
	globals.AccountVerificationService = verificationService
}
