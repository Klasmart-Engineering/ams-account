package globalsetup

import (
	"fmt"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/services/accountverificationservice"
	"bitbucket.org/calmisland/go-server-account/accountdatabase/accountdynamodb"
	"bitbucket.org/calmisland/go-server-account/avatars"
	"bitbucket.org/calmisland/go-server-aws/awsdynamodb"
	"bitbucket.org/calmisland/go-server-aws/awss3"
	"bitbucket.org/calmisland/go-server-aws/awssqs"
	"bitbucket.org/calmisland/go-server-configs/configs"
	"bitbucket.org/calmisland/go-server-geoip/geoip"
	"bitbucket.org/calmisland/go-server-geoip/services/maxmind"
	"bitbucket.org/calmisland/go-server-messages/sendmessagequeue"
	"bitbucket.org/calmisland/go-server-requests/tokens/accesstokens"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"github.com/getsentry/sentry-go"
)

// Setup Setup
func Setup() {
	setupSentry()

	setupAccountDatabase()
	setupAccessTokenSystems()
	setupPasswordPolicyValidator()
	setupPasswordHasher()
	setupMessageQueue()
	setupGeoIP()
	setupAvatarStorage()
	setupAccountVerificationService()

	globals.Verify()
}

func setupSentry() {
	if err := sentry.Init(sentry.ClientOptions{}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
}

func setupAccountDatabase() {
	var accountDatabaseConfig awsdynamodb.ClientConfig
	err := configs.ReadEnvConfig(&accountDatabaseConfig)
	if err != nil {
		panic(err)
	}

	ddbClient, err := awsdynamodb.NewClient(&accountDatabaseConfig)
	if err != nil {
		panic(err)
	}

	globals.AccountDatabase, err = accountdynamodb.New(ddbClient)
	if err != nil {
		panic(err)
	}
}

func setupAccessTokenSystems() {
	var validatorConfig accesstokens.ValidatorConfig
	err := configs.ReadEnvConfig(&validatorConfig)
	if err != nil {
		panic(err)
	}

	globals.AccessTokenValidator, err = accesstokens.NewValidator(validatorConfig)
	if err != nil {
		panic(err)
	}
}

func setupPasswordPolicyValidator() {
	var passwordPolicyConfig passwords.PasswordPolicyConfig
	err := configs.ReadEnvConfig(&passwordPolicyConfig)
	if err != nil {
		panic(err)
	}

	globals.PasswordPolicyValidator, err = passwords.NewPasswordPolicyValidator(passwordPolicyConfig)
	if err != nil {
		panic(err)
	}
}

func setupPasswordHasher() {
	var passwordHashConfig passwords.PasswordHashConfig
	err := configs.ReadEnvConfig(&passwordHashConfig)
	if err != nil {
		panic(err)
	}

	globals.PasswordHasher, err = passwords.NewPasswordHasher(passwordHashConfig)
	if err != nil {
		panic(err)
	}
}

func setupMessageQueue() {
	var queueConfig awssqs.QueueConfig
	err := configs.ReadEnvConfig(&queueConfig)
	if err != nil {
		panic(err)
	}

	messageSendQueue, err := awssqs.NewQueue(queueConfig)
	if err != nil {
		panic(err)
	}

	globals.MessageSendQueue, err = sendmessagequeue.New(sendmessagequeue.QueueConfig{
		Queue: messageSendQueue,
	})
	if err != nil {
		panic(err)
	}
}

func ActivateGeoIPService() error {
	var config maxmind.GeoIPConfig
	err := configs.ReadEnvConfig(&config)
	if err != nil {
		return err
	}

	service, err := maxmind.NewGeoIPService(config)
	if err != nil {
		return err
	}

	geoip.SetDefaultService(service)
	return nil
}

func setupGeoIP() {
	if err := ActivateGeoIPService(); err != nil {
		panic(err)
	}

	globals.GeoIPService = geoip.GetDefaultService()
}

func setupAvatarStorage() {
	var avatarStorageConfig avatars.StorageConfig
	err := configs.ReadEnvConfig(&avatarStorageConfig)
	if err != nil {
		panic(err)
	}

	var s3StorageConfig awss3.StorageConfig
	err = configs.ReadEnvConfig(&s3StorageConfig)
	if err != nil {
		panic(err)
	}

	avatarStorageConfig.Storage, err = awss3.NewStorage(&s3StorageConfig)
	if err != nil {
		panic(err)
	}

	globals.AvatarStorage, err = avatars.NewStorage(avatarStorageConfig)
	if err != nil {
		panic(err)
	}
}

func setupAccountVerificationService() {
	var accountVerificationConfig accountverificationservice.Config
	err := configs.ReadEnvConfig(&accountVerificationConfig)
	if err != nil {
		panic(err)
	}

	globals.AccountVerificationService, err = accountverificationservice.New(accountVerificationConfig)
	if err != nil {
		panic(err)
	}
}
