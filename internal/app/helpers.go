package app

import (
	"log"
	"os"
	"strconv"
)

func MustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}

func EnvString(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func MustEnvInt(key string) int {
	v := MustEnv(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("environment variable %q must be an integer, got %q", key, v)
	}
	return i
}

func EnvBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))
	return v
}
