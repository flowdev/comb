package gomme

import (
	"fmt"
	"strings"
)

// Input is the input data for all the parsers.
// It can be either raw bytes for binary parsers or UTF-8 encoded text for text parsers.
// The parsers store and advance the position within the data but never change the data itself.
// This allows good error reporting including the full line of text containing the error.
type Input struct {
	// Go is fundamentally working with bytes and can interpret them as strings or as containing runes.
	// There are no standard library functions for handling []rune or the like.
	Bytes []byte
	Text  bool
	Pos   uint // position in the sequence a.k.a. the *byte* index
}

// NewInputFromString creates a new input data structure suitable for parsing from a string.
// Using this function will set the Input to text mode.
func NewInputFromString(input string) Input {
	return Input{Bytes: []byte(input), Text: true, Pos: 0}
}

// NewInputFromBytes creates a new input data structure suitable for parsing from a slice of bytes.
// Using this function will set the Input to binary mode.
func NewInputFromBytes(input []byte) Input {
	return Input{Bytes: input, Text: false, Pos: 0}
}

func (ip Input) AtEnd() bool {
	return ip.Pos >= uint(len(ip.Bytes))
}

func (ip Input) BytesRemaining() uint {
	return uint(len(ip.Bytes)) - ip.Pos
}

func (ip Input) CurrentString() string {
	return string(ip.Bytes[ip.Pos:])
}

func (ip Input) CurrentBytes() []byte {
	return ip.Bytes[ip.Pos:]
}

func (ip Input) StringTo(remaining Input) string {
	return string(ip.BytesTo(remaining))
}

func (ip Input) BytesTo(remaining Input) []byte {
	if remaining.Pos < ip.Pos {
		return []byte{}
	}
	if remaining.Pos > uint(len(ip.Bytes)) {
		return ip.Bytes[ip.Pos:]
	}
	return ip.Bytes[ip.Pos:remaining.Pos]
}

func (ip Input) MoveBy(countBytes uint) Input {
	ip2 := ip
	ip2.Pos += countBytes
	ulen := uint(len(ip2.Bytes))
	if ip2.Pos > ulen { // prevent overrun
		ip2.Pos = ulen
	}
	return ip2
}

// Separator is a generic type alias for separators (byte, rune, []byte or string)
type Separator interface {
	~rune | ~byte | ~string | ~[]byte
}

// Result is a generic parser result.
// NoWayBack allows to signal that the result is final and won't get better by parsing an alternative path.
type Result[Output any] struct {
	Output    Output
	NoWayBack bool
	Err       *Error
	Remaining Input
}

// Parser defines the type of a generic Parser function
type Parser[Output any] func(input Input) Result[Output]

// Success creates a Result with an output set from
// the result of successful parsing.
func Success[Output any](output Output, r Input) Result[Output] {
	return Result[Output]{Output: output, Err: nil, Remaining: r}
}

// Failure creates a Result with an error set from
// the result of failed parsing.
func Failure[Output any](err *Error, input Input) Result[Output] {
	var output Output
	return Result[Output]{Output: output, Err: err, Remaining: input}
}

// Error represents a parsing error. It holds the input that was being parsed,
// the error that was produced, whether this is a fatal error or there is no way back
// plus what was expected to match.
// If the error is fatal, we have to stop parsing of the file completely.
// If there is no way back we might be able to continue parsing AFTER the error position.
type Error struct {
	Input    Input
	Err      error
	Fatal    bool
	Expected []string
}

// NewError produces a new Error from the provided input and names of
// parsers expected to succeed.
func NewError(input Input, expected ...string) *Error {
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
