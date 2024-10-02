// Package redis demonstrates the usage of the gomme package to parse Redis'
// [RESP protocol] messages.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
package redis

import (
	"errors"
	"fmt"
	"github.com/oleiade/gomme/pcb"
	"strconv"
	"strings"

	"github.com/oleiade/gomme"
)

// ParseRESPMessage parses a Redis' [RESP protocol] message.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
func ParseRESPMessage(input string) (RESPMessage, error) {
	if len(input) < 3 {
		return RESPMessage{}, fmt.Errorf("malformed message %s; reason: %w", input, ErrMessageTooShort)
	}

	if !isValidMessageKind(MessageKind(input[0])) {
		return RESPMessage{}, fmt.Errorf("malformed message %s; reason: %w %c", input, ErrInvalidPrefix, input[0])
	}

	if input[len(input)-2] != '\r' || input[len(input)-1] != '\n' {
		return RESPMessage{}, fmt.Errorf("malformed message %s; reason: %w", input, ErrInvalidSuffix)
	}

	parser := pcb.Alternative(
		SimpleString(),
		Error(),
		Integer(),
		BulkString(),
		Array(),
	)

	return gomme.RunOnString(input, parser)
}

// ErrMessageTooShort is returned when a message is too short to be valid.
// A [RESP protocol] message is at least 3 characters long: the message kind
// prefix, the message content (which can be empty), and the gomme.CRLF suffix.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
var ErrMessageTooShort = errors.New("message too short")

// ErrInvalidPrefix is returned when a message kind prefix is not recognized.
// Valid [RESP Protocol] message kind prefixes are "+", "-", ":", and "$".
//
// [RESP Protocol]: https://redis.io/docs/reference/protocol-spec/
var ErrInvalidPrefix = errors.New("invalid message prefix")

// ErrInvalidSuffix is returned when a message suffix is not recognized.
// Every [RESP protocol] message ends with a gomme.CRLF.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
var ErrInvalidSuffix = errors.New("invalid message suffix")

// RESPMessage is a parsed Redis' [RESP protocol] message.
//
// It can hold either a simple string, an error, an integer, a bulk string,
// or an array. The kind of the message is available in the Kind field.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
type RESPMessage struct {
	Kind         MessageKind
	SimpleString *SimpleStringMessage
	Error        *ErrorStringMessage
	Integer      *IntegerMessage
	BulkString   *BulkStringMessage
	Array        *ArrayMessage
}

// MessageKind is the kind of a Redis' [RESP protocol] message.
type MessageKind string

// The many different kinds of Redis' [RESP protocol] messages map
// to their respective protocol message's prefixes.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
const (
	SimpleStringKind MessageKind = "+"
	ErrorKind        MessageKind = "-"
	IntegerKind      MessageKind = ":"
	BulkStringKind   MessageKind = "$"
	ArrayKind        MessageKind = "*"
	InvalidKind      MessageKind = "?"
)

// SimpleStringMessage is a simple string message parsed from a Redis'
// [RESP protocol] message.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
type SimpleStringMessage struct {
	Content string
}

// SimpleString is a parser for Redis' RESP protocol simple strings.
//
// Simple strings are strings that are not expected to contain newlines.
// Simple strings start with a "+" character, and end with a gomme.CRLF.
//
// Once parsed, the content of the simple string is available in the
// simpleString field of the result's RESPMessage.
func SimpleString() gomme.Parser[RESPMessage] {
	mapFn := func(message string) (RESPMessage, error) {
		if strings.ContainsAny(message, "\r\n") {
			return RESPMessage{}, fmt.Errorf("malformed simple string: %s", message)
		}

		return RESPMessage{
			Kind: SimpleStringKind,
			SimpleString: &SimpleStringMessage{
				Content: message,
			},
		}, nil
	}

	return pcb.Preceded(
		pcb.String(string(SimpleStringKind)),
		pcb.Map(pcb.UntilString("\r\n"), mapFn),
	)
}

// ErrorStringMessage is a parsed error string message from a Redis'
// [RESP protocol] message.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
type ErrorStringMessage struct {
	Kind    string
	Message string
}

// Error is a parser for Redis' RESP protocol errors.
//
// Errors are strings that start with a "-" character, and end with a gomme.CRLF.
//
// The error message is available in the Error field of the result's
// RESPMessage.
func Error() gomme.Parser[RESPMessage] {
	mapFn := func(message string) (RESPMessage, error) {
		if strings.ContainsAny(message, "\r\n") {
			return RESPMessage{}, fmt.Errorf("malformed error string: %s", message)
		}

		return RESPMessage{
			Kind: ErrorKind,
			Error: &ErrorStringMessage{
				Kind:    "ERR",
				Message: message,
			},
		}, nil
	}

	return pcb.Preceded(
		pcb.String(string(ErrorKind)),
		pcb.Map(pcb.UntilString("\r\n"), mapFn),
	)
}

// IntegerMessage is a parsed integer message from a Redis' [RESP protocol]
// message.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
type IntegerMessage struct {
	Value int
}

// Integer is a parser for Redis' RESP protocol integers.
//
// Integers are signed numerical values represented as string messages
// that start with a ":" character, and end with a gomme.CRLF.
//
// The integer value is available in the IntegerMessage field of the result's
// RESPMessage.
func Integer() gomme.Parser[RESPMessage] {
	mapFn := func(message string) (RESPMessage, error) {
		value, err := strconv.Atoi(message)
		if err != nil {
			return RESPMessage{}, err
		}

		return RESPMessage{
			Kind: IntegerKind,
			Integer: &IntegerMessage{
				Value: value,
			},
		}, nil
	}

	return pcb.Preceded(
		pcb.String(string(IntegerKind)),
		pcb.Map(pcb.UntilString("\r\n"), mapFn),
	)
}

// BulkStringMessage is a parsed bulk string message from a Redis' [RESP protocol]
// message.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
type BulkStringMessage struct {
	Data []byte
}

// BulkString is a parser for Redis' RESP protocol bulk strings.
//
// Bulk strings are binary-safe strings up to 512MB in size.
// Bulk strings start with a "$" character, and end with a gomme.CRLF.
//
// The bulk string's data is available in the BulkString field of the result's
// RESPMessage.
func BulkString() gomme.Parser[RESPMessage] {
	mapFn := func(length int64, message string) (RESPMessage, error) {
		if length < 0 {
			if length < -1 {
				return RESPMessage{}, fmt.Errorf(
					"unable to parse bulk string; "+
						"reason: negative length %d",
					length,
				)
			}

			if length == -1 && len(message) != 0 {
				return RESPMessage{}, fmt.Errorf(
					"malformed bulkstring: declared message size -1, and actual size differ %d",
					len(message),
				)
			}
		} else if int64(len(message)) != length {
			return RESPMessage{}, fmt.Errorf(
				"malformed bulkstring: declared message size %d, and actual size differ %d; message: %q",
				length, len(message), message,
			)
		}

		return RESPMessage{
			Kind: BulkStringKind,
			BulkString: &BulkStringMessage{
				Data: []byte(message),
			},
		}, nil
	}

	return pcb.Map2(
		sizePrefix(pcb.String(string(BulkStringKind))),
		pcb.Optional(
			pcb.UntilString("\r\n"),
		),
		mapFn,
	)
}

// ArrayMessage is a parsed array message from a Redis' [RESP protocol] message.
//
// [RESP protocol]: https://redis.io/docs/reference/protocol-spec/
type ArrayMessage struct {
	Elements []RESPMessage
}

// Array is a parser for Redis' RESP protocol arrays.
//
// Arrays are sequences of RESP messages.
// Arrays start with a "*" character, and end with a gomme.CRLF.
//
// The array's messages are available in the Array field of the result's
// RESPMessage.
func Array() gomme.Parser[RESPMessage] {
	mapFn := func(size int64, message []RESPMessage) (RESPMessage, error) {
		if int(size) == -1 {
			if len(message) != 0 {
				return RESPMessage{}, fmt.Errorf(
					"malformed array: declared message size -1, and actual size differ %d",
					len(message),
				)
			}
		} else {
			if len(message) != int(size) {
				return RESPMessage{}, fmt.Errorf(
					"malformed array: declared message size %d, and actual size differ %d",
					size,
					len(message),
				)
			}
		}

		return RESPMessage{
			Kind: ArrayKind,
			Array: &ArrayMessage{
				Elements: message,
			},
		}, nil
	}

	return pcb.Map2(
		sizePrefix(pcb.String(string(ArrayKind))),
		pcb.Many0(
			pcb.Alternative(
				SimpleString(),
				Error(),
				Integer(),
				BulkString(),
			),
		),
		mapFn,
	)
}

func sizePrefix(prefix gomme.Parser[string]) gomme.Parser[int64] {
	return pcb.Delimited(
		prefix,
		pcb.Int64(),
		pcb.CRLF(),
	)
}

func isValidMessageKind(kind MessageKind) bool {
	return kind == SimpleStringKind ||
		kind == ErrorKind ||
		kind == IntegerKind ||
		kind == BulkStringKind ||
		kind == ArrayKind
}
