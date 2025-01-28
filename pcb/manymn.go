package pcb

import (
	"github.com/oleiade/gomme"
)

// noSeparator is a parser that parses the empty string.
// It basically turns SeparatedMN into ManyMN.
func noSeparator[S gomme.Separator]() gomme.Parser[S] {
	p := func(state gomme.State) (gomme.State, S, *gomme.ParserError) {
		var zero S
		return state, zero, nil
	}
	return gomme.NewParser("noSeparator", p, Forbidden("noSeparator"))
}

// SeparatedMN applies an element parser and a separator parser repeatedly in order
// to produce a slice of elements.
//
// Because the `SeparatedListMN` is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed even if the separator parser fails to
// match at the end.
//
// The parser will fail if both parsers together accepted an empty input
// in order to prevent infinite loops.
func SeparatedMN[Output any, S gomme.Separator](
	parse gomme.Parser[Output], separator gomme.Parser[S],
	atLeast, atMost int,
	parseSeparatorAtEnd bool,
) gomme.Parser[[]Output] {
	if atLeast < 0 {
		panic("SeparatedMN is unable to handle negative `atLeast`")
	}
	if atMost < 0 {
		panic("SeparatedMN is unable to handle negative `atMost`")
	}

	sd := &separatedData[Output, S]{
		parse:               parse,
		separator:           separator,
		atLeast:             atLeast,
		atMost:              atMost,
		parseSeparatorAtEnd: parseSeparatorAtEnd,
	}
	return gomme.NewBranchParser[[]Output]("FirstSuccessful", sd.children, sd.parseAfterChild)
}

type separatedData[Output any, S gomme.Separator] struct {
	id                  uint64
	parse               gomme.Parser[Output]
	separator           gomme.Parser[S]
	atLeast             int
	atMost              int
	parseSeparatorAtEnd bool
}

func (sd *separatedData[Output, S]) children() []gomme.AnyParser {
	return []gomme.AnyParser{sd.parse, sd.separator}
}

func (sd *separatedData[Output, S]) parseAfterChild(childID int32, childResult gomme.ParseResult) gomme.ParseResult {
	gomme.Debugf("SeparatedMN.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childResult.Error != nil {
		return childResult // we can't avoid any errors by going another path
	}

	if childID >= 0 && childID != sd.parse.ID() && childID != sd.separator.ID() {
		childResult.Error = childResult.EndState.NewSemanticError(
			"unable to parse after child with unknown ID %d", childID)
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
			return gomme.ParseResult{StartState: state, EndState: remaining, Output: outputs, Error: nil}
		}

		nState, out, err := sd.parse.Parse(remaining)
		if err != nil {
			if remaining.SaveSpotMoved(nState) || count < sd.atLeast { // fail
				return gomme.ParseResult{StartState: remaining, EndState: nState, Output: out, Error: err}
			}
			return gomme.ParseResult{StartState: state, EndState: retState, Output: outputs, Error: nil}
		}
		outputs = append(outputs, out)
		count++

		retState = nState
		sepState, sepOut, err := sd.separator.Parse(nState)
		if err != nil {
			if nState.SaveSpotMoved(sepState) || count < sd.atLeast { // fail
				return gomme.ParseResult{StartState: nState, EndState: sepState, Output: sepOut, Error: err}
			}
			return gomme.ParseResult{StartState: state, EndState: retState, Output: outputs, Error: nil}
		}
		if sd.parseSeparatorAtEnd {
			retState = sepState
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if !remaining.Moved(sepState) {
			err = sepState.NewSyntaxError(
				"many %s (empty element incl. separator => endless loop)", sd.parse.Expected())
			return gomme.ParseResult{StartState: remaining, EndState: sepState, Output: outputs, Error: err}
		}
		remaining = sepState
	}
}
