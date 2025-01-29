package pcb

import (
	"github.com/oleiade/gomme"
)

// FirstSuccessful tests a list of parsers in order, one by one,
// until one succeeds.
// All parsers have to be of the same type.
//
// If no parser succeeds, this combinator produces an error Result.
func FirstSuccessful[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[Output] {
	if len(parsers) == 0 {
		panic("FirstSuccessful(missing parsers)")
	}

	fsd := &firstSuccessfulData[Output]{parsers: parsers}

	return gomme.NewBranchParser[Output]("FirstSuccessful", fsd.children, fsd.parseAfterChild)
}

type firstSuccessfulData[Output any] struct {
	parsers []gomme.Parser[Output]
}

func (fsd *firstSuccessfulData[Output]) children() []gomme.AnyParser {
	children := make([]gomme.AnyParser, len(fsd.parsers))
	for i, p := range fsd.parsers {
		children[i] = p
	}
	return children
}

func (fsd *firstSuccessfulData[Output]) parseAfterChild(childID int32, childResult gomme.ParseResult,
) gomme.ParseResult {
	gomme.Debugf("FirstSuccessful.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childID >= 0 && (childResult.Error == nil || childResult.StartState.SaveSpotMoved(childResult.EndState)) {
		return childResult
	}

	state := childResult.EndState
	idx := 0
	bestState := childResult.EndState
	bestOut, _ := childResult.Output.(Output)
	bestErr := childResult.Error
	if childID >= 0 {
		state = childResult.StartState
		idx = fsd.indexForID(childID)
		if idx < 0 {
			childResult.Error = childResult.EndState.NewSemanticError(
				"unable to parse after child with unknown ID %d", childID)
			return childResult
		}
		idx++
	}

	for i := idx; i < len(fsd.parsers); i++ {
		p := fsd.parsers[i]
		nState, out, err := p.Parse(state)
		if err == nil || state.SaveSpotMoved(nState) {
			return gomme.ParseResult{StartState: state, EndState: nState, Output: out, Error: err}
		}

		// may the best error win:
		if i == 0 {
			bestState = nState
			bestOut = out
			bestErr = err
		} else {
			other := false
			bestState, other = gomme.BetterOf(bestState, nState)
			if other {
				bestOut = out
				bestErr = err
			}
		}
	}
	return gomme.ParseResult{StartState: state, EndState: bestState, Output: bestOut, Error: bestErr}
}

func (fsd *firstSuccessfulData[Output]) indexForID(id int32) int {
	for i, p := range fsd.parsers {
		if p.ID() == id {
			return i
		}
	}
	return -1
}
