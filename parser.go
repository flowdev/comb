package gomme

import "sync"

type prsr[Output any] struct {
	id        int32
	expected  string
	parser    func(State) (State, Output, *ParserError)
	recoverer func(State) int
	saveSpot  bool
}

// NewParser is THE way to create parsers.
func NewParser[Output any](
	expected string,
	parse func(State) (State, Output, *ParserError),
	recover Recoverer,
) Parser[Output] {
	p := &prsr[Output]{
		id:        -1,
		expected:  expected,
		parser:    parse,
		recoverer: recover,
	}
	return p
}

func (p *prsr[Output]) ID() int32 {
	return p.id
}
func (p *prsr[Output]) Expected() string {
	return p.expected
}
func (p *prsr[Output]) It(state State) (State, Output, *ParserError) {
	nState, out, err := p.parser(state)
	if err != nil {
		err.parserID = p.id
	}
	return nState, out, err
}
func (p *prsr[Output]) Parse(state State) ParseResult {
	nState, out, err := p.parser(state)
	if err != nil {
		err.parserID = p.id
	}
	return ParseResult{
		ID:       p.id,
		StartPos: state.CurrentPos(),
		State:    nState,
		Output:   out,
		Error:    err,
	}
}
func (p *prsr[Output]) IsSaveSpot() bool {
	return p.saveSpot
}
func (p *prsr[Output]) setSaveSpot() {
	p.saveSpot = true
}
func (p *prsr[Output]) Recover(state State) int {
	return p.recoverer(state)
}
func (p *prsr[Output]) IsStepRecoverer() bool {
	return p.recoverer == nil
}
func (p *prsr[Output]) SwapRecoverer(newRecoverer Recoverer) {
	p.recoverer = newRecoverer // this isn't concurrency safe, but it only happens in the initialization phase
}
func (p *prsr[Output]) setID(id int32) {
	p.id = id
}

type lazyprsr[Output any] struct {
	cachedPrsr   Parser[Output]
	once         sync.Once
	makePrsr     func() Parser[Output]
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
		lp.cachedPrsr.SwapRecoverer(lp.newRecoverer)
	}
}

func (lp *lazyprsr[Output]) ID() int32 {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.ID()
}
func (lp *lazyprsr[Output]) Expected() string {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Expected()
}
func (lp *lazyprsr[Output]) It(state State) (State, Output, *ParserError) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.It(state)
}
func (lp *lazyprsr[Output]) Parse(state State) ParseResult {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Parse(state)
}
func (lp *lazyprsr[Output]) IsSaveSpot() bool {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.IsSaveSpot()
}
func (lp *lazyprsr[Output]) setSaveSpot() {
	lp.once.Do(lp.ensurePrsr)
	lp.cachedPrsr.setSaveSpot()
}
func (lp *lazyprsr[Output]) Recover(state State) int {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Recover(state)
}
func (lp *lazyprsr[Output]) IsStepRecoverer() bool {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.IsStepRecoverer()
}
func (lp *lazyprsr[Output]) SwapRecoverer(newRecoverer Recoverer) {
	if lp.cachedPrsr == nil {
		lp.newRecoverer = newRecoverer
		return
	}
	lp.cachedPrsr.SwapRecoverer(newRecoverer)
}
func (lp *lazyprsr[Output]) setID(id int32) {
	lp.once.Do(lp.ensurePrsr)
	lp.cachedPrsr.setID(id)
}

// BaseBranchParser is the base of all branch parsers.
// It reduces the implementation of a branch parser to these methods:
//
//	Expected() string
//	Children() []AnyParser
//	ParseAfterChild(childResult ParseResult) ParseResult
//
// Expected and Children should be trivial and ParseAfterChild is the real implementation.
// ParseAfterChild will be called with a child result that has a childID < 0
// if it should parse from the beginning.
// In case of an error, ParseAfterChild has to return a zero value for Output (not nil).
type BaseBranchParser[Output any] struct {
	id              int32
	children        func() []AnyParser
	parseAfterChild func(childResult ParseResult) ParseResult
}

func NewBaseBranchParser[Output any](
	children func() []AnyParser,
	parseAfterChild func(childResult ParseResult) ParseResult,
) BaseBranchParser[Output] {
	return BaseBranchParser[Output]{id: -1, children: children, parseAfterChild: parseAfterChild}
}
func (bbp *BaseBranchParser[Output]) ID() int32 {
	return bbp.id
}
func (bbp *BaseBranchParser[Output]) It(state State) (State, Output, *ParserError) {
	r := bbp.Parse(state)
	out, ok := r.Output.(Output)
	if !ok {
		return r.State, ZeroOf[Output](), r.Error
	}
	return r.State, out, r.Error
}
func (bbp *BaseBranchParser[Output]) Parse(state State) ParseResult {
	return bbp.parseAfterChild(ParseResult{ID: -1, State: state})
}
func (bbp *BaseBranchParser[Output]) IsSaveSpot() bool {
	return false
}
func (bbp *BaseBranchParser[Output]) setSaveSpot() {
	panic("a branch parser can never be a save spot")
}
func (bbp *BaseBranchParser[Output]) Recover(_ State) int {
	panic("must not use a branch parser for recovering from an error")
}
func (bbp *BaseBranchParser[Output]) IsStepRecoverer() bool {
	return true
}
func (bbp *BaseBranchParser[Output]) SwapRecoverer(_ Recoverer) {
	panic("a branch parser can never have a special recoverer")
}
func (bbp *BaseBranchParser[Output]) setID(id int32) {
	bbp.id = id
}

// SaveSpot applies a sub-parser and marks the new state as a
// point of no return if successful.
// It really serves 3 slightly different purposes:
//
//  1. Prevent a `FirstSuccessful` parser from trying later sub-parsers even
//     in case of an error.
//  2. Prevent other unnecessary backtracking in case of an error.
//  3. Mark a parser as a potential safe place to recover to
//     when recovering from an error.
//
// So you don't need this parser at all if your input is always correct.
// SaveSpot is THE cornerstone of good and performant parsing otherwise.
//
// Note:
//   - Parsers that accept the empty input or only perform look ahead are
//     NOT allowed as sub-parsers.
//     SaveSpot tests the optional recoverer of the parser during the
//     construction phase to provoke an early panic.
//     This way we won't have a panic at the runtime of the parser.
//   - Only leaf parsers MUST be given to SaveSpot as sub-parsers.
//     SaveSpot will treat the sub-parser as a leaf parser.
func SaveSpot[Output any](parse Parser[Output]) Parser[Output] {
	// call Recoverer to make a Forbidden recoverer panic during the construction phase
	recoverer := parse.Recover
	if recoverer != nil {
		recoverer(NewFromBytes([]byte{}, true))
	}

	if _, ok := parse.(BranchParser); ok {
		panic("a branch parser can never be a save spot")
	}

	nParse := func(state State) (State, Output, *ParserError) {
		nState, output, err := parse.It(state)
		if err == nil {
			nState.saveSpot = nState.input.pos // move the mark!
		}
		return nState, output, err
	}
	sp := NewParser[Output](parse.Expected(), nParse, parse.Recover)
	sp.setSaveSpot()
	return sp
}
