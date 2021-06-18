package utils

import "log"

// HandleError handles errors
func HandleError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
