package comb

import (
	"math"
	"slices"
)

// ============================================================================
// Interfaces And Function For Parser Preparation
//

// AnyParser is an internal interface used by PreparedParser.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
type AnyParser interface {
	ID() int32
	ParseAny(parentID int32, state State) (State, interface{}, *ParserError) // top -> down
	parseAnyAfterError(err *ParserError, state State,
	) (lastParentID int32, newState State, output interface{}, newErr *ParserError) // used by parseAll (bottom -> up)
	IsSaveSpot() bool
	Recover(State, interface{}) (int, interface{})
	IsStepRecoverer() bool
	setID(int32)     // only sets own ID
	setParent(int32) // sets initial parent ID
}

// BranchParser is a more internal interface used by orchestrators.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
// BranchParser just adds 2 methods to the Parser and AnyParser interfaces.
type BranchParser interface {
	children() []AnyParser
	parseAfterError(err *ParserError, childID int32, childStartState, childState State, childOut interface{}, childErr *ParserError,
	) (lastParentID int32, newState State, output interface{}, newErr *ParserError) // bottom -> up
}

// ============================================================================
// PreparedParser: Data Structures And Construction
//

type PreparedParser[Output any] struct {
	parsers        []AnyParser
	recoverers     []AnyParser
	stepRecoverers []AnyParser
}

// NewPreparedParser prepares a parser for error recovery.
// Call this directly if you have a parser that you want to run on many inputs.
// You can use this together with RunOnState.
func NewPreparedParser[Output any](p Parser[Output]) *PreparedParser[Output] {
	pp := &PreparedParser[Output]{
		parsers:        make([]AnyParser, 0, 64),
		recoverers:     make([]AnyParser, 0, 64),
		stepRecoverers: make([]AnyParser, 0, 64),
	}
	pp.registerParsers(p, -1)
	return pp
}

func (pp *PreparedParser[Output]) registerParsers(ap AnyParser, parentID int32) {
	if ap.ID() >= 0 {
		Debugf("registerParsers - parser (ID: %d) is already registered with parent %d", ap.ID(), parentID)
		return
	}
	id := int32(len(pp.parsers))
	ap.setID(id)
	ap.setParent(parentID)
	pp.parsers = append(pp.parsers, ap)

	if bp, ok := ap.(BranchParser); ok {
		for _, cp := range bp.children() {
			pp.registerParsers(cp, id)
		}
	} else if ap.IsSaveSpot() {
		if ap.IsStepRecoverer() {
			pp.stepRecoverers = append(pp.stepRecoverers, ap)
		} else {
			pp.recoverers = append(pp.recoverers, ap)
		}
	}
}

// ============================================================================
// PreparedParser: parseAll
//

func (pp *PreparedParser[Output]) parseAll(state State) (Output, error) {
	var id int32 = 0 // this is always the root parser
	recoverCache := slices.Repeat([]int{RecoverWasteUnknown}, len(pp.parsers))
	p := pp.parsers[id]

	// TOP->DOWN: Normal parsing starts with the root parser (ID=0)
	// and goes all the way down to the leaf parsers until an error is found.
	// The plain `parse...` methods are used.
	nState, aOut, err := p.ParseAny(ParentUnknown, state)
	out, _ := aOut.(Output)
	nextID := id
	for err != nil {
		Debugf("parseAll - got Error=%v", err)
		nState = nState.SaveError(err)
		if nState.AtEnd() || nState.constant.maxErrors <= 0 { // give up
			Debugf("parseAll - at EOF or recovery is turned off")
			return out, nState.Errors()
		}
		nState, nextID = pp.handleError(nState, err, recoverCache)
		if nextID < 0 { // give up
			Debugf("parseAll - no recoverer found")
			return out, nState.Errors()
		}
		p = pp.parsers[nextID]

		// BOTTOM->UP: Recovery parsing starts with a leaf parser
		// and goes all the way up to the root parser (with or without error).
		// The `parse...AfterError` methods are used.
		var newErr, nextErr *ParserError
		childID := nextID
		state = nState
		nextID, nState, aOut, newErr = p.parseAnyAfterError(err, state)
		if newErr != nil { // should never happen (or the recoverer didn't do its job)
			nextErr = newErr
		}
		for nextID >= 0 { // force the new result through all levels (error or not)
			p = pp.parsers[nextID]
			id = nextID
			nextID, nState, aOut, newErr = (p.(BranchParser)).parseAfterError(err, childID, state, nState, aOut, newErr)
			if newErr != nil && nextErr == nil {
				nextErr = newErr
			}
			Debugf("parseAll - parent (ID=%d) new Error?=%v", nextID, newErr)
			childID = id
		}
		err = nextErr
	}
	out, _ = aOut.(Output)
	return out, nState.Errors()
}

func (pp *PreparedParser[Output]) handleError(state State, err *ParserError, recoverCache []int,
) (newState State, nextID int32) {
	Debugf("handleError - parserID=%d, pos=%d, Error=%v", err.parserID, state.CurrentPos(), err)

	minWaste, minRec := pp.findMinWaste(err, state, recoverCache)

	if minWaste < 0 {
		Debugf("handleError - no recoverer found")
		return state.MoveBy(state.BytesRemaining()), RecoverWasteTooMuch
	}
	Debugf("handleError - best recoverer: ID=%d, waste=%d", minRec.ID(), minWaste)
	return state.MoveBy(minWaste), minRec.ID()
}

func (pp *PreparedParser[Output]) findMinWaste(pe *ParserError, state State, recoverCache []int,
) (minWaste int, minRec AnyParser) {
	failed := false
	minRec = pp.parsers[pe.parserID] // try the failed parser first
	minWaste = math.MaxInt
	if !minRec.IsStepRecoverer() {
		minWaste = pp.recover(pe, state, minRec, recoverCache)
		Debugf("findMinWaste - failed parser has fast recoverer: ID=%d, waste=%d", pe.parserID, minWaste)
		if minWaste < 0 { // recoverer is either forbidden or unsuccessful
			minWaste = math.MaxInt
		}
		failed = true
	}
	for _, rec := range pp.recoverers { // try all fast recoverers
		waste, data := rec.Recover(state, pe.ParserData(rec.ID()))
		if data != nil {
			pe.StoreParserData(rec.ID(), data)
		}
		if waste >= 0 && waste < minWaste {
			if waste == 0 { // it can't get better than this
				Debugf("findMinWaste - optimal fast recoverer: ID=%d, waste=%d", rec.ID(), waste)
				return waste, rec
			}
			minRec = rec
			minWaste = waste
		}
	}
	Debugf("findMinWaste - best fast recoverer: ID=%d, waste=%d", minRec.ID(), minWaste)
	stepRecs := pp.stepRecoverers
	if !failed {
		stepRecs = make([]AnyParser, len(pp.stepRecoverers)+1)
		copy(stepRecs, pp.stepRecoverers)
		stepRecs[len(pp.stepRecoverers)] = pp.parsers[pe.parserID]
		Debugf("findMinWaste - failed parser has slow recoverer: ID=%d", pe.parserID)
	}
	return pp.findMinStepWaste(stepRecs, state, pe, minWaste, minRec)
}

func (pp *PreparedParser[Output]) recover(pe *ParserError, state State, rec AnyParser, recoverCache []int) int {
	var data interface{}

	waste := recoverCache[rec.ID()]
	if waste < RecoverWasteUnknown {
		return waste
	}
	pos := state.CurrentPos()
	if waste >= 0 && waste >= pos {
		return waste - pos
	}
	waste, data = rec.Recover(state, pe.ParserData(rec.ID()))
	if data != nil {
		pe.StoreParserData(rec.ID(), data)
	}
	recoverCache[rec.ID()] = waste
	if waste >= 0 {
		recoverCache[rec.ID()] = pos + waste
	}
	return waste
}

func (pp *PreparedParser[Output]) findMinStepWaste(
	stepRecs []AnyParser, state State, err *ParserError, waste int, rec AnyParser,
) (minWaste int, minRec AnyParser) {
	maxWaste := waste
	if maxWaste == math.MaxInt {
		Debugf("findMinStepWaste - ALL fast recoverers failed!")
	}
	curState := state
	minWaste = 0
	for curState.BytesRemaining() > 0 && minWaste < maxWaste {
		for _, sr := range stepRecs {
			_, _, _, nErr := sr.parseAnyAfterError(err, curState)
			if nErr == nil {
				Debugf("findMinStepWaste - best slow recoverer: ID=%d, waste=%d", sr.ID(), minWaste)
				return minWaste, sr
			}
		}
		curState = curState.Delete1()
		minWaste = state.ByteCount(curState)
	}
	Debugf("findMinStepWaste - ALL slow recoverers failed!")
	if waste == math.MaxInt {
		return RecoverWasteTooMuch, rec
	}
	return waste, rec
}
