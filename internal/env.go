package internal

import (
	"fmt"
	"os"
)

func MustEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("env [%s] not set.", key))
	}
	return val
}
