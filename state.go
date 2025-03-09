package comb

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ============================================================================
// This File contains only the State data structure and all of its methods.
// ============================================================================

// State represents the current state of a parser.
type State struct {
	constant *ConstState
	pos      int           // current position in the input a.k.a. the *byte* index
	prevNl   int           // position of newline preceding 'pos' (-1 for line==1)
	line     int           // current line number
	safeSpot int           // mark set by the SafeSpot parser
	errors   []ParserError // errors that have been handled
}

// ============================================================================
// Handle Input
//

func (st State) AtEnd() bool {
	return st.pos >= st.constant.n
}

func (st State) BytesRemaining() int {
	return st.constant.n - st.pos
}

func (st State) CurrentString() string {
	if st.constant.binary && len(st.constant.text) < st.constant.n {
		st.constant.text = string(st.constant.bytes)
	}
	return st.constant.text[st.pos:]
}

func (st State) CurrentBytes() []byte {
	if !st.constant.binary && len(st.constant.bytes) < st.constant.n {
		st.constant.bytes = []byte(st.constant.text)
	}
	return st.constant.bytes[st.pos:]
}

func (st State) CurrentPos() int {
	return st.pos
}

func (st State) StringTo(remaining State) string {
	if remaining.pos < st.pos {
		return ""
	}
	if st.constant.binary && len(st.constant.text) < st.constant.n {
		st.constant.text = string(st.constant.bytes)
	}
	if remaining.pos > len(st.constant.text) {
		return st.constant.text[st.pos:]
	}
	return st.constant.text[st.pos:remaining.pos]
}

func (st State) BytesTo(remaining State) []byte {
	if remaining.pos < st.pos {
		return []byte{}
	}
	if !st.constant.binary && len(st.constant.bytes) < st.constant.n {
		st.constant.bytes = []byte(st.constant.text)
	}
	if remaining.pos > len(st.constant.bytes) {
		return st.constant.bytes[st.pos:]
	}
	return st.constant.bytes[st.pos:remaining.pos]
}

func (st State) ByteCount(remaining State) int {
	if remaining.pos < st.pos {
		return 0 // we never go back so we don't give negative count back
	}
	n := st.constant.n
	if remaining.pos > n {
		return n - st.pos
	}
	return remaining.pos - st.pos
}

func (st State) MoveBy(countBytes int) State {
	if countBytes < 0 {
		countBytes = 0
	}

	pos := st.pos
	n := min(st.constant.n, pos+countBytes)
	st.pos = n

	if !st.constant.binary {
		moveText := st.constant.text[pos:n]
		lastNlPos := strings.LastIndexByte(moveText, '\n') // this is Unicode safe!!!
		if lastNlPos >= 0 {
			st.prevNl += lastNlPos + 1 // this works even if '\n' wasn't found at all
			st.line += strings.Count(moveText, "\n")
		}
	}

	return st
}

func (st State) MoveBackTo(pos int) State {
	if pos <= 0 {
		st.pos = 0
		st.prevNl = -1
		st.line = 1
		return st
	}

	if pos >= st.pos {
		return st
	}

	curPos := st.pos
	st.pos = pos

	if !st.constant.binary {
		moveText := st.constant.text[pos:curPos]
		lastNlPos := strings.LastIndexByte(st.constant.text[:pos], '\n') // this is Unicode safe!!!
		st.line -= strings.Count(moveText, "\n")
		st.prevNl = lastNlPos // this works even if '\n' wasn't found at all
	}

	return st
}

func (st State) Moved(other State) bool {
	return st.pos != other.pos
}

// Delete1 moves forward in the input, thus simulating deletion of input.
// For binary input it moves forward by a byte otherwise by a UNICODE rune.
func (st State) Delete1() State {
	if st.constant.binary {
		return st.MoveBy(1)
	}

	r, size := utf8.DecodeRuneInString(st.CurrentString())
	if r == utf8.RuneError {
		return st.MoveBy(1) // try to correct the error
	}
	return st.MoveBy(size)
}

// ============================================================================
// Handle success and failure
//

// SaveError saves an error and returns the new state.
func (st State) SaveError(err *ParserError) State {
	if err != nil {
		st.errors = append(st.errors, *err)
	}
	if st.constant.maxErrors > 0 && len(st.errors) >= st.constant.maxErrors {
		st.errors = append(st.errors, *(st.NewSemanticError("too many errors, giving up")))
		st.MoveBy(st.BytesRemaining()) // give up: move to end
	}
	return st
}

// NewSyntaxError creates a syntax error with the message and arguments at
// the current state position.
// For syntax errors `expected ` is prepended to the message and the usual
// position and source line including marker are appended.
func (st State) NewSyntaxError(msg string, args ...interface{}) *ParserError {
	return st.NewSemanticError(`expected `+msg, args...)
}

// NewSemanticError creates a semantic error with the message and arguments at
// the current state position.
// The usual position and source line including marker are appended to the message.
func (st State) NewSemanticError(msg string, args ...interface{}) *ParserError {
	newErr := &ParserError{text: fmt.Sprintf(msg, args...), pos: st.pos, binary: st.constant.binary, parserID: -1}
	if st.constant.binary { // the rare binary case is misusing the text case data a bit...
		newErr.line, newErr.col, newErr.srcLine = st.bytesAround(st.pos)
	} else {
		newErr.line, newErr.col, newErr.srcLine = st.textAround(st.pos)
	}
	return newErr
}

// HasError returns true if any errors are registered.
// (Errors that would be returned by State.Errors())
func (st State) HasError() bool {
	return len(st.errors) > 0
}

// ============================================================================
// Produce error messages and give them back
//

// CurrentSourceLine returns the source line corresponding to the current position
// including [line:column] at the start and a marker at the exact error position.
// This should be used for reporting errors that are detected later.
// The binary case is handled accordingly.
func (st State) CurrentSourceLine() string {
	if st.constant.binary {
		return formatBinaryLine(st.bytesAround(st.pos))
	} else {
		return formatSrcLine(st.textAround(st.pos))
	}
}

func (st State) bytesAround(pos int) (line, col int, srcLine string) {
	start := max(0, pos-8)
	end := min(start+16, st.constant.n)
	if end-start < 16 { // try to fill up from the other end...
		start = max(0, end-16)
	}
	srcLine = string(st.constant.bytes[start:end])
	return start, pos - start, srcLine
}

func (st State) textAround(pos int) (line, col int, srcLine string) {
	if pos < 0 {
		pos = 0
	}
	if len(st.constant.text) == 0 {
		return 1, 0, ""
	}
	if pos > st.prevNl { // pos is ahead of prevNL => search forward
		return st.whereForward(pos, st.line, st.prevNl)
	} else if pos <= st.prevNl-pos { // pos is too far back => search from start
		return st.whereForward(pos, 1, -1)
	} else { // pos is just a little back => search backward
		return st.whereBackward(pos, st.line, st.prevNl)
	}
}
func (st State) whereForward(pos, lineNum, prevNl int) (line, col int, srcLine string) {
	text := st.constant.text
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
	text := st.constant.text
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
		return lineNum, pos - prevNl - 1, string(st.constant.text[prevNl+1 : nextNl]), true
	}
	return 1, 0, "", false
}

// Errors returns all error messages accumulated by the state as a Go error.
// Multiple errors have been joined (by errors.Join()).
func (st State) Errors() error {
	if len(st.errors) == 0 {
		return nil
	}

	goErrors := make([]error, len(st.errors))
	for i, pe := range st.errors {
		goErrors[i] = errors.New(singleErrorMsg(pe))
	}

	return errors.Join(goErrors...)
}

// MoveSafeSpot returns the state with the safe spot moved to the current position.
func (st State) MoveSafeSpot() State {
	st.safeSpot = max(st.safeSpot, st.pos)
	return st
}

// SafeSpotMoved is true iff the safe spot is different between the 2 states.
func (st State) SafeSpotMoved(other State) bool {
	return st.safeSpot != other.safeSpot
}
