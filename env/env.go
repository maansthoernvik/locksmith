// The env package provides utilities to register expected environment variables and their type.
// You can register either required or optional variables to be parsed from the current environment.
package env

import (
	"fmt"
	"os"
	"strconv"
)

type ErrorNotFound struct {
	name string
}

func newErrorNotFound(name string) error {
	return &ErrorNotFound{name: name}
}

func (err *ErrorNotFound) Error() string {
	return fmt.Sprintf("Did not find variable '%s'", err.name)
}

func GetOptionalBool(name string, def bool) (bool, error) {
	if v, e := os.LookupEnv(name); e {
		return strconv.ParseBool(v)
	}
	return def, nil
}

func GetRequiredBool(name string) (bool, error) {
	if v, e := os.LookupEnv(name); e {
		return strconv.ParseBool(v)
	}
	return false, newErrorNotFound(name)
}

func GetOptionalString(name string, def string) (string, error) {
	if v, e := os.LookupEnv(name); e {
		return v, nil
	}
	return def, nil
}

func GetRequiredString(name string) (string, error) {
	if v, e := os.LookupEnv(name); e {
		return v, nil
	}
	return "", newErrorNotFound(name)
}

func GetOptionalInteger(name string, def int) (int, error) {
	if v, e := os.LookupEnv(name); e {
		// by setting a base of 0, the base is implied by the string's format
		i64, err := strconv.ParseInt(v, 0, 0)
		return int(i64), err
	}
	return def, nil
}

func GetRequiredInteger(name string) (int, error) {
	if v, e := os.LookupEnv(name); e {
		// by setting a base of 0, the base is implied by the string's format
		i64, err := strconv.ParseInt(v, 0, 0)
		return int(i64), err
	}
	return 0, newErrorNotFound(name)
}
