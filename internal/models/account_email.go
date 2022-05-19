package models

const (
	TABLE_NAME_ACCOUNT_EMAIL = "account_emails"
	ACCOUNT_EMAIL_GSI_ACCID  = "accId"
)

type AccountEmail struct {
	Email string `dynamo:"email,hash"`
	AccID string `dynamo:"accId" index:"accId,hash"`
}
