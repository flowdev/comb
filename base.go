// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

import (
	"sync"
)

// Use the stringer package from the Go team for printing of names of enums:
//go:generate go run golang.org/x/tools/cmd/stringer@latest -linecomment -type ParsingMode,Ternary

// DefaultMaxDel of 3 is a compromise between speed and optimal fault tolerance
// (ANTLR is using 1)
const DefaultMaxDel = 3

// ParsingMode is needed for error handling. See `ERROR_HANDLING.md` for details.
type ParsingMode int

const (
	// ParsingModeHappy - normal parsing until failure (forward)
	ParsingModeHappy ParsingMode = iota // happy
	// ParsingModeError - find previous NoWayBack (backward)
	ParsingModeError // error
	// ParsingModeHandle - find witness parser (1) again (forward)
	ParsingModeHandle // handle
	// ParsingModeRewind - find witness parser (1) again (backward)
	ParsingModeRewind // rewind
	// ParsingModeEscape - find the (best) next NoWayBack (forward)
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
// to reach a safe state (NoWayBack).
// If it can't recover it should return -1.
//
// A Recoverer is used for recovering from an error in the input.
// It helps to move forward to the next safe spot (NoWayBack).
// A Recoverer will be used by the NoWayBack parser if it's sub-parser
// provides it.
// Otherwise, NoWayBack will have to try the sub-parser until it succeeds moving
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
	NoWayBackRecoverer(State) int
}

type prsr[Output any] struct {
	expected           string
	it                 func(State) (State, Output)
	possibleWitness    bool
	recoverer          Recoverer // will be requested only by the NoWayBack parser
	noWayBackRecoverer Recoverer // will be requested in escape mode to find the best NoWayBack parser
}

// NewParser is THE way to create parsers.
func NewParser[Output any](
	expected string,
	parse func(State) (State, Output),
	possibleWitness bool,
	recover Recoverer,
	noWayBackRecoverer Recoverer,
) Parser[Output] {
	p := prsr[Output]{
		expected:           expected,
		it:                 parse,
		possibleWitness:    possibleWitness,
		recoverer:          recover,
		noWayBackRecoverer: noWayBackRecoverer,
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
		expected:           p.expected,
		it:                 p.it,
		recoverer:          newRecoverer,
		noWayBackRecoverer: p.noWayBackRecoverer,
	}
}

func (p prsr[Output]) NoWayBackRecoverer(state State) int {
	if p.noWayBackRecoverer == nil {
		return -1
	}
	return p.noWayBackRecoverer(state)
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

func (lp *lazyprsr[Output]) NoWayBackRecoverer(state State) int {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.NoWayBackRecoverer(state)
}

// ============================================================================
// Running a parser
//

// RunOnString runs a parser on text input and returns the output and error(s).
func RunOnString[Output any](maxDel int, del Deleter, input string, parse Parser[Output]) (Output, error) {
	newState, output := RunOnState(NewFromString(maxDel, del, input), parse)
	if len(newState.oldErrors) == 0 {
		return output, nil
	}
	return ZeroOf[Output](), pcbErrorsToGoErrors(newState)
}

// RunOnBytes runs a parser on binary input and returns the output and error(s).
// This is useful for binary or mixed binary/text parsers.
func RunOnBytes[Output any](maxDel int, del Deleter, input []byte, parse Parser[Output]) (Output, error) {
	newState, output := RunOnState(NewState(maxDel, del, input), parse)
	if len(newState.oldErrors) == 0 {
		return output, nil
	}
	return ZeroOf[Output](), pcbErrorsToGoErrors(newState)
}

func RunOnState[Output any](state State, parse Parser[Output]) (State, Output) {
	var output Output

	id := NewBranchParserID()
	newState := state

	for {
		switch newState.ParsingMode() {
		case ParsingModeHappy: // normal parsing
			newState, output = parse.It(state)
		case ParsingModeError: // find previous NoWayBack (backward)
			state = IWitnessed(state, id, 0, newState)
			state.mode = ParsingModeHandle
			if newState.errHand.err != nil {
				state.oldErrors = append(state.oldErrors, *state.errHand.err)
			}
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
	// Go is fundamentally working with bytes and can interpret them as strings or as containing runes.
	// There are no standard library functions for handling []rune or the like.
	bytes  []byte
	pos    int // current position in the sequence a.k.a. the *byte* index
	prevNl int // position of newline preceding 'pos' (-1 for line==1)
	line   int // current line number
}

// NewFromString creates a new parser state from the input data.
func NewFromString(maxDel int, del Deleter, input string) State {
	if del == nil {
		del = DefaultTextDeleter
	}
	state := NewState(maxDel, del, []byte(input))
	return state
}

// NewState creates a new parser state from the input data.
func NewState(maxDel int, del Deleter, input []byte) State {
	if maxDel < 0 {
		maxDel = DefaultMaxDel
	}
	if del == nil {
		del = DefaultBinaryDeleter
	}
	return State{
		input:                  Input{bytes: input, line: 1, prevNl: -1},
		noWayBackMark:          -1,
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
