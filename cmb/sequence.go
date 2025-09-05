package cmb

import (
	"slices"

	"github.com/flowdev/comb"
)

// Sequence uses a list of parsers in order, one by one,
// until one fails, or they are all successful.
// The output of all successful parsers is returned as output.
// All parsers have to be of the same type.
//
// If any parser fails, this combinator produces an error result.
func Sequence[Output any](parsers ...comb.Parser[Output]) comb.Parser[[]Output] {
	if len(parsers) == 0 {
		panic("Sequence(missing parsers)")
	}

	fsd := &sequenceData[Output]{parsers: parsers}

	p := comb.NewBranchParser[[]Output]("Sequence", fsd.children, fsd.parseAfterChild)
	fsd.id = p.ID
	return p
}

type sequenceData[Output any] struct {
	id      func() int32
	parsers []comb.Parser[Output]
}

// partialSeqResult is internal to the parsing method and methods and functions called by it.
type partialSeqResult[Output any] struct {
	outs []Output
}

func (sd *sequenceData[Output]) children() []comb.AnyParser {
	children := make([]comb.AnyParser, len(sd.parsers))
	for i, p := range sd.parsers {
		children[i] = p
	}
	return children
}

func (sd *sequenceData[Output]) parseAfterChild(
	childID int32,
	childStartState, childState comb.State,
	childOut interface{},
	childErr *comb.ParserError,
	data interface{},
) (comb.State, []Output, *comb.ParserError, interface{}) {
	var res partialSeqResult[Output]
	var out Output
	idx := 0

	comb.Debugf("Sequence.parseAfterChild - childID=%d, pos=%d", childID, childState.CurrentPos())

	if childID >= 0 { // on the way up: Fetch
		res, _ = data.(partialSeqResult[Output])
		out, _ = childOut.(Output)
		idx = sd.indexForID(childID)
		if idx < 0 {
			childErr = childState.NewSemanticError("parsing after child with unknown ID %d", childID)
			childState = childState.SaveError(childErr)
			idx = -1 // will be incremented before usage
		} else {
			res.outs[idx] = out
		}

		if childErr != nil {
			return childState, res.outs, childErr, res
		}
		idx++
	} else {
		res.outs = make([]Output, len(sd.parsers))
	}

	for i := idx; i < len(sd.parsers); i++ {
		p := sd.parsers[i]
		childStartState = childState
		childState, childOut, childErr = p.ParseAny(sd.id(), childStartState)
		out, _ = childOut.(Output)
		res.outs[i] = out
		if childErr != nil {
			return childState, res.outs, childErr, res
		}
	}
	return childState, res.outs, nil, nil
}

func (sd *sequenceData[Output]) indexForID(id int32) int {
	return slices.IndexFunc(sd.parsers, func(p comb.Parser[Output]) bool {
		return p.ID() == id
	})
}
