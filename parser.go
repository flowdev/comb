package comb

import (
	"sync"
)

// ParserID is the base of every comb parser.
// It enables registering of all parsers and error recovery.
type ParserID int32

func (pid *ParserID) ID() int32 {
	return int32(*pid)
}
func (pid *ParserID) setID(id int32) {
	*pid = ParserID(id)
}

// ============================================================================
// Leaf Parser
//

type prsr[Output any] struct {
	ParserID
	expected  string
	parser    func(State) (State, Output, *ParserError)
	recoverer func(State) int
	saveSpot  bool
}

// NewParser is THE way to create leaf parsers.
// recover can be nil to signal that there is no optimized recoverer available.
// In case of an error the parser will be called again and again moving forward
// one byte/rune at a time instead.
func NewParser[Output any](
	expected string,
	parse func(State) (State, Output, *ParserError),
	recover Recoverer,
) Parser[Output] {
	p := &prsr[Output]{
		expected:  expected,
		parser:    parse,
		recoverer: recover,
	}
	p.setID(-1)
	return p
}

func (p *prsr[Output]) Expected() string {
	return p.expected
}
func (p *prsr[Output]) Parse(state State) (State, Output, *ParserError) {
	nState, out, err := p.parser(state)
	if err != nil && err.parserID < 0 {
		err.parserID = p.ID()
	}
	return nState, out, err
}
func (p *prsr[Output]) parse(state State) ParseResult {
	nState, output, err := p.Parse(state)
	return ParseResult{StartState: state, EndState: nState, Output: output, Error: err}
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

// ============================================================================
// Branch Parser
//

type brnchprsr[Output any] struct {
	ParserID
	expected      string
	childs        func() []AnyParser
	prsAfterChild func(childID int32, childResult ParseResult) ParseResult
}

// NewBranchParser is THE way to create branch parsers.
// parseAfterChild will be called with a childID < 0 if it should parse from
// the beginning.
func NewBranchParser[Output any](
	expected string,
	children func() []AnyParser,
	parseAfterChild func(childID int32, childResult ParseResult) ParseResult,
) Parser[Output] {
	return &brnchprsr[Output]{
		ParserID:      ParserID(-1),
		expected:      expected,
		childs:        children,
		prsAfterChild: parseAfterChild,
	}
}
func (bp *brnchprsr[Output]) Expected() string {
	return bp.expected
}
func (bp *brnchprsr[Output]) Parse(state State) (State, Output, *ParserError) {
	result := bp.parseAfterChild(-1, ParseResult{EndState: state})
	if out, ok := result.Output.(Output); ok {
		return result.EndState, out, result.Error
	}
	return result.EndState, ZeroOf[Output](), result.Error
}
func (bp *brnchprsr[Output]) parse(state State) ParseResult {
	return bp.parseAfterChild(-1, ParseResult{EndState: state})
}
func (bp *brnchprsr[Output]) IsSaveSpot() bool {
	return false
}
func (bp *brnchprsr[Output]) setSaveSpot() {
	panic("a branch parser can never be a save spot")
}
func (bp *brnchprsr[Output]) Recover(_ State) int {
	return RecoverWasteNever // never recover with a branch parser
}
func (bp *brnchprsr[Output]) IsStepRecoverer() bool {
	return false
}
func (bp *brnchprsr[Output]) SwapRecoverer(_ Recoverer) {
	panic("a branch parser can never have a special recoverer")
}
func (bp *brnchprsr[Output]) children() []AnyParser {
	return bp.childs()
}
func (bp *brnchprsr[Output]) parseAfterChild(childID int32, childResult ParseResult) ParseResult {
	bp.ensureIDs()
	childResult = childResult.prepareOutputFor(bp.ID())
	result := bp.prsAfterChild(childID, childResult)
	result.setID(bp.ID())
	if result.Error != nil && result.Error.parserID < 0 {
		result.Error.parserID = bp.ID()
	}
	return result
}
func (bp *brnchprsr[Output]) ensureIDs() { // only needed if Parse was called directly
	if bp.ID() < 0 { // ensure sane IDs
		bp.setID(0)
		for i, child := range bp.childs() {
			child.setID(int32(i + 1))
		}
	}
}

// ============================================================================
// Lazy Branch Parser
//

type lazyprsr[Output any] struct {
	cachedPrsr   Parser[Output]
	once         sync.Once
	makePrsr     func() Parser[Output]
	newRecoverer Recoverer
}

// LazyBranchParser just stores a function that creates the parser and evaluates the function later.
// This allows to defer the call to NewParser() and thus to define recursive grammars.
// Only branch parsers need this ability. A leaf parser can't be recursive by definition.
func LazyBranchParser[Output any](makeParser func() Parser[Output]) Parser[Output] {
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
func (lp *lazyprsr[Output]) Parse(state State) (State, Output, *ParserError) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Parse(state)
}
func (lp *lazyprsr[Output]) parse(state State) ParseResult {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.parse(state)
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

// ============================================================================
// Save Spot Parser
//

// SafeSpot applies a sub-parser and marks the new state as a
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
// SafeSpot is THE cornerstone of good and performant parsing otherwise.
//
// NOTE:
//   - Parsers that accept the empty input or only perform look ahead are
//     NOT allowed as sub-parsers.
//     SafeSpot tests the optional recoverer of the parser during the
//     construction phase to do a timely panic.
//     This way we won't have to panic at the runtime of the parser.
//   - Only leaf parsers MUST be given to SafeSpot as sub-parsers.
//     SafeSpot will treat the sub-parser as a leaf parser.
//     Any error will look as if coming from SafeSpot itself.
func SafeSpot[Output any](p Parser[Output]) Parser[Output] {
	// call Recoverer to find a Forbidden recoverer during the construction phase and panic
	recoverer := p.Recover
	if recoverer != nil && recoverer(NewFromBytes([]byte{}, true, 0)) == RecoverWasteNever {
		panic("can't make parser with Forbidden recoverer a safe spot")
	}

	if _, ok := p.(BranchParser); ok {
		panic("a branch parser can never be a save spot")
	}

	nParse := func(state State) (State, Output, *ParserError) {
		nState, output, err := p.Parse(state)
		if err == nil {
			nState = nState.MoveSafeSpot() // move the mark!
		}
		return nState, output, ClaimError(err)
	}
	sp := NewParser[Output](p.Expected(), nParse, p.Recover)
	sp.setSaveSpot()
	return sp
}
