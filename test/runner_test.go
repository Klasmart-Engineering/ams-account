package test

import (
	"testing"

	"bitbucket.org/calmisland/account-lambda-funcs/src/server"
	"bitbucket.org/calmisland/go-server-account/testaccountdatabase"
	"bitbucket.org/calmisland/go-server-account/testaccountdatabase/testaccountdynamodb"
	"bitbucket.org/calmisland/go-server-configs/configs"
)

// TestAccount execute a defined list of tests, using the project configuration environment
func TestAccount(t *testing.T) {
	err := configs.UpdateConfigDirectoryPath("../" + configs.DefaultConfigFolderName)
	if err != nil {
		panic(err)
	}
	server.Setup()
	testaccountdatabase.RunTestSuite(t, new(testaccountdynamodb.AccountDynamoDBClient))
}
