package environment

import (
	"os"
	"strconv"
	"strings"
)

// GetEnv retrieves an environment variable or returns a default value if not found
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvAsBool retrieves an environment variable as a boolean or returns a default value if not found
func GetEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	
	return boolValue
}

// GetEnvAsInt retrieves an environment variable as an integer or returns a default value if not found
func GetEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return intValue
}

// GetEnvAsSlice retrieves an environment variable as a slice or returns a default value if not found
func GetEnvAsSlice(key, sep string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	return strings.Split(value, sep)
}
