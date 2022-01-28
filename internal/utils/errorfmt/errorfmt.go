package errorfmt

import (
	"fmt"
)

// Error formats an error with its normal error value
func Error(err error) string {
	return fmt.Sprintf("%v", err)
}

// Full formats an error with all available info, i.e. the stack trace is included
func Full(err error) string {
	return fmt.Sprintf("%+v", err)
}
