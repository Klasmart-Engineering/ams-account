package models

import "os"

func GetTableName(tableName string) string {
	var ret string = tableName

	if os.Getenv("SERVER_STAGE") != "prod" {
		ret += "_alphabeta"
	}

	return ret
}
