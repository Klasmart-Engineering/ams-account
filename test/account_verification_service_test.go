package test_test

import (
	"os"
	"testing"

	"bitbucket.org/calmisland/account-lambda-funcs/src/services/accountverificationservice"
)

func TestVerificationLinkByVerificationToken(t *testing.T) {
	os.Setenv("HOST_PASS_FRONTAPP", "https://beta-pass.badanamu.net")
	link := accountverificationservice.GetVerificationLinkByToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN0ZXZlLnNvbmdAY2FsbWlkLmNvbSIsImV4cGlyZUF0IjoxNjAzMTU3NjMxLCJwaG9uZU5yIjoiIiwicHciOiIkMmEkMTAkcDIvWjJjNC90ZG91SGNJeW5pVUJNZXQ5VXcydkZpaUM2L3J1WTVGM2lvYUZobXU4NEF6aVMiLCJ2ZXJpZmljYXRpb25Db2RlIjoiRDJBWTJWWSJ9.DbjnIf2iv2ghCoS98ZuqlTxmSoYAr0-YNCrlxdaukWQ",
		"D2AY2VY", "en")

	if link != "https://beta-pass.badanamu.net/#/verify_email_with_token?verificationToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InN0ZXZlLnNvbmdAY2FsbWlkLmNvbSIsImV4cGlyZUF0IjoxNjAzMTU3NjMxLCJwaG9uZU5yIjoiIiwicHciOiIkMmEkMTAkcDIvWjJjNC90ZG91SGNJeW5pVUJNZXQ5VXcydkZpaUM2L3J1WTVGM2lvYUZobXU4NEF6aVMiLCJ2ZXJpZmljYXRpb25Db2RlIjoiRDJBWTJWWSJ9.DbjnIf2iv2ghCoS98ZuqlTxmSoYAr0-YNCrlxdaukWQ&code=D2AY2VY&lang=en" {
		t.Error("Verification link is wrong")
	}

}
