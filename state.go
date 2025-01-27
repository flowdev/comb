package gomme

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
	input     Input
	saveSpot  int           // mark set by the SafeSpot parser
	recover   bool          // recover from errors
	oldErrors []ParserError // errors that are or have been handled
}

// ============================================================================
// Handle Input
//

func (st State) AtEnd() bool {
	return st.input.pos >= st.input.n
}

func (st State) BytesRemaining() int {
	return st.input.n - st.input.pos
}

func (st State) CurrentString() string {
	if st.input.binary && len(st.input.text) < st.input.n {
		st.input.text = string(st.input.bytes)
	}
	return st.input.text[st.input.pos:]
}

func (st State) CurrentBytes() []byte {
	if !st.input.binary && len(st.input.bytes) < st.input.n {
		st.input.bytes = []byte(st.input.text)
	}
	return st.input.bytes[st.input.pos:]
}

func (st State) CurrentPos() int {
	return st.input.pos
}

func (st State) StringTo(remaining State) string {
	if remaining.input.pos < st.input.pos {
		return ""
	}
	if st.input.binary && len(st.input.text) < st.input.n {
		st.input.text = string(st.input.bytes)
	}
	if remaining.input.pos > len(st.input.text) {
		return st.input.text[st.input.pos:]
	}
	return st.input.text[st.input.pos:remaining.input.pos]
}

func (st State) BytesTo(remaining State) []byte {
	if remaining.input.pos < st.input.pos {
		return []byte{}
	}
	if !st.input.binary && len(st.input.bytes) < st.input.n {
		st.input.bytes = []byte(st.input.text)
	}
	if remaining.input.pos > len(st.input.bytes) {
		return st.input.bytes[st.input.pos:]
	}
	return st.input.bytes[st.input.pos:remaining.input.pos]
}

func (st State) ByteCount(remaining State) int {
	if remaining.input.pos < st.input.pos {
		return 0 // we never go back so we don't give negative count back
	}
	n := st.input.n
	if remaining.input.pos > n {
		return n - st.input.pos
	}
	return remaining.input.pos - st.input.pos
}

func (st State) MoveBy(countBytes int) State {
	if countBytes < 0 {
		countBytes = 0
	}

	pos := st.input.pos
	n := min(st.input.n, pos+countBytes)
	st.input.pos = n

	if !st.input.binary {
		moveText := st.input.text[pos:n]
		lastNlPos := strings.LastIndexByte(moveText, '\n') // this is Unicode safe!!!
		if lastNlPos >= 0 {
			st.input.prevNl += lastNlPos + 1 // this works even if '\n' wasn't found at all
			st.input.line += strings.Count(moveText, "\n")
		}
	}

	return st
}

func (st State) MoveBackTo(pos int) State {
	if pos <= 0 {
		st.input.pos = 0
		st.input.prevNl = -1
		st.input.line = 1
		return st
	}

	if pos >= st.input.pos {
		return st
	}

	curPos := st.input.pos
	st.input.pos = pos

	if !st.input.binary {
		moveText := st.input.text[pos:curPos]
		lastNlPos := strings.LastIndexByte(st.input.text[:pos], '\n') // this is Unicode safe!!!
		st.input.line -= strings.Count(moveText, "\n")
		st.input.prevNl = lastNlPos // this works even if '\n' wasn't found at all
	}

	return st
}

func (st State) Moved(other State) bool {
	return st.input.pos != other.input.pos
}

// Delete1 moves forward in the input, thus simulating deletion of input.
// For binary input it moves forward by a byte otherwise by a UNICODE rune.
func (st State) Delete1() State {
	if st.input.binary {
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
		st.oldErrors = append(st.oldErrors, *err)
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
	newErr := &ParserError{text: fmt.Sprintf(msg, args...), pos: st.input.pos, binary: st.input.binary, parserID: -1}
	if st.input.binary { // the rare binary case is misusing the text case data a bit...
		newErr.line, newErr.col, newErr.srcLine = st.bytesAround(st.input.pos)
	} else {
		newErr.line, newErr.col, newErr.srcLine = st.textAround(st.input.pos)
	}
	return newErr
}

// HasError returns true if any handled errors are registered.
// (Errors that would be returned by State.Errors())
func (st State) HasError() bool {
	return len(st.oldErrors) > 0
}

// ============================================================================
// Produce error messages and give them back
//

// CurrentSourceLine returns the source line corresponding to the current position
// including [line:column] at the start and a marker at the exact error position.
// This should be used for reporting errors that are detected later.
// The binary case is handled accordingly.
func (st State) CurrentSourceLine() string {
	if st.input.binary {
		return formatBinaryLine(st.bytesAround(st.input.pos))
	} else {
		return formatSrcLine(st.textAround(st.input.pos))
	}
}

func (st State) bytesAround(pos int) (line, col int, srcLine string) {
	start := max(0, pos-8)
	end := min(start+16, st.input.n)
	if end-start < 16 { // try to fill up from the other end...
		start = max(0, end-16)
	}
	srcLine = string(st.input.bytes[start:end])
	return start, pos - start, srcLine
}

func (st State) textAround(pos int) (line, col int, srcLine string) {
	if pos < 0 {
		pos = 0
	}
	if len(st.input.text) == 0 {
		return 1, 0, ""
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
	text := st.input.text
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
	text := st.input.text
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
		return lineNum, pos - prevNl - 1, string(st.input.text[prevNl+1 : nextNl]), true
	}
	return 1, 0, "", false
}

// Errors returns all error messages accumulated by the state as a Go error.
// Multiple errors have been joined (by errors.Join()).
func (st State) Errors() error {
	if len(st.oldErrors) == 0 {
		return nil
	}

	goErrors := make([]error, len(st.oldErrors))
	for i, pe := range st.oldErrors {
		goErrors[i] = errors.New(singleErrorMsg(pe))
	}

	return errors.Join(goErrors...)
}

// SaveSpotMoved is true iff the saveSpot is different between the 2 states.
func (st State) SaveSpotMoved(other State) bool {
	return st.saveSpot != other.saveSpot
}
