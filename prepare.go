package gomme

import (
	"math"
	"slices"
)

// ============================================================================
// Data Types For General Parser Preparation
//

// ParseResult is the result of a parser.
type ParseResult struct {
	StartState State // state before parsing
	EndState   State // state after parsing
	Output     interface{}
	Error      *ParserError
}

// BranchParser is a more internal interface used by orchestrators.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
// BranchParser just adds 2 methods to the Parser and AnyParser interfaces.
type BranchParser interface {
	children() []AnyParser
	parseAfterChild(childID int32, childResult ParseResult) ParseResult
}

// AnyParser is an internal interface used by PreparedParser.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
type AnyParser interface {
	ID() int32
	parse(state State) ParseResult
	IsSaveSpot() bool
	Recover(state State) int
	IsStepRecoverer() bool
	setID(int32) // only sets own ID
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
	recoverCache   []int
}

// NewPreparedParser prepares a parser for error recovery.
// Call this directly if you have a parser that you want to run on many inputs.
// You can use this together with RunOnState.
func NewPreparedParser[Output any](p Parser[Output]) *PreparedParser[Output] {
	o := &PreparedParser[Output]{
		parsers:        make([]parserData, 0, 64),
		recoverers:     make([]AnyParser, 0, 64),
		stepRecoverers: make([]AnyParser, 0, 64),
	}
	o.registerParsers(p, -1)
	o.recoverCache = slices.Repeat([]int{RecoverWasteUnknown}, len(o.parsers))
	return o
}

func (pp *PreparedParser[Output]) registerParsers(ap AnyParser, parentID int32) {
	id := int32(len(pp.parsers))
	ap.setID(id)
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
	recoverCache := slices.Clone(pp.recoverCache)

	p := pp.parsers[id]
	result := p.parser.parse(state)
	nextID, nState := id, result.EndState
	for result.Error != nil {
		Debugf("parseAll - got Error=%v", result.Error)
		nState = result.EndState.SaveError(result.Error)
		if nState.AtEnd() || !nState.recover { // give up
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
		result = p.parser.parse(nState)
		for p.parentID >= 0 { // force the new result through all levels (error or not)
			childID := nextID
			nextID = p.parentID
			p = pp.parsers[nextID]
			result = (p.parser.(BranchParser)).parseAfterChild(childID, result)
			Debugf("parseAll - parent (ID=%d) Error?=%v", nextID, result.Error)
		}
	}
	return result.Output.(Output), result.EndState.Errors()
}

func (pp *PreparedParser[Output]) handleError(r ParseResult, recoverCache []int) (state State, nextID int32) {
	Debugf("handleError - parserID=%d, pos=%d, Error=%v", r.Error.parserID, r.EndState.CurrentPos(), r.Error)

	minWaste, minRec := pp.findMinWaste(r.EndState, r.Error.parserID, recoverCache)

	if minWaste < 0 {
		Debugf("handleError - no recoverer found")
		return r.EndState.MoveBy(r.EndState.BytesRemaining()), RecoverWasteTooMuch
	}
	Debugf("handleError - best recoverer: ID=%d, waste=%d", minRec.ID(), minWaste)
	return r.EndState.MoveBy(minWaste), minRec.ID()
}

func (pp *PreparedParser[Output]) findMinWaste(state State, id int32, recoverCache []int,
) (minWaste int, minRec AnyParser) {
	failed := false
	minRec = pp.parsers[id].parser // try failed parser first
	minWaste = math.MaxInt
	if !minRec.IsStepRecoverer() {
		minWaste = pp.recover(state, minRec, recoverCache)
		Debugf("findMinWaste - failed parser has fast recoverer: ID=%d, waste=%d", id, minWaste)
		if minWaste < 0 { // recoverer is either forbidden or unsuccessful
			minWaste = math.MaxInt
		}
		failed = true
	}
	for _, rec := range pp.recoverers { // try all fast recoverers
		if waste := rec.Recover(state); waste >= 0 && waste < minWaste {
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
		stepRecs[len(pp.stepRecoverers)] = pp.parsers[id].parser
		Debugf("findMinWaste - failed parser has slow recoverer: ID=%d", id)
	}
	return pp.findMinStepWaste(stepRecs, state, minWaste, minRec)
}

func (pp *PreparedParser[Output]) recover(state State, rec AnyParser, recoverCache []int) int {
	waste := recoverCache[rec.ID()]
	if waste < RecoverWasteUnknown {
		return waste
	}
	pos := state.CurrentPos()
	if waste >= 0 && waste >= pos {
		return waste - pos
	}
	waste = rec.Recover(state)
	recoverCache[rec.ID()] = waste
	if waste == RecoverWasteNever {
		pp.recoverCache[rec.ID()] = waste
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
			result := sr.parse(curState)
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
