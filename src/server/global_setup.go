// +build !china

package server

import (
	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/src/services/accountverificationservice"
	"bitbucket.org/calmisland/go-server-account/accountdatabase/accountdynamodb"
	"bitbucket.org/calmisland/go-server-account/avatar"
	"bitbucket.org/calmisland/go-server-aws/awsdynamodb"
	"bitbucket.org/calmisland/go-server-aws/awss3"
	"bitbucket.org/calmisland/go-server-aws/awssqs"
	"bitbucket.org/calmisland/go-server-configs/configs"
	"bitbucket.org/calmisland/go-server-geoip/geoip"
	"bitbucket.org/calmisland/go-server-geoip/services/maxmind"
	"bitbucket.org/calmisland/go-server-logs/errorreporter"
	"bitbucket.org/calmisland/go-server-logs/errorreporter/slackreporter"
	"bitbucket.org/calmisland/go-server-messages/sendmessagequeue"
	"bitbucket.org/calmisland/go-server-requests/tokens/accesstokens"
	"bitbucket.org/calmisland/go-server-security/passwords"
)

// Setup Setup
func Setup() {
	if err := awsdynamodb.InitializeFromConfigs(); err != nil {
		panic(err)
	}

	accountdynamodb.ActivateDatabase()

	setupAccessTokenSystems()
	setupPasswordPolicyValidator()
	setupPasswordHasher()
	setupMessageQueue()
	setupGeoIP()
	setupAvatarStorage()
	setupAccountVerificationService()
	setupSlackReporter()

	globals.Verify()
}

func setupAccessTokenSystems() {
	var validatorConfig accesstokens.ValidatorConfig
	err := configs.LoadConfig("access_tokens", &validatorConfig, true)
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
	err := configs.LoadConfig("password_policy", &passwordPolicyConfig, true)
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
	err := configs.LoadConfig("password_hash", &passwordHashConfig, true)
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
	err := configs.LoadConfig("message_send_sqs", &queueConfig, true)
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

func setupGeoIP() {
	if err := maxmind.ActivateGeoIPService(); err != nil {
		panic(err)
	}

	globals.GeoIPService = geoip.GetDefaultService()
}

func setupAvatarStorage() {
	var avatarStorageConfig avatar.StorageConfig
	err := configs.LoadConfig("avatar_storage", &avatarStorageConfig, true)
	if err != nil {
		panic(err)
	}

	var s3StorageConfig awss3.StorageConfig
	err = configs.LoadConfig("avatar_storage_s3", &s3StorageConfig, true)
	if err != nil {
		panic(err)
	}

	avatarStorageConfig.Storage, err = awss3.NewStorage(&s3StorageConfig)
	if err != nil {
		panic(err)
	}

	globals.AvatarStorage, err = avatar.NewStorage(avatarStorageConfig)
	if err != nil {
		panic(err)
	}
}

func setupAccountVerificationService() {
	var accountVerificationConfig accountverificationservice.Config
	err := configs.LoadConfig("account_verification", &accountVerificationConfig, true)
	if err != nil {
		panic(err)
	}

	globals.AccountVerificationService, err = accountverificationservice.New(accountVerificationConfig)
	if err != nil {
		panic(err)
	}
}

func setupSlackReporter() {
	var slackReporterConfig slackreporter.Config
	err := configs.LoadConfig("error_reporter_slack", &slackReporterConfig, false)
	if err != nil {
		panic(err)
	}

	// Check if there is a configuration for the Slack error reporter
	if len(slackReporterConfig.HookURL) == 0 {
		return
	}

	reporter, err := slackreporter.New(&slackReporterConfig)
	if err != nil {
		panic(err)
	}

	errorreporter.Active = reporter
}
