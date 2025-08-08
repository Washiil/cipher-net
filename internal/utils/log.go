package utils

import "log"

func Verbose(enabled bool, format string, args ...interface{}) {
	if enabled {
		log.Printf(format, args...)
	}
}
