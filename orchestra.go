package gomme

import (
	"cmp"
	"math"
	"slices"
)

// ParseResult is the result of a (leaf) parser.
type ParseResult struct {
	ID    int32 // ID of the parser that produced the result
	State State
	Error *ParserError
}

// RecoverResult is the result of a recoverer.
type RecoverResult struct {
	ID    int32 // ID of the parser that produced the result
	Waste int   // bytes ignored if this recoverer is chosen
}

// AnyParser is a more internal interface used by orchestrators.
// It intentionally avoids generics for easy storage of parsers in collections
// (slices, maps, ...).
type AnyParser interface {
	SetID(int32) // only sets own ID
	ID() int32
	Parse(state State, store Store) ParseResult
	IsSaveSpot() bool
	Recover(state State) int
}

type BranchParser interface {
	AnyParser
	Children() []AnyParser
	ParseAfterChild(childID int32, state State, store Store) ParseResult
}

type anyParser[Output any] struct {
	id     int32
	parser Parser[Output]
}

func (ap *anyParser[Output]) SetID(id int32) {
	ap.id = id
}

func (ap *anyParser[Output]) ID() int32 {
	return ap.id
}

func (ap *anyParser[Output]) Parse(state State, store Store) ParseResult {
	nState, output, err := ap.parser.It(state)
	store.PutOutput(ap.id, state.CurrentPos(), output)
	return ParseResult{ID: ap.id, State: nState, Error: err}
}

func (ap *anyParser[Output]) IsSaveSpot() bool {
	return ap.parser.IsSaveSpot()
}

func (ap *anyParser[Output]) Recover(state State) int {
	return ap.parser.Recover(state)
}

func (ap *anyParser[Output]) Result(state State) Output {
	output, ok := state.CachedOutput(ap.id)
	if !ok {
		return ZeroOf[Output]()
	}
	return output.(Output)
}

// ParserToAnyParser converts a normal parser to the AnyParser interface used
// internally. Only a BranchParser should use this.
func ParserToAnyParser[Output any](p Parser[Output]) AnyParser {
	if ap, ok := interface{}(p).(AnyParser); ok {
		return ap
	}
	return &anyParser[Output]{parser: p}
}

type Store interface {
	PutData(id int32, pos int, data interface{})
	GetData(id int32, pos int) interface{}
	PutOutput(id int32, pos int, output interface{})
	GetOutput(id int32, pos int) (output interface{}, ok bool)
}

type orchestraResult struct { // data for a single parser execution at a single position
	pos    int
	data   interface{}
	output interface{}
}
type orchestraData struct { // all data about a single parser
	parser   AnyParser
	parentID int32
	data     []orchestraResult
}
type orchestrator[Output any] struct {
	parsers        []orchestraData
	recoverers     []AnyParser
	stepRecoverers []AnyParser
}

func newOrchestrator[Output any](p Parser[Output]) *orchestrator[Output] {
	o := &orchestrator[Output]{
		parsers:    make([]orchestraData, 0, 64),
		recoverers: make([]AnyParser, 0, 64),
	}
	o.registerParsers(ParserToAnyParser(p), -1)
	return o
}
func (o *orchestrator[Output]) parseAll(state State) (State, Output, error) {
	var zero Output
	var id int32 = 0
	p := o.parsers[id]
	result := p.parser.Parse(state, o)
	for nextID := id; result.Error != nil || nextID >= 0; {
		nState := result.State
		if result.Error != nil {
			nState = result.State.SaveError(result.Error)
			if nState.AtEnd() { // give up
				return nState, zero, nState.Errors()
			}
			result.State = nState
			nState, nextID = o.handleError(result)
			if nextID < 0 { // give up
				return nState, zero, nState.Errors()
			}
			p = o.parsers[nextID]
			result = p.parser.Parse(nState, o)
		} else {
			oldID := nextID
			nextID = p.parentID
			if nextID >= 0 {
				p = o.parsers[nextID]
				result = (p.parser.(BranchParser)).ParseAfterChild(oldID, nState, o)
			}
		}
	}
	output, _ := o.GetOutput(id, state.CurrentPos())
	return result.State, output.(Output), result.State.Errors()
}
func (o *orchestrator[Output]) handleError(r ParseResult) (state State, nextID int32) {
	pos := r.State.CurrentPos()
	if !r.State.recover { // error recovery is turned off
		state = r.State.NewSemanticError("error recovery is turned off").MoveBy(r.State.BytesRemaining())
		Debugf("handleError - recovery is turned off: parserID=%d, pos=%d", r.ID, pos)
		return state, -1
	}

	Debugf("handleError - start: parserID=%d, pos=%d", r.ID, pos)

	minWaste, minRec := o.findMinWaste(r.State, r.ID)

	if minWaste < 0 {
		state = r.State.NewSemanticError("unable to recover from error").MoveBy(r.State.BytesRemaining())
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
	if minRec.Recover != nil {
		minWaste = minRec.Recover(state)
		if minWaste < 0 {
			minWaste = math.MaxInt
		}
		failed = true
		Debugf("findMinWaste - failed parser has fast recoverer: ID=%d, waste=%d", id, minWaste)
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
		stepRecs = make([]AnyParser, len(o.stepRecoverers), len(o.stepRecoverers)+1)
		copy(stepRecs, o.stepRecoverers)
		stepRecs = append(stepRecs, o.parsers[id].parser)
		Debugf("findMinWaste - failed parser has slow recoverer: ID=%d", id)
	}
	return o.findMinStepWaste(stepRecs, state, minWaste, minRec)
}
func (o *orchestrator[Output]) findMinStepWaste(stepRecs []AnyParser, state State, waste int, rec AnyParser,
) (minWaste int, minRec AnyParser) {
	maxWaste := waste
	if maxWaste < 0 {
		maxWaste = math.MaxInt
		Debugf("findMinStepWaste - ALL fast recoverers failed!")
	}
	curState := state
	minWaste = 0
	for curState.BytesRemaining() > 0 && minWaste < maxWaste {
		for _, sr := range stepRecs {
			result := sr.Parse(curState, o)
			if result.Error == nil {
				Debugf("findMinStepWaste - best slow recoverer: ID=%d, waste=%d", sr.ID(), minWaste)
				return minWaste, sr
			}
		}
		curState = curState.Delete(1)
		minWaste = state.ByteCount(curState)
	}
	Debugf("findMinStepWaste - ALL slow recoverers failed!")
	return waste, rec
}

func (o *orchestrator[Output]) PutData(id int32, pos int, data interface{}) {
	o.storagePutFunc(id, pos, func(r orchestraResult) orchestraResult {
		r.data = data
		return r
	})
}
func (o *orchestrator[Output]) GetData(id int32, pos int) interface{} {
	return o.storageGetFunc(id, pos).data
}
func (o *orchestrator[Output]) PutOutput(id int32, pos int, output interface{}) {
	o.storagePutFunc(id, pos, func(r orchestraResult) orchestraResult {
		r.output = output
		return r
	})
}
func (o *orchestrator[Output]) GetOutput(id int32, pos int) (output interface{}, ok bool) {
	result := o.storageGetFunc(id, pos)
	if result.pos < 0 {
		return nil, false
	}
	return result.output, true
}
func (o *orchestrator[Output]) storagePutFunc(id int32, pos int, f func(orchestraResult) orchestraResult) {
	results := o.parsers[id].data
	i := slices.IndexFunc(results, func(r orchestraResult) bool {
		return cmp.Compare(r.pos, pos) == 0 || r.pos < 0
	})
	if i == -1 {
		o.parsers[id].data = append(results, f(orchestraResult{pos: pos}))
	}
	results[i] = f(results[i])
}
func (o *orchestrator[Output]) storageGetFunc(id int32, pos int) orchestraResult {
	results := o.parsers[id].data
	i := slices.IndexFunc(results, func(r orchestraResult) bool {
		return cmp.Compare(r.pos, pos) == 0
	})
	if i == -1 {
		return orchestraResult{pos: -1}
	}
	return results[i]
}
func (o *orchestrator[Output]) registerParsers(ap AnyParser, parentID int32) {
	id := int32(len(o.parsers))
	ap.SetID(id)
	o.parsers = append(o.parsers, orchestraData{parser: ap, parentID: parentID})

	if bp, ok := ap.(BranchParser); ok {
		for _, cp := range bp.Children() {
			o.registerParsers(cp, id)
		}
	} else if ap.IsSaveSpot() {
		if ap.Recover == nil {
			o.stepRecoverers = append(o.stepRecoverers, ap)
		} else {
			o.recoverers = append(o.recoverers, ap)
		}
	}
}
