package models

const (
	TABLE_NAME_ACCOUNT_TRANSACTIONS = "account_transactions"
	// ACCOUNT_GSI_ACCID        = "accId"
)

type AccountTransactionItem struct {
	Price          int    `dynamo:"price" json:"price"`
	Currency       string `dynamo:"currency" json:"currency"`
	StartDate      int    `dynamo:"startTm" json:"startTm"`
	ExpirationDate int    `dynamo:"expirationTm" json:"expirationTm"`
}

type AccountTransaction struct {
	AccountID     string                             `dynamo:"accId,hash" json:"accId"`
	TransactionID string                             `dynamo:"transactionId,range" json:"transactionId"`
	Passes        map[string]*AccountTransactionItem `dynamo:"passes" json:"passes"`
	Products      map[string]*AccountTransactionItem `dynamo:"products" json:"products"`
	CreatedDate   int                                `dynamo:"createTm" json:"createTm"`
	UpdatedDate   int                                `dynamo:"updateTm" json:"updateTm"`
}
