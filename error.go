package gomme

import (
	"fmt"
	"strings"
)

// Error represents a parsing error. It holds the input that was being parsed,
// the error that was produced, whether this is a fatal error or there is no way back
// plus what was expected to match.
// If the error is fatal, we have to stop parsing of the file completely.
// If there is no way back we might be able to continue parsing AFTER the error position.
type Error struct {
	Input     InputBytes
	Err       error
	Fatal     bool
	NoWayBack bool
	Expected  []string
}

// NewError produces a new Error from the provided input and names of
// parsers expected to succeed.
func NewError(input InputBytes, expected ...string) *Error {
	return &Error{Input: input, Expected: expected}
}

// Error returns a human readable error string.
func (e *Error) Error() string {
	return fmt.Sprintf("expected %v", strings.Join(e.Expected, ", "))
}

// IsFatal returns true if the error is fatal.
func (e *Error) IsFatal() bool {
	return e.Fatal
}
