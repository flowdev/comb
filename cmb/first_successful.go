package cmb

import (
	"slices"

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

	p := comb.NewBranchParser[Output]("FirstSuccessful", fsd.children, fsd.parseAfterChild)
	fsd.id = p.ID
	return p
}

type firstSuccessfulData[Output any] struct {
	id      func() int32
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

func (fsd *firstSuccessfulData[Output]) parseAfterChild(
	childID int32,
	childStartState, childState comb.State,
	childOut interface{},
	childErr *comb.ParserError,
	data interface{},
) (comb.State, Output, *comb.ParserError, interface{}) {
	var bestRes partialFSResult[Output]
	var bestState comb.State
	var bestOut Output
	var bestErr *comb.ParserError

	comb.Debugf("FirstSuccessful.parseAfterChild - childID=%d, pos=%d", childID, childState.CurrentPos())

	bestState = childState
	if childID >= 0 { // on the way up: Fetch
		bestRes, _ = data.(partialFSResult[Output])
		bestOut, _ = childOut.(Output)
		bestState = childState
		bestErr = childErr

		if childErr == nil {
			return bestState, bestOut, nil, nil
		} else if childStartState.SafeSpotMoved(childState) {
			return bestState, bestOut, bestErr, bestRes // we can't avoid this error by going another path
		}
	}

	idx := 0
	if childID >= 0 {
		idx = fsd.indexForID(childID)
		if idx < 0 {
			bestErr = childState.NewSemanticError("parsing after child with unknown ID %d", childID)
			childStartState = childStartState.SaveError(bestErr)
			bestState = bestState.SaveError(bestErr)
			idx = -1 // will be 0 before usage
		}
		bestRes.pos = childState.CurrentPos()
		idx++
	}

	for i := idx; i < len(fsd.parsers); i++ {
		p := fsd.parsers[i]
		childState, childOut, childErr = p.ParseAny(fsd.id(), childStartState)
		if childErr == nil {
			bestRes.out, _ = childOut.(Output)
			return childState, bestRes.out, nil, nil
		} else if childStartState.SafeSpotMoved(childState) {
			bestRes.out, _ = childOut.(Output)
			bestRes.pos = childState.CurrentPos()
			return childState, bestRes.out, childErr, bestRes // we can't avoid this error by going another path
		}

		// may the best error win:
		if i == 0 {
			bestState = childState
			bestOut, _ = childOut.(Output)
			bestErr = childErr
			bestRes.out, _ = childOut.(Output)
			bestRes.pos = childState.CurrentPos()
		} else {
			_, other := comb.BetterOf(bestState, childState)
			if other {
				bestState = childState
				bestOut, _ = childOut.(Output)
				bestErr = childErr
				bestRes.out, _ = childOut.(Output)
				bestRes.pos = childState.CurrentPos()
			}
		}
	}
	return bestState, bestOut, bestErr, bestRes
}

func (fsd *firstSuccessfulData[Output]) indexForID(id int32) int {
	return slices.IndexFunc(fsd.parsers, func(p comb.Parser[Output]) bool {
		return p.ID() == id
	})
}
