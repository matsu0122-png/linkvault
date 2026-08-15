package service

// ValidationError marks an error as caused by bad input, as opposed to an
// unexpected failure (e.g. the database being unreachable). Handlers use
// errors.As to detect it: a ValidationError's message is safe to show the
// client and maps to 400, while any other error is treated as internal and
// maps to 500 without leaking its message.
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string { return e.msg }

func newValidationError(msg string) error {
	return &ValidationError{msg: msg}
}
