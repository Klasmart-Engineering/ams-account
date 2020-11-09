package defs

import (
	"bitbucket.org/calmisland/account-lambda-funcs/src/utils"
	"bitbucket.org/calmisland/go-server-info/serverinfo"
)

const (
	SERVER_STAGE_BETA      = serverinfo.BetaStageName
	SERVER_STAGE_PROD      = serverinfo.ProductionStageName
	TEST_VERIFICATION_CODE = "TESTONLY"
)

var (
	SERVER_STAGE = utils.GetOsEnvWithDef("SERVER_STAGE", SERVER_STAGE_PROD)
)

func EnsureTestVerificationCode(code string) bool {
	if SERVER_STAGE == SERVER_STAGE_BETA && code == TEST_VERIFICATION_CODE {
		return true
	}
	return false
}
