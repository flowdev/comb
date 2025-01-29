package pcb

import (
	"github.com/oleiade/gomme"
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
func SeparatedMN[Output any, S gomme.Separator](
	parser gomme.Parser[Output], separator gomme.Parser[S],
	atLeast, atMost int,
	parseSeparatorAtEnd bool,
) gomme.Parser[[]Output] {
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
	return gomme.NewBranchParser[[]Output](expected, sd.children, sd.parseAfterChild)
}

type separatedData[Output any, S gomme.Separator] struct {
	id                  uint64
	parser              gomme.Parser[Output]
	separator           gomme.Parser[S]
	atLeast             int
	atMost              int
	parseSeparatorAtEnd bool
}

func (sd *separatedData[Output, S]) children() []gomme.AnyParser {
	if sd.separator == nil {
		return []gomme.AnyParser{sd.parser}
	}
	return []gomme.AnyParser{sd.parser, sd.separator}
}

func (sd *separatedData[Output, S]) parseAfterChild(childID int32, childResult gomme.ParseResult) gomme.ParseResult {
	var zeroSep S

	gomme.Debugf("SeparatedMN.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childResult.Error != nil {
		return childResult // we can't avoid any errors by going another path
	}

	if childID >= 0 && childID != sd.parser.ID() && childID != sd.separator.ID() {
		childResult.Error = childResult.EndState.NewSemanticError(
			"unable to parser after child with unknown ID %d", childID)
		return childResult
	}

	state := childResult.StartState
	remaining := childResult.EndState
	retState := remaining
	outputs := make([]Output, 0, min(32, sd.atMost))
	count := 0
	if childID < 0 {
		state = childResult.EndState
	} else {
		out, _ := childResult.Output.(Output)
		outputs = append(outputs, out)
		count = 1
	}

	for {
		if count >= sd.atMost {
			return gomme.ParseResult{StartState: state, EndState: retState, Output: outputs, Error: nil}
		}

		nState, out, err := sd.parser.Parse(remaining)
		if err != nil {
			if remaining.SaveSpotMoved(nState) || count < sd.atLeast { // fail
				return gomme.ParseResult{StartState: remaining, EndState: nState, Output: out, Error: err}
			}
			return gomme.ParseResult{StartState: state, EndState: retState, Output: outputs, Error: nil}
		}
		outputs = append(outputs, out)
		count++
		retState = nState

		sepState, sepOut := nState, zeroSep
		if sd.separator != nil {
			sepState, sepOut, err = sd.separator.Parse(nState)
			if err != nil {
				if nState.SaveSpotMoved(sepState) || count < sd.atLeast { // fail
					return gomme.ParseResult{StartState: nState, EndState: sepState, Output: sepOut, Error: err}
				}
				return gomme.ParseResult{StartState: state, EndState: retState, Output: outputs, Error: nil}
			}
			if sd.parseSeparatorAtEnd {
				retState = sepState
			}
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if !remaining.Moved(sepState) {
			err = sepState.NewSyntaxError(
				"many %s (empty element incl. separator => endless loop)", sd.parser.Expected())
			return gomme.ParseResult{StartState: remaining, EndState: sepState, Output: outputs, Error: err}
		}
		remaining = sepState
	}
}
