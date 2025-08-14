package cmb

import (
	"github.com/flowdev/comb"
)

// SeparatedMN applies an element parser and a separator parser repeatedly in order
// to produce a slice of elements.
//
// Because SeparatedMN is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed if the separator parser fails to
// match at the end.
//
// If the separator parser is nil, SeparatedMN acts as ManyMN.
//
// The parser will fail if both parsers together accepted an empty input
// to prevent infinite loops.
func SeparatedMN[Output any, S comb.Separator](
	parser comb.Parser[Output], separator comb.Parser[S],
	atLeast, atMost int,
	parseSeparatorAtEnd bool,
) comb.Parser[[]Output] {
	if atLeast < 0 {
		panic("SeparatedMN is unable to handle negative `atLeast`")
	}
	if atMost < 0 {
		panic("SeparatedMN is unable to handle negative `atMost`")
	}

	expected := "SeparatedMN"
	if separator == nil {
		expected = "ManyMN"
	}
	sd := &separatedData[Output, S]{
		parser:              parser,
		separator:           separator,
		atLeast:             atLeast,
		atMost:              atMost,
		parseSeparatorAtEnd: parseSeparatorAtEnd,
	}
	p := comb.NewBranchParser[[]Output](expected, sd.children, sd.parseAfterChild)
	sd.id = p.ID
	return p
}

type separatedData[Output any, S comb.Separator] struct {
	id                  func() int32
	parser              comb.Parser[Output]
	separator           comb.Parser[S]
	atLeast             int
	atMost              int
	parseSeparatorAtEnd bool
}

// partialSepResult is internal to the parsing method and methods and functions called by it.
type partialSepResult[Output any] struct {
	outs []Output
}

func (sd *separatedData[Output, S]) children() []comb.AnyParser {
	if sd.separator == nil {
		return []comb.AnyParser{sd.parser}
	}
	return []comb.AnyParser{sd.parser, sd.separator}
}

func (sd *separatedData[Output, S]) parseAfterChild(
	childID int32,
	childStartState, childState comb.State,
	childOut interface{},
	childErr *comb.ParserError,
	data interface{},
) (comb.State, []Output, *comb.ParserError, interface{}) {
	var partRes partialSepResult[Output]

	comb.Debugf("SeparatedMN.parseAfterChild - childID=%d, pos=%d", childID, childState.CurrentPos())

	if childID >= 0 { // on the way up: Fetch
		partRes, _ = data.(partialSepResult[Output])
	} else {
		partRes.outs = make([]Output, 0, min(32, sd.atMost))
	}

	if childErr != nil {
		if sd.atLeast > len(partRes.outs) || childStartState.SafeSpotMoved(childState) { // fail
			return childState, partRes.outs, childErr, partRes
		}
		return childState, partRes.outs, nil, nil
	}

	if childID >= 0 && childID != sd.parser.ID() && childID != sd.separator.ID() {
		childErr = childState.NewSemanticError("unable to parser after child with unknown ID %d", childID)
		childState = childState.SaveError(childErr)
		return childState, partRes.outs, childErr, partRes
	}

	count := len(partRes.outs)
	if childID == sd.parser.ID() {
		out, _ := childOut.(Output)
		partRes.outs = append(partRes.outs, out)
		count++
	}

	endState := childState    // state including separator
	resultState := childState // state for the result (probably without separator)
	for {
		if count >= sd.atMost {
			return resultState, partRes.outs, nil, nil
		}

		if childID != sd.parser.ID() {
			childStartState = endState
			childState, childOut, childErr = sd.parser.ParseAny(sd.id(), childStartState)
			if childErr != nil {
				if sd.atLeast > count || childStartState.SafeSpotMoved(childState) { // fail
					return childState, partRes.outs, childErr, partRes
				}
				return resultState, partRes.outs, nil, nil // ignore error: we have enough output
			}
			out, _ := childOut.(Output)
			partRes.outs = append(partRes.outs, out)
			count++
			endState = childState
			resultState = childState
			childID = -1 // go on normally
		} else {
			childID = -1 // go on normally
		}

		if sd.separator != nil {
			sepState := childState
			sepState, childOut, childErr = sd.separator.ParseAny(sd.id(), childState)
			if childErr != nil {
				if sd.atLeast > count || childState.SafeSpotMoved(sepState) { // fail
					return sepState, partRes.outs, childErr, partRes
				}
				return childState, partRes.outs, nil, nil // ignore error: we have enough output
			}
			endState = sepState
			if sd.parseSeparatorAtEnd {
				resultState = sepState
			}
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if !childStartState.Moved(endState) {
			childErr = endState.NewSyntaxError("separated %s (endless loop because of empty result AND empty separator)", sd.parser.Expected())
			return endState, partRes.outs, childErr, partRes
		}
	}
}
