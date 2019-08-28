// +build !china

package server

import (
	"bitbucket.org/calmisland/go-server-aws/awsdynamodb"
	"bitbucket.org/calmisland/go-server-aws/awssqs"
	"bitbucket.org/calmisland/go-server-account/accountdatabase/accountdynamodb"
	"bitbucket.org/calmisland/go-server-configs/configs"
	"bitbucket.org/calmisland/go-server-logs/errorreporter"
	"bitbucket.org/calmisland/go-server-logs/errorreporter/slackreporter"
	"bitbucket.org/calmisland/go-server-requests/tokens/accesstokens"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"bitbucket.org/calmisland/go-server-emails/emailqueue"
	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
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
	setupEmailQueue()
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

func setupEmailQueue() {
	var queueConfig awssqs.QueueConfig
	err := configs.LoadConfig("email_send_sqs", &queueConfig, true)
	if err != nil {
		panic(err)
	}

	emailSendQueue, err := awssqs.NewQueue(queueConfig)
	if err != nil {
		panic(err)
	}

	globals.EmailSendQueue, err = emailqueue.NewEmailSendQueue(emailqueue.EmailSendQueueConfig{
		Queue: emailSendQueue,
	})
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
