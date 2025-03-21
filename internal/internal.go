package internal

import "os"

const (
	ConfigFilePath = "config.yml"
)

func EnvOrDefault(key string, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}
