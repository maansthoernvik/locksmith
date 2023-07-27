package env

import (
	"fmt"
	"os"
	"strconv"
)

type EnvError struct {
	message string
}

func NewEnvError(message string) *EnvError {
	return &EnvError{message: message}
}

func (ee EnvError) Error() string {
	return ee.message
}

func OptionalString(name string) string {
	return optionalEnv(name)
}

func RequiredString(name string) string {
	return requiredEnv(name)
}

func OptionalUint16(name string) uint16 {
	val := optionalEnv(name)
	if len(val) > 0 {
		n, err := strconv.ParseUint(val, 10, 64)
		if err == nil {
			return uint16(n)
		}
	}
	return 0
}

func RequiredUint16(name string) uint16 {
	val := requiredEnv(name)
	n, err := strconv.ParseUint(val, 10, 64)
	if err == nil {
		return uint16(n)
	}
	panic(NewEnvError(fmt.Sprintf("Error converting %s to uint16", name)))
}

func optionalEnv(name string) string {
	if val, exists := os.LookupEnv(name); exists {
		return val
	}
	return ""
}

func requiredEnv(name string) string {
	if val, exists := os.LookupEnv(name); exists {
		return val
	}
	panic(NewEnvError(fmt.Sprintf("Could not find environment variable %s", name)))
}
