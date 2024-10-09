// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"
)

const DefaultMaxDel = 3 // DefaultMaxDel of 3 is rather conservative but good enough in real life

type ParsingMode int

const (
	ParsingModeHappy ParsingMode = iota
	ParsingModeError
	ParsingModeHandle
	ParsingModeRecord
	ParsingModeChoose
	ParsingModePlay
)

// Recoverer is a simplified parser that only returns the number of bytes
// to reach a safe state.
// If it can't recover it should return -1.
//
// A Recoverer is used for recovering from an error in the input.
// It helps to move forward to the next safe spot
// (point of no return / NoWayBack parser).
// A Recoverer will be used by the NoWayBack parser if it's sub-parser provides it.
// Otherwise, NoWayBack will have to try the sub-parser until it succeeds moving
// forward 1 byte at a time. :(
type Recoverer func(state State) int

// Parser defines the type of a generic Parser
// A few rules should be followed to prevent unexpected behaviour:
//   - A parser that errors must add an error to the state
//   - A parser that errors should not change position of the states input
//   - A parser that consumed some input must advance with state.MoveBy()
type Parser[Output any] interface {
	Expected() string
	It(state State) (State, Output)
	Recoverer() Recoverer
}

type prsr[Output any] struct {
	expected  string
	it        func(state State) (State, Output)
	recoverer Recoverer
}

// NewParser is THE way to create parsers.
func NewParser[Output any](expected string, parse func(State) (State, Output), recover Recoverer) Parser[Output] {
	return prsr[Output]{
		expected:  expected,
		it:        parse,
		recoverer: recover,
	}
}

func (p prsr[Output]) Expected() string {
	return p.expected
}

func (p prsr[Output]) It(state State) (State, Output) {
	return p.it(state)
}

func (p prsr[Output]) Recoverer() Recoverer {
	return p.recoverer
}

type lazyprsr[Output any] struct {
	once       sync.Once
	makePrsr   func() Parser[Output]
	cachedPrsr Parser[Output]
}

// LazyParser just stores a function that creates the parser and evaluates the function later.
// This allows to defer the call to NewParser() and thus to define recursive grammars.
func LazyParser[Output any](makeParser func() Parser[Output]) Parser[Output] {
	return &lazyprsr[Output]{makePrsr: makeParser}
}

func (lp *lazyprsr[Output]) ensurePrsr() {
	lp.cachedPrsr = lp.makePrsr()
}

func (lp *lazyprsr[Output]) Expected() string {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Expected()
}

func (lp *lazyprsr[Output]) It(state State) (State, Output) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.It(state)
}

func (lp *lazyprsr[Output]) Recoverer() Recoverer {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Recoverer()
}

// RunOnString runs a parser on text input and returns the output and error(s).
func RunOnString[Output any](maxDel int, input string, parse Parser[Output]) (Output, error) {
	return run(NewFromString(maxDel, input), parse)
}

// RunOnBytes runs a parser on binary input and returns the output and error(s).
// This is useful for binary or mixed binary/text parsers.
func RunOnBytes[Output any](maxDel int, input []byte, parse Parser[Output]) (Output, error) {
	return run(NewState(maxDel, input), parse)
}

func run[Output any](state State, parse Parser[Output]) (Output, error) {
	newState, output := HandleAllErrors(state, parse)
	if len(newState.oldErrors) == 0 {
		return output, nil
	}
	return ZeroOf[Output](), pcbErrorsToGoErrors(newState.oldErrors)
}

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
	pos       int // pos is the byte index in the input (state.input.pos)
	line, col int // col is the 0-based byte index within srcLine; convert to 1-based rune index for user
	srcLine   string
}

func ZeroOf[T any]() T {
	var t T
	return t
}

type delErrHand struct {
	min     int        // minimal deletion
	deleted int        // number of bytes that have been "deleted"
	offset  int        // number of bytes the currently handled error happened after state.input.pos
	errors  []pcbError // errors during handling
}

type updErrHand struct {
	minDel int        // minimal deletion
	curDel int        // current deletion (this value goes from minDel to maxDel)
	errors []pcbError // errors during handling
}

type errHand struct {
	binary bool       // true if we should delete bytes in error recovery
	maxDel int        // maximum number of runes or bytes that should be deleted for error recovery
	err    *pcbError  // error that is currently handled
	del    delErrHand // for handling errors by "deleting" input
	upd    updErrHand // for handling errors by "updating" input
}

// State represents the current state of a parser.
// It consists of the Input, the pointOfNoReturn mark
// and a collection of error messages.
type State struct {
	mode            ParsingMode // one of: happy, error, handle, record, choose, play
	input           Input
	pointOfNoReturn int        // mark set by SignalNoWayBack/NoWayBack parser
	newError        *pcbError  // error that hasn't been handled yet
	errHand         errHand    // everything for error handling
	oldErrors       []pcbError // errors that are or have been handled
}

// NewFromString creates a new parser state from the input data.
func NewFromString(maxDel int, input string) State {
	state := NewState(maxDel, []byte(input))
	state.errHand.binary = false
	return state
}

// NewState creates a new parser state from the input data.
func NewState(maxDel int, input []byte) State {
	if maxDel <= 0 {
		maxDel = DefaultMaxDel
	}
	return State{
		input:           Input{bytes: input, line: 1, prevNl: -1},
		pointOfNoReturn: -1,
		errHand:         errHand{binary: true, maxDel: maxDel},
	}
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

func (st State) ByteCount(remaining State) int {
	if remaining.input.pos < st.input.pos {
		return 0 // we never go back so we don't give negative count back
	}
	n := len(st.input.bytes)
	if remaining.input.pos > n {
		return n - st.input.pos
	}
	return remaining.input.pos - st.input.pos
}

func (st State) MoveBy(countBytes uint) State {
	pos := st.input.pos
	n := min(len(st.input.bytes), pos+int(countBytes))
	st.input.pos = n

	moveText := string(st.input.bytes[pos:n])
	lastNlPos := strings.LastIndexByte(moveText, '\n') // this is Unicode safe!!!
	if lastNlPos >= 0 {
		st.input.prevNl += lastNlPos + 1 // this works even if '\n' wasn't found at all
		st.input.line += strings.Count(moveText, "\n")
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

// Failure returns the State with the error, pointOfNoReturn and mode kept from
// the subState.
func (st State) Failure(subState State) State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, subState.pointOfNoReturn)
	st.mode = subState.mode

	if subState.newError != nil { // should be true
		st.newError = subState.newError
	}

	return st
}

// NewError sets an error with the messages in this state at the current position.
// `newState` is the State most advanced in the input (can be the same as this State).
// `futureConsumption` is the input consumption of necessary future parsers.
func (st State) NewError(message string) State {
	line, col, srcLine := st.where(st.input.pos)
	newErr := &pcbError{
		text: message,
		pos:  st.input.pos, line: line, col: col,
		srcLine: srcLine,
	}

	switch st.mode {
	case ParsingModeHappy:
		st.mode = ParsingModeError
		st.newError = newErr
	case ParsingModeError:
		// should NOT happen but keep error furthest in the input
		if st.newError == nil || st.newError.pos < st.input.pos {
			st.newError = newErr
		}
	case ParsingModeHandle:
		if st.handlingNewError(newErr) {
			st.newError = nil
			st.mode = ParsingModeRecord
		} else {
			if st.newError == nil || st.newError.pos < st.input.pos {
				st.newError = newErr
			}
		}
	case ParsingModeRecord:
		// ignore error (we simulate the happy path)
	case ParsingModeChoose:
		// ignore error (we simulate the happy path)
	case ParsingModePlay:
		st.newError = newErr
	}
	return st
}

// Failed returns whether this state is in a failed state or not.
func (st State) Failed() bool {
	return st.newError != nil
}

// ============================================================================
// Produce error messages and give them back
//

func (st State) CurrentSourceLine() string {
	return formatSrcLine(st.where(st.input.pos))
}

func (st State) where(pos int) (line, col int, srcLine string) {
	if len(st.input.bytes) == 0 {
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
		return lineNum, pos - prevNl - 1, string(st.input.bytes[prevNl+1 : nextNl]), true
	}
	return 1, 0, "", false
}

// Error returns a human readable error string.
func (st State) Error() string {
	slices.SortFunc(st.oldErrors, func(a, b pcbError) int { // always keep them sorted
		i := cmp.Compare(a.line, b.line)
		if i != 0 {
			return i
		}
		return cmp.Compare(a.col, b.col)
	})

	fullMsg := strings.Builder{}
	for _, pcbErr := range st.oldErrors {
		fullMsg.WriteString(singleErrorMsg(pcbErr))
		fullMsg.WriteRune('\n')
	}
	if st.newError != nil {
		fullMsg.WriteString(singleErrorMsg(*st.newError))
		fullMsg.WriteRune('\n')
	}

	return fullMsg.String()
}

func singleErrorMsg(pcbErr pcbError) string {
	fullMsg := strings.Builder{}
	fullMsg.WriteString("expected ")
	fullMsg.WriteString(pcbErr.text)
	fullMsg.WriteString(formatSrcLine(pcbErr.line, pcbErr.col, pcbErr.srcLine))

	return fullMsg.String()
}

func formatSrcLine(line, col int, srcLine string) string {
	result := strings.Builder{}
	lineStart := srcLine[:col]
	result.WriteString(lineStart)
	result.WriteRune(0x25B6)
	result.WriteString(srcLine[col:])
	return fmt.Sprintf(" [%d:%d] %q",
		line, utf8.RuneCountInString(lineStart)+1, result.String()) // columns for the user start at 1
}

// ============================================================================
// Parser accounting (for fair decisions about which parser path is best)
//

func pcbErrorsToGoErrors(pcbErrors []pcbError) error {
	if len(pcbErrors) == 0 {
		return nil
	}

	goErrors := make([]error, len(pcbErrors))
	for i, pe := range pcbErrors {
		goErrors[i] = errors.New(singleErrorMsg(pe))
	}

	return errors.Join(goErrors...)
}

func HandleAllErrors[Output any](state State, parse Parser[Output]) (State, Output) {
	var output Output
	var newState State

	curState := state
	curState.errHand.del.min = 1    // we have to delete at least 1 byte to fix anything
	curState.errHand.upd.minDel = 0 // update needn't delete anything at all => insert
	for {
		if !curState.Failed() {
			newState, output = parse.It(curState)
			if !newState.Failed() {
				newState.errHand.err = nil
				break
			}
		} else {
			newState = curState
		}

		if !newState.handlingNewError(newState.newError) {
			newState.mode = ParsingModeHappy
		}

		switch newState.mode {
		case ParsingModeHappy: // if no err handling yet -> start handling & delete single bytes
			newState.oldErrors = append(newState.oldErrors, *newState.newError)
			newState.errHand.err = newState.newError
			newState.mode = ParsingModeError
		case ParsingModeError: // if deleted single bytes -> insert good input and possibly delete some bytes
			newState.errHand.upd.curDel = newState.errHand.upd.minDel
			newState.mode = ParsingModeRecord
		case ParsingModeRecord: // if updated input -> delete a bit more or try all again just harder
			if newState.errHand.upd.curDel < newState.errHand.maxDel {
				newState.errHand.upd.curDel++
			} else {
				newState = state
				oldConsumption := curState.errHand.maxDel
				if newState.errHand.maxDel <= oldConsumption { // we have already reached the end!!!
					newState.errHand.err = nil
					break
				}
				newState.errHand.del.min = oldConsumption
				newState.mode = ParsingModeError
			}
		}

		// let's try again
		newState.newError = nil
		curState = newState
	}

	newState.newError = nil
	newState.errHand.err = nil
	newState.mode = ParsingModeHappy
	newState.errHand.maxDel = 0
	newState.errHand.del = delErrHand{}
	newState.errHand.upd = updErrHand{}
	return newState, output
}

func HandleCurrentError[Output any](state State, parse Parser[Output]) (State, Output) {
	if !state.handlingNewError(state.newError) {
		return state, ZeroOf[Output]()
	}

	switch state.mode {
	case ParsingModeError: // try byte-wise deletion of input first
		var tryState State
		errOffset := state.errHand.err.pos - state.input.pos // this should be 0, but misbehaving parsers...

		for i := state.errHand.del.min; i <= state.errHand.maxDel; i++ {
			tryState = state.MoveBy(uint(i))
			newState, output := parse.It(tryState)
			// It will always be a new error because the position has changed.
			// But if this is called by the first combining parser,
			// the position won't change beyond the `tryState` if it fails directly.
			if !newState.Failed() || tryState.ByteCount(newState) > errOffset {
				newState.errHand.del.deleted = i
				newState.errHand.del.offset = errOffset
				return newState, output
			}
			// we failed again without really moving
			newState.oldErrors = append(newState.oldErrors, *newState.newError) // TODO: ERROR ID or 100 times the same error
			newState.errHand.err = newState.newError
			state.newError = nil
			tryState = newState
		}
	case ParsingModeHandle: // imitate insertion of correct input by ignoring the error
		state.newError = nil
		return state, ZeroOf[Output]()
	case ParsingModeRecord: // insert (ignore the error) + delete (move ahead)
		state.newError = nil
		return state.MoveBy(uint(state.errHand.upd.curDel)), ZeroOf[Output]()
	default:
		// intentionally do nothing
	}

	return state, ZeroOf[Output]()
}

func (st State) handlingNewError(newErr *pcbError) bool {
	if st.errHand.err == nil || newErr == nil {
		return false
	}
	return *st.errHand.err == *newErr
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
