package utils

import "os"

func GetOsEnvWithDef(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
