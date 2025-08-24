package comb

import (
	"math"
	"sync"
)

const (
	ParentUndefined = math.MinInt32 + iota // used for calling the root parser
	ParentUnknown                          // used for bottom-up parsing
)

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
func (pids *ParserIDs) setParent(id int32) {
	pids.parent = id
}

// ============================================================================
// Leaf Parser
//

type prsr[Output any] struct {
	ParserIDs
	expected      string
	parseWithData func(State, interface{}) (State, Output, *ParserError, interface{})
	recoverer     Recoverer
	saveSpot      bool
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
		ParserIDs: ParserIDs{id: -1, parent: ParentUndefined},
		expected:  expected,
		parseWithData: func(state State, data interface{}) (State, Output, *ParserError, interface{}) {
			nState, out, err := parse(state)
			return nState, out, err, nil
		},
		recoverer: recover,
	}
	return p
}

// NewParserWithData is the way to create leaf parsers that have partial results
// they want to save in case of an error.
// recover can be nil to signal that there is no optimized recoverer available.
// In case of an error, the parser will be called again and again moving forward
// one byte/rune at a time instead.
func NewParserWithData[Output any](
	expected string,
	parse func(State, interface{}) (State, Output, *ParserError, interface{}),
	recover Recoverer,
) Parser[Output] {
	p := &prsr[Output]{
		ParserIDs:     ParserIDs{id: -1, parent: ParentUndefined},
		expected:      expected,
		parseWithData: parse,
		recoverer:     recover,
	}
	return p
}

func (p *prsr[Output]) Expected() string {
	return p.expected
}
func (p *prsr[Output]) Parse(state State) (State, Output, *ParserError) {
	nState, out, err, data := p.parseWithData(state, nil)
	if err != nil && data != nil {
		err.StoreParserData(p.ID(), data)
	}
	if err != nil && err.parserID < 0 {
		err.parserID = p.ID()
	}
	return nState, out, err
}
func (p *prsr[Output]) ParseAny(parent int32, state State) (State, interface{}, *ParserError) {
	if parent >= 0 {
		p.setParent(parent)
	}
	return p.Parse(state)
}
func (p *prsr[Output]) parseAnyAfterError(err *ParserError, state State) (int32, State, interface{}, *ParserError) {
	nState, out, newErr, data := p.parseWithData(state, err.ParserData(p.ID()))
	if newErr != nil {
		newErr.StoreParserData(p.ID(), data)
	}
	if newErr != nil && newErr.parserID < 0 {
		newErr.parserID = p.ID()
	}
	return p.ParserIDs.parent, nState, out, newErr
}
func (p *prsr[Output]) IsSaveSpot() bool {
	return p.saveSpot
}
func (p *prsr[Output]) setSaveSpot() {
	p.saveSpot = true
}
func (p *prsr[Output]) Recover(state State, data interface{}) (int, interface{}) {
	return p.recoverer(state, data)
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
	saveSpot      bool
	childs        func() []AnyParser
	prsAfterChild func(childID int32, childStartState, childState State, childOut interface{}, childErr *ParserError, data interface{},
	) (State, Output, *ParserError, interface{})
}

// NewBranchParser is THE way to create branch parsers.
// parseAfterChild is called with a `childID < 0` during normal (top -> down) parsing.
// It will be called with a `childID >= 0` during error recovery (bottom -> up).
func NewBranchParser[Output any](
	expected string,
	children func() []AnyParser,
	parseAfterChild func(childID int32, childStartState, childState State, childOut interface{}, childErr *ParserError, data interface{},
	) (State, Output, *ParserError, interface{}),
) Parser[Output] {
	return &brnchprsr[Output]{
		ParserIDs:     ParserIDs{id: -1, parent: ParentUndefined},
		expected:      expected,
		childs:        children,
		prsAfterChild: parseAfterChild,
	}
}
func (bp *brnchprsr[Output]) Expected() string {
	return bp.expected
}
func (bp *brnchprsr[Output]) Parse(state State) (State, Output, *ParserError) {
	nState, aOut, err := bp.ParseAny(ParentUnknown, state)
	out, _ := aOut.(Output)
	return nState, out, err
}
func (bp *brnchprsr[Output]) ParseAny(parentID int32, state State) (State, interface{}, *ParserError) {
	bp.ensureIDs()
	if parentID >= 0 {
		bp.setParent(parentID)
	}
	nState, out, err, data := bp.prsAfterChild(-1, state, state, nil, nil, nil)
	if err != nil && data != nil {
		err.StoreParserData(bp.ID(), data)
	}
	return nState, out, err
}
func (bp *brnchprsr[Output]) parseAfterError(
	err *ParserError, childID int32, childStartState, childState State, childOut interface{}, childErr *ParserError,
) (int32, State, interface{}, *ParserError) {
	bp.ensureIDs()
	nState, out, nErr, data := bp.prsAfterChild(childID, childStartState, childState, childOut, childErr, err.ParserData(bp.ID()))
	if nErr != nil && data != nil {
		nErr.StoreParserData(bp.ID(), data)
	}
	if nErr != nil && nErr.parserID < 0 {
		nErr.parserID = bp.ID()
	}
	return bp.ParserIDs.parent, nState, out, nErr
}
func (bp *brnchprsr[Output]) parseAnyAfterError(_ *ParserError, _ State) (int32, State, interface{}, *ParserError) {
	panic("a branch parser has to be called with `parseAfterError` instead")
}
func (bp *brnchprsr[Output]) IsSaveSpot() bool {
	return bp.saveSpot
}
func (bp *brnchprsr[Output]) setSaveSpot() {
	bp.saveSpot = true
}
func (bp *brnchprsr[Output]) Recover(_ State, _ interface{}) (int, interface{}) {
	return RecoverNever, nil // never recover with a branch parser
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
func (lp *lazyprsr[Output]) Expected() string {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Expected()
}
func (lp *lazyprsr[Output]) Parse(state State) (State, Output, *ParserError) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Parse(state)
}
func (lp *lazyprsr[Output]) ParseAny(parent int32, state State) (State, interface{}, *ParserError) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.ParseAny(parent, state)
}
func (lp *lazyprsr[Output]) parseAfterError(
	err *ParserError, childID int32, childStartState, childState State, childOut interface{}, childErr *ParserError,
) (int32, State, interface{}, *ParserError) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.(BranchParser).parseAfterError(err, childID, childStartState, childState, childOut, childErr)
}
func (lp *lazyprsr[Output]) parseAnyAfterError(_ *ParserError, _ State) (int32, State, interface{}, *ParserError) {
	panic("a branch parser has to be called with `parseAfterError` instead")
}
func (lp *lazyprsr[Output]) IsSaveSpot() bool {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.IsSaveSpot()
}
func (lp *lazyprsr[Output]) setSaveSpot() {
	lp.once.Do(lp.ensurePrsr)
	lp.cachedPrsr.setSaveSpot()
}
func (lp *lazyprsr[Output]) Recover(state State, data interface{}) (int, interface{}) {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.Recover(state, data)
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
func (lp *lazyprsr[Output]) children() []AnyParser {
	lp.once.Do(lp.ensurePrsr)
	return lp.cachedPrsr.(BranchParser).children()
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
	tstState := NewFromBytes([]byte{}, 0)
	if recoverer != nil {
		waste, _ := recoverer(tstState, nil)
		if waste == RecoverNever {
			panic("can't make parser with Forbidden recoverer a safe spot")
		}
	}

	pp, ok := p.(*prsr[Output])
	if !ok {
		panic("SafeSpot can only be applied to leaf parsers")
	}
	pp.setSaveSpot()
	parse := pp.parseWithData
	pp.parseWithData = func(state State, data interface{}) (State, Output, *ParserError, interface{}) {
		nState, out, err, data2 := parse(state, data)
		if err == nil {
			return nState.MoveSafeSpot(), out, nil, data2
		}
		return nState, out, err, data2
	}
	return pp
	/*
		var sp Parser[Output]
		//nParse := func(state State) (State, Output, *ParserError) {
		//	if p.ID() < 0 {
		//		p.setID(sp.ID()) // share the same ID because we will never have any own error data
		//	}
		//	nState, aOut, err := p.ParseAny(sp.ID(), state)
		//	if err == nil {
		//		nState = nState.MoveSafeSpot() // move the mark!
		//	}
		//	out, _ := aOut.(Output)
		//	return nState, out, ClaimError(err)
		//}
		//sp = NewParser[Output](p.Expected(), nParse, p.Recover)
		nParse := func(
			childID int32,
			childStartState, childState State,
			childOut interface{},
			childErr *ParserError,
			data interface{},
		) (State, Output, *ParserError, interface{}) {
			if childID < 0 {
				childStartState = childState
				childState, childOut, childErr = p.ParseAny(sp.ID(), childStartState)
			}
			if childErr == nil {
				childState = childState.MoveSafeSpot() // move the mark!
			}
			out, _ := childOut.(Output)
			return childState, out, childErr, data
		}
		sp = NewBranchParser[Output](
			p.Expected(),
			func() []AnyParser {
				return []AnyParser{p}
			},
			nParse,
		)
		sp.setSaveSpot()
		return sp
	*/
}
