package pcb

import (
	"github.com/oleiade/gomme"
)

// Sequence applies a sequence of parsers of the same type and
// returns either a slice of results or an error if any parser fails.
// Use one of the MapX parsers for differently typed parsers.
func Sequence[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[[]Output] {
	if len(parsers) == 0 {
		panic("Sequence(missing parsers)")
	}

	seq := &sequenceData[Output]{parsers: parsers}

	return gomme.NewBranchParser[[]Output]("Sequence", seq.children, seq.parseAfterChild)
}

type sequenceData[Output any] struct {
	parsers []gomme.Parser[Output]
}

func (seq *sequenceData[Output]) children() []gomme.AnyParser {
	children := make([]gomme.AnyParser, len(seq.parsers))
	for i, p := range seq.parsers {
		children[i] = p
	}
	return children
}

func (seq *sequenceData[Output]) parseAfterChild(childID int32, childResult gomme.ParseResult,
) gomme.ParseResult {
	gomme.Debugf("Sequence.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childResult.Error != nil {
		return childResult // we can't avoid any errors by going another path
	}

	remaining := childResult.EndState
	state := childResult.EndState
	idx := 0
	if childID >= 0 {
		state = childResult.StartState
		idx = seq.indexForID(childID)
		if idx < 0 {
			childResult.Error = childResult.EndState.NewSemanticError(
				"unable to parse after child with unknown ID %d", childID)
			return childResult
		}
		idx++
	}

	outputs := make([]Output, len(seq.parsers))
	for i := idx; i < len(seq.parsers); i++ {
		parse := seq.parsers[i]
		nState, out, err := parse.Parse(remaining)
		if err != nil {
			return gomme.ParseResult{StartState: remaining, EndState: nState, Output: out, Error: err}
		}
		outputs[i] = out
		remaining = nState
	}

	return gomme.ParseResult{StartState: state, EndState: remaining, Output: outputs, Error: nil}
}

func (seq *sequenceData[Output]) indexForID(id int32) int {
	for i, p := range seq.parsers {
		if p.ID() == id {
			return i
		}
	}
	return -1
}
