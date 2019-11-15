package globals

import (
	"bitbucket.org/calmisland/account-lambda-funcs/src/services/accountverificationservice"
	"bitbucket.org/calmisland/go-server-account/avatar"
	"bitbucket.org/calmisland/go-server-geoip/geoip"
	"bitbucket.org/calmisland/go-server-messages/sendmessagequeue"
	"bitbucket.org/calmisland/go-server-requests/tokens/accesstokens"
	"bitbucket.org/calmisland/go-server-security/passwords"
	"github.com/calmisland/go-errors"
)

var (
	// AccessTokenValidator is the access token validator.
	AccessTokenValidator accesstokens.Validator

	// MessageSendQueue is the message send queue.
	MessageSendQueue sendmessagequeue.Queue

	// PasswordPolicyValidator is a password policy validator.
	PasswordPolicyValidator passwords.PasswordPolicyValidator
	// PasswordHasher is the password hasher.
	PasswordHasher passwords.PasswordHasher

	// GeoIPService is the Geo IP service.
	GeoIPService geoip.Service

	// AvatarStorage is store handle avatar image.
	AvatarStorage avatar.Storage

	// AccountVerificationService is the account verification service.
	AccountVerificationService accountverificationservice.Service
)

// Verify verifies if all variables have been properly set.
func Verify() {
	if AccessTokenValidator == nil {
		panic(errors.New("The access token validator has not been set"))
	}

	if MessageSendQueue == nil {
		panic(errors.New("The message send queue has not been set"))
	}

	if PasswordPolicyValidator == nil {
		panic(errors.New("The password policy validator has not been set"))
	} else if PasswordHasher == nil {
		panic(errors.New("The password hasher has not been set"))
	}

	if GeoIPService == nil {
		panic(errors.New("The Geo IP service has not been set"))
	}

	if AvatarStorage == nil {
		panic(errors.New("The avatar storage has not been set"))
	}

	if AccountVerificationService == nil {
		panic(errors.New("The account verification service has not been set"))
	}
}
