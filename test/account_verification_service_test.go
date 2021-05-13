package test_test

import (
	"os"
	"testing"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/globals"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/services/accountverificationservice"
	"bitbucket.org/calmisland/go-server-configs/configs"
)

func TestVerificationLinkByVerificationToken(t *testing.T) {
	os.Setenv("HOST_PASS_FRONTAPP", "https://beta-pass.badanamu.net")

	var accountVerificationConfig accountverificationservice.Config
	err := configs.ReadEnvConfig(&accountVerificationConfig)
	if err != nil {
		panic(err)
	}

	globals.AccountVerificationService, err = accountverificationservice.New(accountVerificationConfig)
	if err != nil {
		panic(err)
	}

	link := globals.AccountVerificationService.GetVerificationLinkByToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN0ZXZlLnNvbmdAY2FsbWlkLmNvbSIsImV4cGlyZUF0IjoxNjAzMTU3NjMxLCJwaG9uZU5yIjoiIiwicHciOiIkMmEkMTAkcDIvWjJjNC90ZG91SGNJeW5pVUJNZXQ5VXcydkZpaUM2L3J1WTVGM2lvYUZobXU4NEF6aVMiLCJ2ZXJpZmljYXRpb25Db2RlIjoiRDJBWTJWWSJ9.DbjnIf2iv2ghCoS98ZuqlTxmSoYAr0-YNCrlxdaukWQ",
		"D2AY2VY", "en")

	if link != "https://beta-pass.badanamu.net/#/verify_email_with_token?verificationToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN0ZXZlLnNvbmdAY2FsbWlkLmNvbSIsImV4cGlyZUF0IjoxNjAzMTU3NjMxLCJwaG9uZU5yIjoiIiwicHciOiIkMmEkMTAkcDIvWjJjNC90ZG91SGNJeW5pVUJNZXQ5VXcydkZpaUM2L3J1WTVGM2lvYUZobXU4NEF6aVMiLCJ2ZXJpZmljYXRpb25Db2RlIjoiRDJBWTJWWSJ9.DbjnIf2iv2ghCoS98ZuqlTxmSoYAr0-YNCrlxdaukWQ&code=D2AY2VY&lang=en" {
		t.Error("Verification link is wrong")
	}

}
