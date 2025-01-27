package gomme

import "sync"

// ============================================================================
// Leaf Parser
//

type prsr[Output any] struct {
	id        int32
	expected  string
	parser    func(State) (State, Output, *ParserError)
	recoverer func(State) int
	saveSpot  bool
}

// NewParser is THE way to create leaf parsers.
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
func (p *prsr[Output]) Parse(state State) (int32, State, Output, *ParserError) {
	nState, out, err := p.parser(state)
	if err != nil {
		err.parserID = p.id
	}
	return p.id, nState, out, err
}
func (p *prsr[Output]) parse(state State) ParseResult {
	id, nState, output, err := p.Parse(state)
	return ParseResult{
		ID:     id,
		State:  nState,
		Output: output,
		Error:  err,
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

// ============================================================================
// Branch Parser
//

type OutputBranchParser[Output any] interface {
	Parser[Output]
	BranchParser
}

type brnchprsr[Output any] struct {
	id            int32
	name          string
	childs        func() []AnyParser
	prsAfterChild func(childResult ParseResult) ParseResult
}

// NewBranchParser is THE way to create branch parsers.
// parseAfterChild will be called with a child result that has a childID < 0
// if it should parse from the beginning.
func NewBranchParser[Output any](
	name string,
	children func() []AnyParser,
	parseAfterChild func(childResult ParseResult) ParseResult,
) OutputBranchParser[Output] {
	return &brnchprsr[Output]{
		id:            -1,
		name:          name,
		childs:        children,
		prsAfterChild: parseAfterChild,
	}
}
func (bp *brnchprsr[Output]) ID() int32 {
	return bp.id
}
func (bp *brnchprsr[Output]) Expected() string {
	return bp.name
}
func (bp *brnchprsr[Output]) Parse(state State) (int32, State, Output, *ParserError) {
	result := bp.parseAfterChild(ParseResult{ID: -1, State: state})
	if out, ok := result.Output.(Output); ok {
		return result.ID, result.State, out, result.Error
	}
	return result.ID, result.State, ZeroOf[Output](), result.Error
}
func (bp *brnchprsr[Output]) parse(state State) ParseResult {
	return bp.parseAfterChild(ParseResult{ID: -1, State: state})
}
func (bp *brnchprsr[Output]) IsSaveSpot() bool {
	return false
}
func (bp *brnchprsr[Output]) setSaveSpot() {
	panic("a branch parser can never be a save spot")
}
func (bp *brnchprsr[Output]) Recover(_ State) int {
	panic("must not use a branch parser for recovering from an error")
}
func (bp *brnchprsr[Output]) IsStepRecoverer() bool {
	return true
}
func (bp *brnchprsr[Output]) SwapRecoverer(_ Recoverer) {
	panic("a branch parser can never have a special recoverer")
}
func (bp *brnchprsr[Output]) children() []AnyParser {
	return bp.childs()
}
func (bp *brnchprsr[Output]) parseAfterChild(childResult ParseResult) ParseResult {
	result := bp.prsAfterChild(childResult)
	if result.Error != nil && result.Error.ParserID() < 0 {
		result.Error.parserID = bp.id
	}
	if result.ID < 0 {
		result.ID = bp.id
	}
	return result
}
func (bp *brnchprsr[Output]) setID(id int32) {
	bp.id = id
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
func (lp *lazyprsr[Output]) Parse(state State) (int32, State, Output, *ParserError) {
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
// Note:
//   - Parsers that accept the empty input or only perform look ahead are
//     NOT allowed as sub-parsers.
//     SafeSpot tests the optional recoverer of the parser during the
//     construction phase to provoke an early panic.
//     This way we won't have a panic at the runtime of the parser.
//   - Only leaf parsers MUST be given to SafeSpot as sub-parsers.
//     SafeSpot will treat the sub-parser as a leaf parser.
//     SafeSpot will panic if the output of the sub-parser isn't of the right type.
func SafeSpot[Output any](p Parser[Output]) Parser[Output] {
	// call Recoverer to make a Forbidden recoverer panic during the construction phase
	recoverer := p.Recover
	if recoverer != nil {
		recoverer(NewFromBytes([]byte{}, true))
	}

	if _, ok := p.(BranchParser); ok {
		panic("a branch parser can never be a save spot")
	}

	nParse := func(state State) (State, Output, *ParserError) {
		_, nState, output, err := p.Parse(state)
		if err == nil {
			nState.saveSpot = nState.input.pos // move the mark!
		}
		return nState, output, err
	}
	sp := NewParser[Output](p.Expected(), nParse, p.Recover)
	sp.setSaveSpot()
	return sp
}
