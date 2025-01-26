// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

import (
	"context"
	"log"
	"log/slog"
)

// ============================================================================
// Basic types
//

// Separator is a generic type for separators (byte, rune, []byte or string)
type Separator interface {
	~rune | ~byte | ~string | ~[]byte
}

// Recoverer is a simplified parser that only returns the number of bytes
// to reach a SaveSpot.
// If it can't recover it should return -1.
//
// A Recoverer is used for recovering from an error in the input.
// It helps to move forward to the next SaveSpot.
// The basic Recoverer will have to try the parser until it succeeds moving
// forward 1 rune/byte at a time. :(
type Recoverer func(state State) int

// Parser defines the type of a generic Parser.
// A few rules should be followed to prevent unexpected behaviour:
//   - A parser that errors must return the error
//   - A parser that errors should not change position of the states input
//   - A parser that consumed some input must advance with state.MoveBy()
type Parser[Output any] interface {
	ID() int32
	Expected() string
	It(State) (State, Output, *ParserError)
	Parse(state State) ParseResult // used by orchestrator and branch parsers
	IsSaveSpot() bool
	setSaveSpot() // used by SaveSpot parser
	Recover(State) int
	IsStepRecoverer() bool
	SwapRecoverer(Recoverer) // called during construction phase
	setID(int32)             // used by orchestrator; only sets own ID
}

// ============================================================================
// Running a parser
//

// RunOnString runs a parser on text input and returns the output and error(s).
// It uses default values for maximum number of "tokens" to delete for error handling,
// the number of recoverers to try and the deleter to use.
// It also uses the default value for the number of recursions to support.
func RunOnString[Output any](input string, parse Parser[Output]) (Output, error) {
	return RunOnState(NewFromString(input, true), parse)
}

// RunOnBytes runs a parser on binary input and returns the output and error(s).
// It uses default values for maximum number of "tokens" to delete for error handling,
// the number of recoverers to try and the deleter to use.
// It also uses the default value for the number of recursions to support.
// This is useful for binary or mixed binary/text parsers.
func RunOnBytes[Output any](input []byte, parse Parser[Output]) (Output, error) {
	return RunOnState(NewFromBytes(input, true), parse)
}

func RunOnState[Output any](state State, parse Parser[Output]) (Output, error) {
	return newOrchestrator(parse).parseAll(state)
}

// ============================================================================
// Input And Creating a State With It
//

// Input is the input data for all the parsers.
// It can be either UTF-8 encoded text (a.k.a. string) or raw bytes.
// The parsers store and advance the position within the data but never change the data itself.
// This allows good error reporting including the full line of text containing the error.
type Input struct {
	binary bool   // type of input (general)
	bytes  []byte // for binary input and parsers
	text   string // for string input and text parsers
	n      int    // length of the bytes or text
	pos    int    // current position in the input a.k.a. the *byte* index
	prevNl int    // position of newline preceding 'pos' (-1 for line==1)
	line   int    // current line number
}

func newInput(binary bool, bytes []byte, text string) Input {
	n := len(text)
	if binary {
		n = len(bytes)
	}
	return Input{
		binary: binary, bytes: bytes, text: text, n: n,
		pos: 0, prevNl: -1, line: 1,
	}
}

// NewFromString creates a new parser state from the input data.
func NewFromString(input string, recover bool) State {
	return newState(false, nil, input, recover)
}

// NewFromBytes creates a new parser state from the input data.
func NewFromBytes(input []byte, recover bool) State {
	return newState(true, input, "", recover)
}

// newState creates a new parser state from the input data.
func newState(binary bool, bytes []byte, text string, recover bool) State {
	return State{
		input:    newInput(binary, bytes, text),
		saveSpot: -1,
		recover:  recover,
	}
}

// ============================================================================
// Misc. stuff
//

// BetterOf returns the more advanced (in the input) state of the two.
// This should be used for parsers that are alternatives.
// So the best error is handled.
func BetterOf(state, other State) State {
	if state.input.pos < other.input.pos {
		return other
	}
	return state
}

// ZeroOf returns the zero value of some type.
func ZeroOf[T any]() T {
	var t T
	return t
}

// SetDebug sets the log level to debug if enabled or info otherwise.
func SetDebug(enable bool) {
	if enable {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		return
	}
	slog.SetLogLoggerLevel(slog.LevelInfo)
}

// Debugf logs the given message using `log.Printf` if the debug level is enabled.
func Debugf(msg string, args ...interface{}) {
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		log.Printf("DEBUG: "+msg, args...)
	}
}
