// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

import (
	"cmp"
	"slices"
	"strings"
)

// Parser defines the type of a generic Parser function
// A few rules should be followed to prevent unexpected behaviour:
//   - A parser that errors must add an error
//   - A parser that errors should not change position of the states input
//   - A parser that consumed some input should advance with state.MoveBy()
type Parser[Output any] func(State) (State, Output)

// Separator is a generic type for separators (byte, rune, []byte or string)
type Separator interface {
	~rune | ~byte | ~string | ~[]byte
}

// Input is the input data for all the parsers.
// It can be either UTF-8 encoded text (a.k.a. string) or raw bytes.
// The parsers store and advance the position within the data but never change the data itself.
// This allows good error reporting including the full line of text containing the error.
type Input struct {
	// Go is fundamentally working with bytes and can interpret them as strings or as containing runes.
	// There are no standard library functions for handling []rune or the like.
	bytes []byte
	pos   uint // position in the sequence a.k.a. the *byte* index
}

// pcbError is an error message from the parser.
// It consists of the text itself and the position in the input where it happened.
type pcbError struct {
	text      string
	line, col uint
}

func ZeroOf[T any]() T {
	var t T
	return t
}

// State represents the current state of a parser.
// It consists of the Input, the pointOfNoReturn mark
// and a collection of error messages.
type State struct {
	input           Input
	pointOfNoReturn uint // mark set by the NoWayBack parser
	Messages        []pcbError
}

// NewFromString creates a new parser state from the input data.
func NewFromString(input string) State {
	return State{input: Input{bytes: []byte(input)}}
}

// NewFromBytes creates a new parser state from the input data.
// This is useful for binary or mixed binary/text parsers.
func NewFromBytes(input []byte) State {
	return State{input: Input{bytes: input}}
}

func (st State) AtEnd() bool {
	return st.input.pos >= uint(len(st.input.bytes))
}

func (st State) BytesRemaining() uint {
	return uint(len(st.input.bytes)) - st.input.pos
}

func (st State) CurrentString() string {
	return string(st.input.bytes[st.input.pos:])
}

func (st State) CurrentBytes() []byte {
	return st.input.bytes[st.input.pos:]
}

func (st State) StringTo(remaining State) string {
	return string(st.BytesTo(remaining))
}

func (st State) BytesTo(remaining State) []byte {
	if remaining.input.pos < st.input.pos {
		return []byte{}
	}
	if remaining.input.pos > uint(len(st.input.bytes)) {
		return st.input.bytes[st.input.pos:]
	}
	return st.input.bytes[st.input.pos:remaining.input.pos]
}

func (st State) MoveBy(countBytes uint) State {
	st.input.pos += countBytes
	ulen := uint(len(st.input.bytes))
	if st.input.pos > ulen { // prevent overrun
		st.input.pos = ulen
	}
	return st
}

func (st State) Moved(other State) bool {
	return st.input.pos != other.input.pos
}

func (st State) SignalNoWayBack() State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, st.input.pos+1)
	return st
}

func (st State) NoWayBack() bool {
	return st.pointOfNoReturn > st.input.pos
}

// Success return the State with NoWayBack saved from
// the subState.
func (st State) Success(subState State) State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, subState.pointOfNoReturn)
	return st
}

// Failure return the State with errors kept from
// the subState.
func (st State) Failure(subState State) State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, subState.pointOfNoReturn)

	st.Messages = append(st.Messages, subState.Messages...)
	slices.SortFunc(st.Messages, func(a, b pcbError) int { // always keep them sorted
		i := cmp.Compare(a.line, b.line)
		if i != 0 {
			return i
		}
		return cmp.Compare(a.col, b.col)
	})

	return st
}

// Failed returns whether this state is in a failed state or not.
func (st State) Failed() bool {
	return len(st.Messages) > 0
}

// AddError adds the messages to this state at the current position.
func (st State) AddError(message string) State {
	return st.Failure(State{Messages: []pcbError{{text: message, line: st.input.pos}}})
}

// Error returns a human readable error string.
func (st State) Error() string {
	fullMsg := strings.Builder{}
	for _, message := range st.Messages {
		fullMsg.WriteString("expected ")
		fullMsg.WriteString(message.text)
		fullMsg.WriteString("[line, column]: source line")
		fullMsg.WriteRune('\n')
	}

	return fullMsg.String()
}

// BetterOf returns the more advanced (in the input) state of the two.
// This should be used for parsers that are alternatives. So the best error is kept.
func BetterOf(state, other State) State {
	if state.input.pos < other.input.pos {
		return other
	}
	return state
}
