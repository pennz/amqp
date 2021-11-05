package utils

import "log"

// FailOnError will do Fatalf with error message
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
