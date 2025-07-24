package comb

import (
	"math"
	"sync"
)

const ParentUnknown = math.MinInt32

// ParserIDs is the base of every comb parser.
// It enables registering of all parsers and error recovery.
type ParserIDs struct {
	id, parent int32
}

func (pids *ParserIDs) ID() int32 {
	return pids.id
}
func (pids *ParserIDs) setID(id int32) {
	pids.id = id
}
func (pids *ParserIDs) LastParent() int32 {
	return pids.parent
}
func (pids *ParserIDs) setParent(id int32) {
	pids.parent = id
}

// ============================================================================
// Leaf Parser
//

type prsr[Output any] struct {
	ParserIDs
	expected  string
	parser    func(State) (State, Output, *ParserError)
	recoverer Recoverer
	saveSpot  bool
}

// NewParser is THE way to create simple leaf parsers.
// recover can be nil to signal that there is no optimized recoverer available.
// In case of an error, the parser will be called again and again moving forward
// one byte/rune at a time instead.
func NewParser[Output any](
	expected string,
	parse func(State) (State, Output, *ParserError),
	recover Recoverer,
) Parser[Output] {
	p := &prsr[Output]{
		ParserIDs: ParserIDs{id: -1, parent: ParentUnknown},
		expected:  expected,
		parser:    parse,
		recoverer: recover,
	}
	return p
}

func (p *prsr[Output]) Expected() string {
	return p.expected
}
func (p *prsr[Output]) Parse(parent int32, state State) (State, Output, *ParserError) {
	if parent != ParentUnknown {
		p.setParent(parent)
	}
	nState, out, err := p.parser(state)
	if err != nil && err.parserID < 0 {
		err.parserID = p.ID()
	}
	return nState, out, err
}
func (p *prsr[Output]) parse(parent int32, state State) ParseResult {
	nState, output, err := p.Parse(parent, state)
	return ParseResult{StartState: state, EndState: nState, Output: output, Error: err}
}
func (p *prsr[Output]) IsSaveSpot() bool {
	return p.saveSpot
}
func (p *prsr[Output]) setSaveSpot() {
	p.saveSpot = true
}
func (p *prsr[Output]) Recover(pe *ParserError, state State) int {
	return p.recoverer(pe, state)
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
	ParserIDs
	expected      string
	childs        func() []AnyParser
	prsAfterError func(pe *ParserError, childID int32, childResult ParseResult) ParseResult
}

// NewBranchParser is THE way to create branch parsers.
// parseAfterError will be called with a nil error and childID < 0
// if it should parse from the beginning.
func NewBranchParser[Output any](
	expected string,
	children func() []AnyParser,
	parseAfterError func(pe *ParserError, childID int32, childResult ParseResult) ParseResult,
) Parser[Output] {
	return &brnchprsr[Output]{
		ParserIDs:     ParserIDs{id: -1, parent: -1},
		expected:      expected,
		childs:        children,
		prsAfterError: parseAfterError,
	}
}
func (bp *brnchprsr[Output]) Expected() string {
	return bp.expected
}
func (bp *brnchprsr[Output]) Parse(parent int32, state State) (State, Output, *ParserError) {
	result := bp.parseAfterError(nil, -1, parent, ParseResult{EndState: state})
	if out, ok := result.Output.(Output); ok {
		return result.EndState, out, result.Error
	}
	return result.EndState, ZeroOf[Output](), result.Error
}
func (bp *brnchprsr[Output]) parse(parent int32, state State) ParseResult {
	return bp.parseAfterError(nil, -1, parent, ParseResult{EndState: state})
}
func (bp *brnchprsr[Output]) IsSaveSpot() bool {
	return false
}
func (bp *brnchprsr[Output]) setSaveSpot() {
	panic("a branch parser can never be a save spot")
}
func (bp *brnchprsr[Output]) Recover(_ *ParserError, _ State) int {
	return RecoverNever // never recover with a branch parser
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
func (bp *brnchprsr[Output]) parseAfterError(pe *ParserError, childID, parentID int32, childResult ParseResult) ParseResult {
	bp.ensureIDs()
	if parentID != ParentUnknown {
		bp.setParent(parentID)
	}
	return bp.prsAfterError(pe, childID, childResult)
}
func (bp *brnchprsr[Output]) ensureIDs() { // only needed if Parse was called directly
	if bp.ID() < 0 { // ensure sane IDs
		bp.id = 0
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
// This allows deferring the call to NewParser() and thus to define recursive grammars.
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
func (lp *lazyprsr[Output]) LastParent() int32 {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.LastParent()
}
func (lp *lazyprsr[Output]) Expected() string {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Expected()
}
func (lp *lazyprsr[Output]) Parse(parent int32, state State) (State, Output, *ParserError) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Parse(parent, state)
}
func (lp *lazyprsr[Output]) parse(parent int32, state State) ParseResult {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.parse(parent, state)
}
func (lp *lazyprsr[Output]) IsSaveSpot() bool {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.IsSaveSpot()
}
func (lp *lazyprsr[Output]) setSaveSpot() {
	lp.once.Do(lp.ensurePrsr)
	lp.cachedPrsr.setSaveSpot()
}
func (lp *lazyprsr[Output]) Recover(pe *ParserError, state State) int {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Recover(pe, state)
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
func (lp *lazyprsr[Output]) setParent(id int32) {
	lp.once.Do(lp.ensurePrsr)
	lp.cachedPrsr.setParent(id)
}

// ============================================================================
// Dynamic Parser
//

type DynamicParser[Output any] interface {
	ID() int32
	LastParent() int32
	Expected() string
	Parse(parent int32, state State) (State, Output, *ParserError)
	ParseAfterError(err *ParserError, childID, parentID int32, state State) (State, Output, *ParserError)
	IsSaveSpot() bool
	Recover(*ParserError, State) int
	IsStepRecoverer() bool
	SwapRecoverer(Recoverer)
	SetID(int32)     // sets own ID
	SetParent(int32) // sets ID of current parent
}

type dynprsr[Output any] struct {
	parser DynamicParser[Output]
}

// NewDynamicParser just stores a DynamicParser and delegates every call to it.
// This allows the implementation of very flexible parsers.
func NewDynamicParser[Output any](parser DynamicParser[Output]) Parser[Output] {
	return &dynprsr[Output]{parser: parser}
}

func (dp *dynprsr[Output]) ID() int32 {
	return dp.parser.ID()
}
func (dp *dynprsr[Output]) LastParent() int32 {
	return dp.parser.LastParent()
}
func (dp *dynprsr[Output]) Expected() string {
	return dp.parser.Expected()
}
func (dp *dynprsr[Output]) Parse(parent int32, state State) (State, Output, *ParserError) {
	return dp.parser.Parse(parent, state)
}
func (dp *dynprsr[Output]) parse(parent int32, state State) ParseResult {
	nState, output, err := dp.parser.Parse(parent, state)
	result := ParseResult{
		StartState: state,
		EndState:   nState,
		Output:     output,
		Error:      err,
	}
	if err != nil {
		result.EndState = state
	}
	return result
}
func (dp *dynprsr[Output]) IsSaveSpot() bool {
	return dp.parser.IsSaveSpot()
}
func (dp *dynprsr[Output]) setSaveSpot() {
	panic("a dynamic parser handles save spots itself")
}
func (dp *dynprsr[Output]) parseAfterError(pe *ParserError, childID, parentID int32, childResult ParseResult) ParseResult {
	nState, output, err := dp.parser.ParseAfterError(pe, childID, parentID, childResult.EndState)
	childResult.StartState = childResult.EndState
	childResult.Output = output
	childResult.Error = err
	if err == nil {
		childResult.EndState = nState
	}
	return childResult
}

func (dp *dynprsr[Output]) Recover(pe *ParserError, state State) int {
	return dp.parser.Recover(pe, state)
}
func (dp *dynprsr[Output]) IsStepRecoverer() bool {
	return dp.parser.IsStepRecoverer()
}
func (dp *dynprsr[Output]) SwapRecoverer(newRecoverer Recoverer) {
	dp.parser.SwapRecoverer(newRecoverer)
}
func (dp *dynprsr[Output]) setID(id int32) {
	dp.parser.SetID(id)
}
func (dp *dynprsr[Output]) setParent(id int32) {
	dp.parser.SetParent(id)
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
	var sp Parser[Output]

	// call Recoverer to find a Forbidden recoverer during the construction phase and panic
	recoverer := p.Recover
	tstState := NewFromBytes([]byte{}, 0)
	if recoverer != nil && recoverer(tstState.NewSyntaxError(1, "just a test"), tstState) == RecoverNever {
		panic("can't make parser with Forbidden recoverer a safe spot")
	}

	if _, ok := p.(BranchParser); ok {
		panic("a branch parser can never be a save spot")
	}

	nParse := func(state State) (State, Output, *ParserError) {
		nState, output, err := p.Parse(sp.ID(), state)
		if err == nil {
			nState = nState.MoveSafeSpot() // move the mark!
		}
		return nState, output, ClaimError(err, sp.ID())
	}
	sp = NewParser[Output](p.Expected(), nParse, p.Recover)
	sp.setSaveSpot()
	return sp
}
