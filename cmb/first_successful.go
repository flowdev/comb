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
	err *comb.ParserError, childID int32, childResult comb.ParseResult,
) comb.ParseResult {
	var bestRes partialFSResult[Output]
	var bestResult comb.ParseResult

	comb.Debugf("FirstSuccessful.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childID >= 0 { // on the way up: Fetch
		o := err.ParserData(fsd.id())
		bestRes, _ = o.(partialFSResult[Output])

		if childResult.Error == nil {
			return childResult
		} else if childResult.StartState.SafeSpotMoved(childResult.EndState) {
			childResult.Error.StoreParserData(fsd.id(), bestRes) // we can't avoid this error by going another path
			return childResult
		}
	}

	idx := 0
	startResult := childResult
	bestResult.Error = childResult.Error
	if childID >= 0 {
		idx = fsd.indexForID(childID)
		if idx < 0 {
			childResult.Error = childResult.EndState.NewSemanticError(fsd.id(),
				"unable to parse after child with unknown ID %d", childID)
			childResult.Error.StoreParserData(fsd.id(), bestRes)
			return childResult
		}
		startResult.EndState = childResult.StartState
		bestResult = childResult
		bestRes.out, _ = childResult.Output.(Output)
		bestRes.pos = childResult.EndState.CurrentPos()
		idx++
	}

	for i := idx; i < len(fsd.parsers); i++ {
		p := fsd.parsers[i]
		result := comb.RunParser(p, fsd.id(), startResult)
		if result.Error == nil {
			return result
		} else if startResult.EndState.SafeSpotMoved(result.EndState) {
			result.Error.StoreParserData(fsd.id(), bestRes) // we can't avoid this error by going another path
			return result
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
	if bestResult.Error != nil {
		bestResult.Error.StoreParserData(fsd.id(), bestRes)
	}
	return bestResult
}

func (fsd *firstSuccessfulData[Output]) indexForID(id int32) int {
	for i, p := range fsd.parsers {
		if p.ID() == id {
			return i
		}
	}
	return -1
}
