package internal

import (
	"os"
	"strconv"
)

// ContainsString returns boolean depending on existence item in items slice.
func ContainsString(item string, items []string) bool {
	for _, s := range items {
		if s == item {
			return true
		}
	}
	return false
}

// GetEnvOrDefaultString returns environment variable value or default value if variable not set.
func GetEnvOrDefaultString(envVar, defaultValue string) string {
	v := os.Getenv(envVar)
	if v == "" {
		return defaultValue
	}
	return v
}

// GetEnvOrDefaultInt returns environment variable value or default value (int) if variable not set.
func GetEnvOrDefaultInt(envVar string, defaultValue int) int {
	v, err := strconv.Atoi(os.Getenv(envVar))
	if err != nil {
		return defaultValue
	}
	return v
}
