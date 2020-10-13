package test_test

import (
	"os"
	"testing"
	"time"

	"bitbucket.org/calmisland/account-lambda-funcs/src/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/src/services/account_jwt_service"
	"bitbucket.org/calmisland/go-server-security/passwords"
)

func TestJwtToken(t *testing.T) {
	passwordHashConfig := passwords.PasswordHashConfig{
		DefaultCost: 10,
		SecureCost:  13,
	}
	t.Log(passwordHashConfig)
	var errNewPwHasher error
	globals.PasswordHasher, errNewPwHasher = passwords.NewPasswordHasher(passwordHashConfig)
	if errNewPwHasher != nil {
		t.Error(errNewPwHasher)
	}
	t.Log(globals.PasswordHasher)

	email := "steve.song@calmid.com"
	phoneNumber := "+821012341234"
	password := "asdf"
	verificationCode := "1234"
	language := "ko"
	expireAt := time.Now().Unix()

	token, err := account_jwt_service.CreateToken(&account_jwt_service.TokenMapClaims{
		Email:            email,
		PhoneNumber:      phoneNumber,
		Password:         password,
		VerificationCode: verificationCode,
		Language:         language,
		ExpireAt:         expireAt,
	})
	if err != nil {
		t.Error(err)
	}
	claims, err := account_jwt_service.VerifyToken(token)

	if claims.Email != email {
		t.Error("Email attribute is incorrect")
	}
	t.Log(verificationCode)
	t.Log(claims.VerificationCode)

	// // Verify that the current password is correct
	if !globals.PasswordHasher.VerifyPasswordHash(verificationCode, claims.VerificationCode) { // Verifies the password
		t.Error(`Encrypted code does not match`)
	}

}

func TestEncryptHashedCode(t *testing.T) {
	code := "AD1XK7"
	encrypted := account_jwt_service.EncryptHashedCode(code)
	t.Log(encrypted)

	// Verify that the current password is correct
	if !globals.PasswordHasher.VerifyPasswordHash(code, encrypted) { // Verifies the password
		t.Error(`Encrypted code does not match`)
	}

}

func TestGetSecret(t *testing.T) {
	jwtSecretEnv := os.Getenv("JWT_SECRET")
	secret := account_jwt_service.GetSecret()
	t.Log(string(secret))
	if jwtSecretEnv == "" && string(secret) == "C@almIsl@nd" {
		t.Log(`default jwt secret`)
	} else if string(secret) != jwtSecretEnv {
		t.Error(`could not get jwt secret`)
	}
}
