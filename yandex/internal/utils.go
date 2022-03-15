package internal

import (
	"os"
	"strconv"
)

func ContainsString(item string, items []string) bool {
	for _, s := range items {
		if s == item {
			return true
		}
	}
	return false
}

func GetEnvOrDefaultString(envVar, defaultValue string) string {
	v := os.Getenv(envVar)
	if v == "" {
		return defaultValue
	}
	return v
}

func GetEnvOrDefaultInt(envVar string, defaultValue int) int {
	v, err := strconv.Atoi(os.Getenv(envVar))
	if err != nil {
		return defaultValue
	}
	return v
}
