package errorfmt

import (
	"fmt"
)

func Error(err error) string {
	return fmt.Sprintf("%v", err)
}
func Full(err error) string {
	return fmt.Sprintf("%+v", err)
}
