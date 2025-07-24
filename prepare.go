package comb

import (
	"math"
	"slices"
)

// ParseResult is the result of a parser.
type ParseResult struct {
	StartState State // state before parsing
	EndState   State // state after parsing
	Output     interface{}
	Error      *ParserError
}

// ============================================================================
// Interfaces And Function For Parser Preparation
//

// AnyParser is an internal interface used by PreparedParser.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
type AnyParser interface {
	ID() int32
	LastParent() int32
	parse(parent int32, state State) ParseResult
	IsSaveSpot() bool
	Recover(*ParserError, State) int
	IsStepRecoverer() bool
	setID(int32)     // only sets own ID
	setParent(int32) // sets current parent ID
}

// BranchParser is a more internal interface used by orchestrators.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
// BranchParser just adds 2 methods to the Parser and AnyParser interfaces.
type BranchParser interface {
	children() []AnyParser
	parseAfterError(pe *ParserError, childID, parentID int32, childResult ParseResult) ParseResult
}

// RunParser runs any parser and is able to handle branch parsers specially.
// That is necessary to run child parsers of branch parsers correctly.
func RunParser(ap AnyParser, parent int32, inResult ParseResult) ParseResult {
	if bp, ok := ap.(BranchParser); ok {
		return bp.parseAfterError(nil, -1, parent, inResult)
	}
	return ap.parse(parent, inResult.EndState)
}

// ============================================================================
// PreparedParser: Data Structures And Construction
//

type parserData struct { // all data about a single parser
	parser   AnyParser
	parentID int32
}
type PreparedParser[Output any] struct {
	parsers        []parserData
	recoverers     []AnyParser
	stepRecoverers []AnyParser
}

// NewPreparedParser prepares a parser for error recovery.
// Call this directly if you have a parser that you want to run on many inputs.
// You can use this together with RunOnState.
func NewPreparedParser[Output any](p Parser[Output]) *PreparedParser[Output] {
	pp := &PreparedParser[Output]{
		parsers:        make([]parserData, 0, 64),
		recoverers:     make([]AnyParser, 0, 64),
		stepRecoverers: make([]AnyParser, 0, 64),
	}
	pp.registerParsers(p, -1)
	return pp
}

func (pp *PreparedParser[Output]) registerParsers(ap AnyParser, parentID int32) {
	id := int32(len(pp.parsers))
	ap.setID(id)
	ap.setParent(parentID)
	pp.parsers = append(pp.parsers, parserData{parser: ap, parentID: parentID})

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
	var zero Output
	var id int32 = 0 // this is always the root parser
	recoverCache := slices.Repeat([]int{RecoverWasteUnknown}, len(pp.parsers))
	p := pp.parsers[id]

	// TOP->DOWN: Normal parsing starts with the root parser (ID=0)
	// and goes all the way down to the leaf parsers until an error is found.
	// The childID is ALWAYS < 0.
	// ParseResult.AddOutput and .setID are used;
	//   .FetchOutput and .prepareOutputFor are NOT used.
	result := p.parser.parse(-1, state)
	nextID, nState := id, result.EndState
	for result.Error != nil {
		pe := result.Error
		Debugf("parseAll - got Error=%v", pe)
		nState = result.EndState.SaveError(pe)
		if nState.AtEnd() || nState.constant.maxErrors <= 0 { // give up
			Debugf("parseAll - at EOF or recovery is turned off")
			return zero, nState.Errors()
		}
		result.EndState = nState
		nState, nextID = pp.handleError(result, recoverCache)
		if nextID < 0 { // give up
			Debugf("parseAll - no recoverer found")
			return zero, nState.Errors()
		}
		p = pp.parsers[nextID]
		result.EndState = nState

		// BOTTOM->UP: Recovery parsing starts with a leaf parser
		// and goes all the way up to the root parser (with or without error).
		// The childID is NEVER < 0 and err is NEVER nil.
		result = RunParser(p.parser, ParentUnknown, result) // should always be successful (or the recoverer didn't do its job)
		parentID := p.parser.LastParent()
		for parentID >= 0 { // force the new result through all levels (error or not)
			if result.Error != nil {
				pe = result.Error
			}
			childID := nextID
			nextID = parentID
			p = pp.parsers[nextID]
			result = (p.parser.(BranchParser)).parseAfterError(pe, childID, ParentUnknown, result)
			Debugf("parseAll - parent (ID=%d) new Error?=%v", nextID, result.Error)
			parentID = p.parser.LastParent()
		}
	}
	out, _ := result.Output.(Output)
	return out, result.EndState.Errors()
}

func (pp *PreparedParser[Output]) handleError(r ParseResult, recoverCache []int) (state State, nextID int32) {
	Debugf("handleError - parserID=%d, pos=%d, Error=%v", r.Error.parserID, r.EndState.CurrentPos(), r.Error)

	minWaste, minRec := pp.findMinWaste(r.Error, r.EndState, recoverCache)

	if minWaste < 0 {
		Debugf("handleError - no recoverer found")
		return r.EndState.MoveBy(r.EndState.BytesRemaining()), RecoverWasteTooMuch
	}
	Debugf("handleError - best recoverer: ID=%d, waste=%d", minRec.ID(), minWaste)
	return r.EndState.MoveBy(minWaste), minRec.ID()
}

func (pp *PreparedParser[Output]) findMinWaste(pe *ParserError, state State, recoverCache []int,
) (minWaste int, minRec AnyParser) {
	failed := false
	minRec = pp.parsers[pe.parserID].parser // try the failed parser first
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
		if waste := rec.Recover(pe, state); waste >= 0 && waste < minWaste {
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
		stepRecs[len(pp.stepRecoverers)] = pp.parsers[pe.parserID].parser
		Debugf("findMinWaste - failed parser has slow recoverer: ID=%d", pe.parserID)
	}
	return pp.findMinStepWaste(stepRecs, state, minWaste, minRec)
}

func (pp *PreparedParser[Output]) recover(pe *ParserError, state State, rec AnyParser, recoverCache []int) int {
	waste := recoverCache[rec.ID()]
	if waste < RecoverWasteUnknown {
		return waste
	}
	pos := state.CurrentPos()
	if waste >= 0 && waste >= pos {
		return waste - pos
	}
	waste = rec.Recover(pe, state)
	recoverCache[rec.ID()] = waste
	if waste >= 0 {
		recoverCache[rec.ID()] = pos + waste
	}
	return waste
}

func (pp *PreparedParser[Output]) findMinStepWaste(stepRecs []AnyParser, state State, waste int, rec AnyParser,
) (minWaste int, minRec AnyParser) {
	maxWaste := waste
	if maxWaste == math.MaxInt {
		Debugf("findMinStepWaste - ALL fast recoverers failed!")
	}
	curState := state
	minWaste = 0
	for curState.BytesRemaining() > 0 && minWaste < maxWaste {
		for _, sr := range stepRecs {
			result := sr.parse(-1, curState)
			if result.Error == nil {
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
