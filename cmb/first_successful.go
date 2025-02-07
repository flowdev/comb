package cmb

import (
	"github.com/flowdev/comb"
)

// FirstSuccessful tests a list of parsers in order, one by one,
// until one succeeds.
// All parsers have to be of the same type.
//
// If no parser succeeds, this combinator produces an error Result.
func FirstSuccessful[Output any](parsers ...comb.Parser[Output]) comb.Parser[Output] {
	if len(parsers) == 0 {
		panic("FirstSuccessful(missing parsers)")
	}

	fsd := &firstSuccessfulData[Output]{parsers: parsers}

	return comb.NewBranchParser[Output]("FirstSuccessful", fsd.children, fsd.parseAfterChild)
}

type firstSuccessfulData[Output any] struct {
	parsers []comb.Parser[Output]
}

// partialFSResult is internal to the parsing method and methods and functions called by it.
type partialFSResult[Output any] struct {
	out Output
	pos int
}

func (fsd *firstSuccessfulData[Output]) children() []comb.AnyParser {
	children := make([]comb.AnyParser, len(fsd.parsers))
	for i, p := range fsd.parsers {
		children[i] = p
	}
	return children
}

func (fsd *firstSuccessfulData[Output]) parseAfterChild(childID int32, childResult comb.ParseResult,
) comb.ParseResult {
	var bestRes partialFSResult[Output]
	var bestResult comb.ParseResult

	comb.Debugf("FirstSuccessful.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childID >= 0 { // on the way up: Fetch
		var o interface{}
		o, childResult = childResult.FetchOutput()
		bestRes, _ = o.(partialFSResult[Output])
	}

	if childID >= 0 && (childResult.Error == nil || childResult.StartState.SaveSpotMoved(childResult.EndState)) {
		return childResult.AddOutput(bestRes) // we can't avoid this error by going another path
	}

	idx := 0
	startResult := childResult
	bestResult.Error = childResult.Error
	if childID >= 0 {
		idx = fsd.indexForID(childID)
		if idx < 0 {
			childResult.Error = childResult.EndState.NewSemanticError(
				"unable to parse after child with unknown ID %d", childID)
			return childResult.AddOutput(bestRes)
		}
		startResult.EndState = childResult.StartState
		bestResult = childResult
		bestRes.out, _ = childResult.Output.(Output)
		bestRes.pos = childResult.EndState.CurrentPos()
		idx++
	}

	for i := idx; i < len(fsd.parsers); i++ {
		p := fsd.parsers[i]
		result := comb.RunParser(p, startResult)
		if result.Error == nil || startResult.EndState.SaveSpotMoved(result.EndState) {
			return result.AddOutput(bestRes)
		}

		// may the best error win:
		if i == 0 {
			bestResult = result
			bestRes.out, _ = result.Output.(Output)
			bestRes.pos = result.EndState.CurrentPos()
		} else {
			_, other := comb.BetterOf(bestResult.EndState, result.EndState)
			if other {
				bestResult = result
				bestRes.out, _ = result.Output.(Output)
				bestRes.pos = result.EndState.CurrentPos()
			}
		}
	}
	return bestResult.AddOutput(bestRes)
}

func (fsd *firstSuccessfulData[Output]) indexForID(id int32) int {
	for i, p := range fsd.parsers {
		if p.ID() == id {
			return i
		}
	}
	return -1
}
