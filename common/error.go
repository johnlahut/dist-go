package common

import "log"

// FailOnError is a generic error logging function
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
