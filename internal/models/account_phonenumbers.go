package models

const (
	TABLE_NAME_ACCOUNT_PHONE_NUMBER = "account_phonenumbers"
	ACCOUNT_PHONENUMBER_GSI_ACCID   = "accId"
)

type AccountPhoneNumber struct {
	PhoneNumber string `dynamo:"phoneNr,hash"`
	AccID       string `dynamo:"accId" index:"accId,hash"`
}
