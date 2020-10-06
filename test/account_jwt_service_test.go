package test_test

import (
	"testing"
	"time"

	"bitbucket.org/calmisland/account-lambda-funcs/src/services/account_jwt_service"
)

func TestJwtToken(t *testing.T) {
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
	if claims.VerificationCode != verificationCode {
		t.Error("VerificationCode attribute is incorrect")
	}

	if err != nil {
		t.Error(err)
	}

}
