package util

import "log"

func Log(key string, args ...interface{}) {
	log.Printf(key, args...)
}
