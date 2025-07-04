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
// in order to prevent infinite loops.
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
	return comb.NewBranchParser[[]Output](expected, sd.children, sd.parseAfterChild)
}

type separatedData[Output any, S comb.Separator] struct {
	id                  uint64
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
	_ *comb.ParserError, childID int32, childResult comb.ParseResult,
) comb.ParseResult {
	var partRes partialSepResult[Output]

	comb.Debugf("SeparatedMN.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childID >= 0 { // on the way up: Fetch
		var o interface{}
		o, childResult = childResult.FetchOutput()
		partRes, _ = o.(partialSepResult[Output])
	} else {
		partRes.outs = make([]Output, 0, min(32, sd.atMost))
	}

	if childResult.Error != nil {
		if sd.atLeast > 0 || childResult.StartState.SafeSpotMoved(childResult.EndState) { // fail
			return childResult.AddOutput(partRes)
		}
		childResult.Error = nil // ignore error: we have enough output
		childResult.Output = partRes.outs
		return childResult.AddOutput(partRes)
	}

	if childID >= 0 && childID != sd.parser.ID() && childID != sd.separator.ID() {
		childResult.Error = childResult.EndState.NewSemanticError(
			"unable to parser after child with unknown ID %d", childID)
		return childResult.AddOutput(partRes)
	}

	endResult := childResult
	count := len(partRes.outs)
	if childID < 0 {
		childResult.StartState = childResult.EndState
	} else if childID == sd.parser.ID() {
		out, _ := childResult.Output.(Output)
		partRes.outs = append(partRes.outs, out)
		count++
	}

	for {
		if count >= sd.atMost {
			endResult.Output = partRes.outs
			return endResult.AddOutput(partRes)
		}

		endResult = comb.RunParser(sd.parser, childResult)
		if endResult.Error != nil {
			if sd.atLeast > count || childResult.EndState.SafeSpotMoved(endResult.EndState) { // fail
				return endResult.AddOutput(partRes)
			}
			endResult.Error = nil // ignore error: we have enough output
			endResult.Output = partRes.outs
			return endResult.AddOutput(partRes)
		}
		out, _ := endResult.Output.(Output)
		partRes.outs = append(partRes.outs, out)
		count++

		sepResult := endResult
		if sd.separator != nil {
			sepResult = comb.RunParser(sd.separator, endResult)
			if sepResult.Error != nil {
				if sd.atLeast > count || endResult.EndState.SafeSpotMoved(sepResult.EndState) { // fail
					return sepResult.AddOutput(partRes)
				}
				endResult.Error = nil // ignore error: we have enough output
				endResult.Output = partRes.outs
				return endResult.AddOutput(partRes)
			}
			if sd.parseSeparatorAtEnd {
				endResult = sepResult
			}
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if !childResult.EndState.Moved(sepResult.EndState) {
			sepResult.Error = sepResult.EndState.NewSyntaxError(
				"many %s (empty element incl. separator => endless loop)", sd.parser.Expected())
			sepResult.Output = partRes.outs
			return sepResult.AddOutput(partRes)
		}
		childResult = sepResult
	}
}
