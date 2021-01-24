package env

import (
	"fmt"
	"os"
)

func GetOrElse(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetOrPanic(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(fmt.Errorf("expecting env var %v to exist", key))
	}
	return value
}
