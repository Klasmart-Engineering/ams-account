package v2

import (
	"net/http"
	"os"

	"bitbucket.org/calmisland/account-lambda-funcs/internal/helpers"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/models"
	"bitbucket.org/calmisland/account-lambda-funcs/internal/utils"
	"github.com/labstack/echo/v4"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

// HandleDeletionAccount handles account deletion requests.
func HandleDeletionAccount(c echo.Context) error {
	accountID := helpers.GetAccountID(c)

	sess := session.Must(session.NewSession())
	db := dynamo.New(sess, &aws.Config{Region: aws.String(os.Getenv("DYNAMODB_REGION"))})

	tableAccount := db.Table(models.GetTableName(models.TABLE_NAME_ACCOUNT))

	var resultAccount models.Account
	err := tableAccount.Get("id", accountID).One(&resultAccount)

	if err != nil {
		if err.Error() == "dynamo: no item found" {
			return utils.EchoHandleHTTPError(http.StatusNotFound, err)
		} else {
			return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	}

	var resultAccountEmail models.AccountEmail
	tableAccountEmail := db.Table(models.GetTableName(models.TABLE_NAME_ACCOUNT_EMAIL))
	err = tableAccountEmail.Get("accId", accountID).Index(models.ACCOUNT_GSI_ACCID).One(&resultAccountEmail)

	if err != nil {
		if err.Error() == "dynamo: no item found" {
			return utils.EchoHandleHTTPError(http.StatusNotFound, err)
		} else {
			return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	} else {
		err = tableAccountEmail.Delete("email", resultAccountEmail.Email).Run()

		if err != nil {
			utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	}

	err = tableAccount.Delete("id", accountID).Run()

	if err != nil {
		return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}
