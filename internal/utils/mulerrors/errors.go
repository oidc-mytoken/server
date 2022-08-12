package mulerrors

// Errors is an error type that supports multiple errors
type Errors []error

func (e Errors) Error() string {
	if len(e) == 1 {
		return e[0].Error()
	}

	msg := "multiple errors:"
	for _, err := range e {
		msg += "\n" + err.Error()
	}
	return msg
}
