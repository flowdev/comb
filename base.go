// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

import (
	"cmp"
	"fmt"
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
	bytes  []byte
	pos    int // current position in the sequence a.k.a. the *byte* index
	prevNl int // position of newline preceding 'pos' (-1 for line==1)
	line   int // current line number
}

// pcbError is an error message from the parser.
// It consists of the text itself and the position in the input where it happened.
type pcbError struct {
	text      string
	line, col int
	srcLine   string
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
	pointOfNoReturn int        // mark set by SignalNoWayBack/NoWayBack parser
	newErrors       []pcbError // errors that haven't been handled yet
	oldErrors       []pcbError // errors that have been handled already
}

// NewFromString creates a new parser state from the input data.
func NewFromString(input string) State {
	return NewFromBytes([]byte(input))
}

// NewFromBytes creates a new parser state from the input data.
// This is useful for binary or mixed binary/text parsers.
func NewFromBytes(input []byte) State {
	return State{input: Input{bytes: input, line: 1, prevNl: -1}, pointOfNoReturn: -1}
}

// ============================================================================
// Handle Input
//

func (st State) AtEnd() bool {
	return st.input.pos >= len(st.input.bytes)
}

func (st State) BytesRemaining() uint {
	return uint(len(st.input.bytes) - st.input.pos)
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
	if remaining.input.pos > len(st.input.bytes) {
		return st.input.bytes[st.input.pos:]
	}
	return st.input.bytes[st.input.pos:remaining.input.pos]
}

func (st State) MoveBy(countBytes uint) State {
	pos := st.input.pos
	n := min(len(st.input.bytes), pos+int(countBytes))
	st.input.pos = n

	moveText := string(st.input.bytes[pos:n])
	lastNlPos := strings.LastIndexByte(moveText, '\n') // this is Unicode safe!!!
	if lastNlPos >= 0 {
		st.input.prevNl += lastNlPos + 1 // this works even if '\n' wasn't found at all
		nlCount := strings.Count(moveText, "\n")
		st.input.line += nlCount
	}

	return st
}

func (st State) Moved(other State) bool {
	return st.input.pos != other.input.pos
}

// ============================================================================
// Handle success and failure
//

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

	st.newErrors = append(st.newErrors, subState.newErrors...)
	slices.SortFunc(st.newErrors, func(a, b pcbError) int { // always keep them sorted
		i := cmp.Compare(a.line, b.line)
		if i != 0 {
			return i
		}
		return cmp.Compare(a.col, b.col)
	})

	return st
}

// AddError adds the messages to this state at the current position.
func (st State) AddError(message string) State {
	line, col, srcLine := st.where(st.input.pos)

	return st.Failure(State{newErrors: []pcbError{
		{text: message, line: line, col: col, srcLine: srcLine},
	}, pointOfNoReturn: -1})
}

// Failed returns whether this state is in a failed state or not.
func (st State) Failed() bool {
	return len(st.newErrors) > 0
}

// ============================================================================
// Produce error messages and give them back
//

func (st State) where(pos int) (line, col int, srcLine string) {
	if len(st.input.bytes) == 0 {
		return 1, 1, ""
	}
	if pos > st.input.prevNl { // pos is ahead of prevNL => search forward
		return st.whereForward(pos, st.input.line, st.input.prevNl)
	} else if pos <= st.input.prevNl-pos { // pos is too far back => search from start
		return st.whereForward(pos, 1, -1)
	} else { // pos is just a little back => search backward
		return st.whereBackward(pos, st.input.line, st.input.prevNl)
	}
}
func (st State) whereForward(pos, lineNum, prevNl int) (line, col int, srcLine string) {
	text := string(st.input.bytes)
	var nextNl int // Position of next newline or end

	for {
		nextNl = strings.IndexByte(text[prevNl+1:], '\n')
		if nextNl < 0 {
			nextNl = len(text)
		} else {
			nextNl += prevNl + 1
		}

		stop := false
		line, col, srcLine, stop = st.tryWhere(prevNl, pos, nextNl, lineNum)
		if stop {
			return line, col, srcLine
		}
		prevNl = nextNl
		lineNum++
	}
}
func (st State) whereBackward(pos, lineNum, nextNl int) (line, col int, srcLine string) {
	text := string(st.input.bytes)
	var prevNl int // Line start (position of preceding newline)

	for {
		prevNl = strings.LastIndexByte(text[0:nextNl], '\n')
		lineNum--

		stop := false
		line, col, srcLine, stop = st.tryWhere(prevNl, pos, nextNl, lineNum)
		if stop {
			return line, col, srcLine
		}
		nextNl = prevNl
	}
}
func (st State) tryWhere(prevNl int, pos int, nextNl int, lineNum int) (line, col int, srcLine string, stop bool) {
	if prevNl < pos && pos <= nextNl {
		return lineNum, pos - prevNl, string(st.input.bytes[prevNl+1 : nextNl]), true
	}
	return 0, 0, "", false
}

// Error returns a human readable error string.
func (st State) Error() string {
	fullMsg := strings.Builder{}
	for _, pcbErr := range st.newErrors {
		srcLine := strings.Builder{}
		fullMsg.WriteString("expected ")
		fullMsg.WriteString(pcbErr.text)

		srcLine.WriteString(pcbErr.srcLine[:pcbErr.col])
		srcLine.WriteRune(0x25B6)
		srcLine.WriteString(pcbErr.srcLine[pcbErr.col:])
		fullMsg.WriteString(fmt.Sprintf(" [%d, %d]: %q\n", pcbErr.line, pcbErr.col, srcLine.String()))
	}

	return fullMsg.String()
}

// ============================================================================
// Misc. stuff
//

// SignalNoWayBack sets a point of no return mark at the current position.
func (st State) SignalNoWayBack() State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, st.input.pos)
	return st
}

// NoWayBack is true iff we crossed a point of no return.
func (st State) NoWayBack() bool {
	return st.pointOfNoReturn >= st.input.pos
}

// BetterOf returns the more advanced (in the input) state of the two.
// This should be used for parsers that are alternatives. So the best error is kept.
func BetterOf(state, other State) State {
	if state.input.pos < other.input.pos {
		return other
	}
	return state
}
