package globals

import (
	"github.com/calmisland/go-errors"
	"bitbucket.org/calmisland/go-server-requests/tokens/accesstokens"
	"bitbucket.org/calmisland/go-server-emails/emailqueue"
	"bitbucket.org/calmisland/go-server-security/passwords"
)

var (
	// AccessTokenValidator is the access token validator.
	AccessTokenValidator accesstokens.Validator

	// EmailSendQueue is the email send queue.
	EmailSendQueue emailqueue.EmailSendQueue

	// PasswordPolicyValidator is a password policy validator.
	PasswordPolicyValidator passwords.PasswordPolicyValidator
	// PasswordHasher is the password hasher.
	PasswordHasher passwords.PasswordHasher
)

// Verify verifies if all variables have been properly set.
func Verify() {
	if AccessTokenValidator == nil {
		panic(errors.New("The access token validator has not been set"))
	}

	if EmailSendQueue == nil {
		panic(errors.New("The email send queue has not been set"))
	}

	if PasswordPolicyValidator == nil {
		panic(errors.New("The password policy validator has not been set"))
	} else if PasswordHasher == nil {
		panic(errors.New("The password hasher has not been set"))
	}
}
