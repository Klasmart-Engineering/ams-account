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
	err = tableAccountEmail.Get("accId", accountID).Index(models.ACCOUNT_EMAIL_GSI_ACCID).One(&resultAccountEmail)

	if err != nil {
		if err.Error() == "dynamo: no item found" {
			// DO NOTHING
		} else {
			return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	} else {
		err = tableAccountEmail.Delete("email", resultAccountEmail.Email).Run()

		if err != nil {
			return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	}

	var resultAccountPhoneNumber models.AccountPhoneNumber
	accountPhoneNumberTableName := models.GetTableName(models.TABLE_NAME_ACCOUNT_PHONE_NUMBER)
	tableAccountPhoneNumber := db.Table(accountPhoneNumberTableName)
	err = tableAccountPhoneNumber.Get("accId", accountID).Index(models.ACCOUNT_PHONENUMBER_GSI_ACCID).One(&resultAccountPhoneNumber)

	if err != nil {
		if err.Error() == "dynamo: no item found" {
			// DO NOTHING
		} else {
			return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	} else {
		err = tableAccountPhoneNumber.Delete("phoneNr", resultAccountPhoneNumber.PhoneNumber).Run()

		if err != nil {
			return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	}

	transactionTableName := models.GetTableName(models.TABLE_NAME_ACCOUNT_TRANSACTIONS)
	tableTransaction := db.Table(transactionTableName)

	var resultTransaction []models.AccountTransaction
	err = tableTransaction.Get("accId", accountID).All(&resultTransaction)

	if err != nil {

		return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
	}

	for i := 0; i < len(resultTransaction); i++ {
		transactionItem := resultTransaction[i]

		err = tableTransaction.Delete("accId", accountID).Range("transactionId", transactionItem.TransactionID).Run()

		if err != nil {
			return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
		}
	}

	err = tableAccount.Delete("id", accountID).Run()

	if err != nil {
		return utils.EchoHandleHTTPError(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}
