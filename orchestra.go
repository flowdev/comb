package gomme

import (
	"math"
)

// ============================================================================
// AnyParser interface and implementation
//

// ParseResult is the result of a (leaf) parser.
type ParseResult struct {
	ID     int32       // ID of the parser that produced the result
	State  State       // state after parsing
	Output interface{} // output of the parser
	Error  *ParserError
}

// BranchParser is a more internal interface used by orchestrators.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
// BranchParser just adds 2 methods to the Parser and AnyParser interfaces.
type BranchParser interface {
	children() []AnyParser
	parseAfterChild(childResult ParseResult) ParseResult
}

// AnyParser is an internal interface used by the orchestrator.
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
// orchestra data structures and construction
//

type parserData struct { // all data about a single parser
	parser   AnyParser
	parentID int32
}
type orchestrator[Output any] struct {
	parsers        []parserData
	recoverers     []AnyParser
	stepRecoverers []AnyParser
}

func newOrchestrator[Output any](p Parser[Output]) *orchestrator[Output] {
	o := &orchestrator[Output]{
		parsers:    make([]parserData, 0, 64),
		recoverers: make([]AnyParser, 0, 64),
	}
	o.registerParsers(p, -1)
	return o
}
func (o *orchestrator[Output]) registerParsers(ap AnyParser, parentID int32) {
	id := int32(len(o.parsers))
	ap.setID(id)
	o.parsers = append(o.parsers, parserData{parser: ap, parentID: parentID})

	if bp, ok := ap.(BranchParser); ok {
		for _, cp := range bp.children() {
			o.registerParsers(cp, id)
		}
	} else if ap.IsSaveSpot() {
		if ap.IsStepRecoverer() {
			o.stepRecoverers = append(o.stepRecoverers, ap)
		} else {
			o.recoverers = append(o.recoverers, ap)
		}
	}
}

// ============================================================================
// parser orchestration: parseAll
//

func (o *orchestrator[Output]) parseAll(state State) (Output, error) {
	var zero Output
	var id int32 = 0 // this is always the root parser
	p := o.parsers[id]
	result := p.parser.parse(state)
	nextID := id
	for result.Error != nil || nextID != id {
		nState := result.State.SaveError(result.Error)
		if nState.AtEnd() { // give up
			return zero, nState.Errors()
		}
		result.State = nState
		nState, nextID = o.handleError(result)
		if nextID < 0 { // give up
			return zero, nState.Errors()
		}
		p = o.parsers[nextID]
		result = p.parser.parse(nState)
		for p.parentID >= 0 && result.ID == nextID {
			nextID = p.parentID
			p = o.parsers[nextID]
			result = (p.parser.(BranchParser)).parseAfterChild(result)
		}
	}
	return result.Output.(Output), result.State.Errors()
}
func (o *orchestrator[Output]) handleError(r ParseResult) (state State, nextID int32) {
	pos := r.State.CurrentPos()
	if !r.State.recover { // error recovery is turned off
		state = r.State.SaveError(r.State.NewSemanticError("error recovery is turned off")).MoveBy(r.State.BytesRemaining())
		Debugf("handleError - recovery is turned off: parserID=%d, pos=%d", r.Error.parserID, pos)
		return state, -1
	}

	Debugf("handleError - start: parserID=%d, pos=%d", r.Error.parserID, pos)

	minWaste, minRec := o.findMinWaste(r.State, r.ID)

	if minWaste < 0 {
		state = r.State.MoveBy(r.State.BytesRemaining())
		Debugf("handleError - no recoverer found")
		return state, -1
	}
	Debugf("handleError - best recoverer: ID=%d, waste=%d", minRec.ID(), minWaste)
	return r.State.MoveBy(minWaste), minRec.ID()
}
func (o *orchestrator[Output]) findMinWaste(state State, id int32) (minWaste int, minRec AnyParser) {
	failed := false
	minRec = o.parsers[id].parser // try failed parser first
	minWaste = math.MaxInt
	if !minRec.IsStepRecoverer() {
		minWaste = minRec.Recover(state)
		Debugf("findMinWaste - failed parser has fast recoverer: ID=%d, waste=%d", id, minWaste)
		if minWaste < 0 {
			minWaste = math.MaxInt
		}
		failed = true
	}
	for _, rec := range o.recoverers { // try all fast recoverers
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
	stepRecs := o.stepRecoverers
	if !failed {
		stepRecs = make([]AnyParser, len(o.stepRecoverers)+1)
		copy(stepRecs, o.stepRecoverers)
		stepRecs[len(o.stepRecoverers)] = o.parsers[id].parser
		Debugf("findMinWaste - failed parser has slow recoverer: ID=%d", id)
	}
	return o.findMinStepWaste(stepRecs, state, minWaste, minRec)
}
func (o *orchestrator[Output]) findMinStepWaste(stepRecs []AnyParser, state State, waste int, rec AnyParser,
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
		return -1, rec
	}
	return waste, rec
}
