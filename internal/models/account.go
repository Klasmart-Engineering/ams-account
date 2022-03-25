package models

const (
	TABLE_NAME_ACCOUNT = "accounts"
)

type Account struct {
	ID string `dynamo:"id"`
}
