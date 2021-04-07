package mytokenlib

// MytokenError is a error type from the mytoken library
type MytokenError struct {
	err          string
	errorDetails string
}

func (err *MytokenError) Error() string {
	e := err.err
	if err.errorDetails != "" {
		e += ": " + err.errorDetails
	}
	return e
}

func newMytokenErrorFromError(e string, err error) *MytokenError {
	return &MytokenError{
		err:          e,
		errorDetails: err.Error(),
	}
}

const unexpectedResponse = "unexpected response from mytoken server"
const errorWhileHttp = "error while sending http request"
