package internal

import (
	"log/slog"
	"os"
)

func EnvOrDefault(key string, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		slog.Warn("env not set, using default value.",
			"key", key, "default", defaultValue)
		return defaultValue
	}

	slog.Info("env set.", "key", key, "default", defaultValue)
	return v
}

func CalcMovingAvg(n uint64, currentAvg, nextVal float64) float64 {
	return (float64(n)*currentAvg + nextVal) / float64(n+1)
}
