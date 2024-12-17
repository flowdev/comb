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
	"sync"
)

// Use the stringer package from the Go team for printing of names of enums:
//go:generate go run golang.org/x/tools/cmd/stringer@latest -linecomment -type ParsingMode,Ternary

// DefaultMaxDel of 3 is a compromise between speed and optimal fault tolerance
// (ANTLR is using 1)
// It is the maximum number of times a deleter is called in order to recover from an error.
const DefaultMaxDel = 3

// DefaultMaxRecover of 3 is a compromise between speed and minimal waste by recoverers.
// It is the number of times a recoverer is called in order to recover from an error.
// The actual number will only be smaller if there aren't enough recoverers
// to be found moving forward.
// The actual number can be larger because the recoverers of the FirstSuccessful parser
// are all tried or none.
const DefaultMaxRecover = 3

// ParsingMode is needed for error handling. See `ERROR_HANDLING.md` for details.
type ParsingMode int

const (
	// ParsingModeHappy - normal parsing until failure (forward)
	ParsingModeHappy ParsingMode = iota // happy
	// ParsingModeError - find previous SaveSpot (backward)
	ParsingModeError // error
	// ParsingModeHandle - find witness parser (1) again (forward)
	ParsingModeHandle // handle
	// ParsingModeRewind - find witness parser (1) again (backward)
	ParsingModeRewind // rewind
	// ParsingModeEscape - find the (best) next SaveSpot (forward)
	ParsingModeEscape // escape
)

type Ternary int

const (
	TernaryNo    Ternary = iota // no
	TernaryMaybe                // maybe
	TernaryYes                  // yes
)

// Separator is a generic type for separators (byte, rune, []byte or string)
type Separator interface {
	~rune | ~byte | ~string | ~[]byte
}

// Recoverer is a simplified parser that only returns the number of bytes
// to reach a safe state (SaveSpot).
// If it can't recover it should return -1.
//
// A Recoverer is used for recovering from an error in the input.
// It helps to move forward to the next safe spot (SaveSpot).
// A Recoverer will be used by the SaveSpot parser if it's sub-parser
// provides it.
// Otherwise, SaveSpot will have to try the sub-parser until it succeeds moving
// forward 1 token at a time. :(
type Recoverer func(state State) int

// Deleter is a simplified parser that only moves the position in the input
// forward. This simulates deletion of parts of the input without changing it.
//
// A Deleter is used for recovering from errors by `deleting` tokens in the input.
// `count` is the number of tokens to be deleted.
// Each Deleter implementation defines itself what a token really is.
type Deleter func(state State, count int) State

// ============================================================================
// Parser Interface And Its Implementations
//

// Parser defines the type of a generic Parser
// A few rules should be followed to prevent unexpected behaviour:
//   - A parser that errors must add an error to the state
//   - A parser that errors should not change position of the states input
//   - A parser that consumed some input must advance with state.MoveBy()
type Parser[Output any] interface {
	Expected() string
	It(State) (State, Output)
	PossibleWitness() bool
	MyRecoverer() Recoverer
	SwapMyRecoverer(Recoverer) Parser[Output]
	SaveSpotRecoverer(State) int
}

// ParserToZeroOutput converts a parser of one output type to a different
// output type with zero value.
// So only use this if the output of the parser is waste.
func ParserToZeroOutput[Output, S any](other Parser[S]) Parser[Output] {
	var zero Output

	parse := func(state State) (State, Output) {
		state, _ = other.It(state)
		return state, zero
	}
	return NewParser[Output](
		other.Expected(),
		parse,
		other.PossibleWitness(),
		other.MyRecoverer(),
		other.SaveSpotRecoverer,
	)
}

type prsr[Output any] struct {
	expected          string
	it                func(State) (State, Output)
	possibleWitness   bool
	recoverer         Recoverer // will be requested only by the SaveSpot parser
	saveSpotRecoverer Recoverer // will be requested in escape mode to find the best SaveSpot parser
}

// NewParser is THE way to create parsers.
func NewParser[Output any](
	expected string,
	parse func(State) (State, Output),
	possibleWitness bool,
	recover Recoverer,
	saveSpotRecoverer Recoverer,
) Parser[Output] {
	p := prsr[Output]{
		expected:          expected,
		it:                parse,
		possibleWitness:   possibleWitness,
		recoverer:         recover,
		saveSpotRecoverer: saveSpotRecoverer,
	}
	if recover == nil {
		p.recoverer = DefaultRecoverer(p)
	}
	return p
}

func (p prsr[Output]) Expected() string {
	return p.expected
}

func (p prsr[Output]) It(state State) (State, Output) {
	return p.it(state)
}

func (p prsr[Output]) PossibleWitness() bool {
	return p.possibleWitness
}

func (p prsr[Output]) MyRecoverer() Recoverer {
	return p.recoverer
}

func (p prsr[Output]) SwapMyRecoverer(newRecoverer Recoverer) Parser[Output] {
	return prsr[Output]{ // make it concurrency safe without locking
		expected:          p.expected,
		it:                p.it,
		recoverer:         newRecoverer,
		saveSpotRecoverer: p.saveSpotRecoverer,
	}
}

func (p prsr[Output]) SaveSpotRecoverer(state State) int {
	if p.saveSpotRecoverer == nil {
		return -1
	}
	return p.saveSpotRecoverer(state)
}

type lazyprsr[Output any] struct {
	once         sync.Once
	makePrsr     func() Parser[Output]
	cachedPrsr   Parser[Output]
	newRecoverer Recoverer
}

// LazyParser just stores a function that creates the parser and evaluates the function later.
// This allows to defer the call to NewParser() and thus to define recursive grammars.
func LazyParser[Output any](makeParser func() Parser[Output]) Parser[Output] {
	return &lazyprsr[Output]{makePrsr: makeParser}
}

func (lp *lazyprsr[Output]) ensurePrsr() {
	lp.cachedPrsr = lp.makePrsr()
	if lp.newRecoverer != nil {
		lp.cachedPrsr = lp.cachedPrsr.SwapMyRecoverer(lp.newRecoverer)
		lp.newRecoverer = nil
	}
}

func (lp *lazyprsr[Output]) Expected() string {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Expected()
}

func (lp *lazyprsr[Output]) It(state State) (State, Output) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.It(state)
}

func (lp *lazyprsr[Output]) PossibleWitness() bool {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.PossibleWitness()
}

func (lp *lazyprsr[Output]) MyRecoverer() Recoverer {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.MyRecoverer()
}

func (lp *lazyprsr[Output]) SwapMyRecoverer(newRecoverer Recoverer) Parser[Output] {
	if lp.cachedPrsr == nil {
		return &lazyprsr[Output]{ // return a new instance that can't be in a stale cache somewhere
			once:         sync.Once{},
			makePrsr:     lp.makePrsr,
			cachedPrsr:   lp.cachedPrsr,
			newRecoverer: newRecoverer,
		}
	}
	return lp.cachedPrsr.SwapMyRecoverer(newRecoverer)
}

func (lp *lazyprsr[Output]) SaveSpotRecoverer(state State) int {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.SaveSpotRecoverer(state)
}

// ============================================================================
// Running a parser
//

// RunOnString runs a parser on text input and returns the output and error(s).
// It uses default values for maximum number of "tokens" to delete for error handling,
// the number of recoverers to try and the deleter to use.
func RunOnString[Output any](input string, parse Parser[Output]) (Output, error) {
	newState, output := RunOnState(NewFromString(-1, nil, -1, input), parse)
	if err := newState.Errors(); err != nil {
		return ZeroOf[Output](), err
	}
	return output, nil
}

// RunOnBytes runs a parser on binary input and returns the output and error(s).
// It uses default values for maximum number of "tokens" to delete for error handling,
// the number of recoverers to try and the deleter to use.
// This is useful for binary or mixed binary/text parsers.
func RunOnBytes[Output any](input []byte, parse Parser[Output]) (Output, error) {
	newState, output := RunOnState(NewFromBytes(-1, nil, -1, input), parse)
	if err := newState.Errors(); err != nil {
		return ZeroOf[Output](), err
	}
	return output, nil
}

func RunOnState[Output any](state State, parse Parser[Output]) (State, Output) {
	var output Output

	id := NewBranchParserID()
	newState := state

	for {
		switch newState.ParsingMode() {
		case ParsingModeHappy: // normal parsing
			newState, output = parse.It(state)
		case ParsingModeError: // find previous SaveSpot (backward)
			state = IWitnessed(state, id, 0, newState)
			if state.ParsingMode() == ParsingModeError {
				state.mode = ParsingModeHandle
				if newState.errHand.err != nil {
					state.oldErrors = append(state.oldErrors, *state.errHand.err)
					state.errHand.err = nil
				}
			}
			Debugf("RunOnState - error -> %s: curDel=%d, ignoreErrParser=%t", state.mode, state.errHand.curDel, state.errHand.ignoreErrParser)
			newState, output = parse.It(state)
		case ParsingModeHandle: // find error again (forward)
			state = state.Preserve(newState)
			newState, output = HandleWitness(state, id, 0, parse)
		case ParsingModeRewind: // go back to error / witness parser (1) (backward)
			state = state.Preserve(newState)
			newState, output = HandleWitness(state, id, 0, parse)
		case ParsingModeEscape: // escape the mess the hard way: use recoverer (forward)
			newState, output = parse.It(state.Preserve(newState))
		}
		if newState.mode == ParsingModeHappy {
			return newState, output
		}
		if newState.mode == ParsingModeEscape && newState.AtEnd() { // stop riding a dead horse
			return newState, output
		}
		Debugf("RunOnState - %s: curDel=%d, ignoreErrParser=%t", state.mode, state.errHand.curDel, state.errHand.ignoreErrParser)
	}
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
func NewFromString(maxDel int, del Deleter, maxRecover int, input string) State {
	if del == nil {
		del = DefaultTextDeleter
	}
	return newState(maxDel, del, maxRecover, false, nil, input)
}

// NewFromBytes creates a new parser state from the input data.
func NewFromBytes(maxDel int, del Deleter, maxRecover int, input []byte) State {
	if del == nil {
		del = DefaultTextDeleter
	}
	return newState(maxDel, del, maxRecover, true, input, "")
}

// newState creates a new parser state from the input data.
func newState(maxDel int, del Deleter, maxRecover int, binary bool, bytes []byte, text string) State {
	if maxDel < 0 {
		maxDel = DefaultMaxDel
	}
	if maxRecover < 0 {
		maxRecover = DefaultMaxRecover
	}
	return State{
		input:                  newInput(binary, bytes, text),
		saveSpot:               -1,
		maxDel:                 maxDel,
		deleter:                del,
		recovererWasteCache:    make(map[uint64][]cachedWaste),
		recovererWasteIdxCache: make(map[uint64][]cachedWasteIdx),
		parserCache:            make(map[uint64][]ParserResult),
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

// IndexOrMinFunc returns the index of the matching value in x,
// using cmp to compare elements.
// It will return the index of the minimal value in x, if no match was found.
// It returns -1 if x is empty. If there is more than one minimal or
// matching element according to the cmp function, it returns the first one.
func IndexOrMinFunc[S ~[]E, E any](x S, match E, cmp func(a, b E) int) int {
	switch len(x) {
	case 0:
		return -1
	case 1:
		return 0
	}
	m := x[0]
	idx := 0
	for i := 1; i < len(x); i++ {
		v := x[i]
		if cmp(v, match) == 0 {
			return i
		}
		if cmp(v, m) < 0 {
			m = v
			idx = i
		}
	}
	return idx
}
