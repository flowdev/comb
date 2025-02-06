package pcb

import (
	"github.com/flowdev/comb"
)

// Optional applies an optional child parser. Will return a zero value
// if not successful.
// Optional will ignore any parsing error except if a SafeSpot is active.
func Optional[Output any](parser gomme.Parser[Output]) gomme.Parser[Output] {
	return gomme.NewBranchParser[Output](
		"Optional",
		func() []gomme.AnyParser {
			return []gomme.AnyParser{parser}
		}, func(childID int32, childResult gomme.ParseResult) gomme.ParseResult {
			var out Output
			gomme.Debugf("Optional.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())
			if childID >= 0 { // on the way up: Fetch
				var o interface{}
				o, childResult = childResult.FetchOutput()
				out, _ = o.(Output)
			}
			endResult := childResult
			if childID >= 0 {
				if childID != parser.ID() {
					childResult.Error = childResult.EndState.NewSemanticError(
						"unable to parse after child with unknown ID %d", childID)
					return childResult
				}
				out, _ = childResult.Output.(Output)
			} else {
				endResult = gomme.RunParser(parser, childResult)
				childResult.StartState = childResult.EndState
			}
			if endResult.Error != nil && childResult.StartState.SaveSpotMoved(endResult.EndState) { // we can't ignore the error
				return endResult.AddOutput(out)
			}
			if endResult.Error != nil { // successful result without input consumption
				endResult.EndState = endResult.StartState
				endResult.Error = nil
				return endResult.AddOutput(out)
			}
			return endResult.AddOutput(out)
		},
	)
}

// Peek tries to apply the provided parser without consuming any input.
// It effectively allows to look ahead in the input.
//
// NOTE:
//   - SafeSpot isn't honored here because we aren't officially parsing anything.
//   - Even though Peek accepts a parser as argument it behaves like a leaf parser
//     to the outside. There will be no error recovery as we don't parse anything.
func Peek[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	peekParse := func(state gomme.State) (gomme.State, Output, *gomme.ParserError) {
		_, out, err := parse.Parse(state)
		return state, out, gomme.ClaimError(err)
	}
	return gomme.NewParser[Output]("Peek", peekParse, Forbidden())
}

// Not tries to apply the provided parser without consuming any input.
// Not succeeds if the parser fails and succeeds if the parser fails.
// It effectively allows to look ahead in the input.
// An error returned should be handled (or ignored) by the parent parser.
//
// NOTE:
//   - SafeSpot isn't honored here because we aren't officially parsing anything.
//   - Even though Not accepts a parser as argument it behaves like a leaf parser
//     to the outside. There will be no error recovery as we don't parse anything.
//   - The returned boolean value indicates its own success and not the given parsers.
func Not[Output any](parser gomme.Parser[Output]) gomme.Parser[bool] {
	expected := "not " + parser.Expected()
	notParse := func(state gomme.State) (gomme.State, bool, *gomme.ParserError) {
		_, _, err := parser.Parse(state)
		if err != nil {
			return state, true, nil
		}
		return state, false, state.NewSyntaxError(expected)
	}
	return gomme.NewParser[bool](expected, notParse, Forbidden())
}

// Assign returns the provided value if the parser succeeds, otherwise
// it returns an error result.
func Assign[Output1, Output2 any](value Output1, parser gomme.Parser[Output2]) gomme.Parser[Output1] {
	return MapN[Output2, interface{}, interface{}, interface{}, interface{}](
		"Assign",
		parser, nil, nil, nil, nil,
		1,
		func(_ Output2) (Output1, error) {
			return value, nil
		}, nil, nil, nil, nil,
	)
}

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any](prefix gomme.Parser[OP], parser gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return MapN[OP, O, OS, interface{}, interface{}](
		"Delimited",
		prefix, parser, suffix, nil, nil, 3, nil, nil,
		func(output1 OP, output2 O, output3 OS) (O, error) {
			return output2, nil
		}, nil, nil)
}

// Prefixed parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
func Prefixed[OP, O any](prefix gomme.Parser[OP], parser gomme.Parser[O]) gomme.Parser[O] {
	return MapN[OP, O, interface{}, interface{}, interface{}](
		"Prefixed",
		prefix, parser, nil, nil, nil, 2, nil,
		func(output1 OP, output2 O) (O, error) {
			return output2, nil
		}, nil, nil, nil)
}

// Suffixed parses a result from the main parser, it then
// parses the result from the suffix parser and discards it; only
// returning the result of the main parser.
func Suffixed[O, OS any](parser gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return MapN[O, OS, interface{}, interface{}, interface{}](
		"Suffixed",
		parser, suffix, nil, nil, nil, 2, nil,
		func(output1 O, output2 OS) (O, error) {
			return output1, nil
		}, nil, nil, nil)
}

// Map applies a function to the successful result of 1 parser.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map[PO1 any, MO any](parse gomme.Parser[PO1], fn func(PO1) (MO, error)) gomme.Parser[MO] {
	return MapN[PO1, interface{}, interface{}, interface{}, interface{}](
		"Map",
		parse, nil, nil, nil, nil, 1, fn, nil, nil, nil, nil)
}

// Map2 applies a function to the successful result of 2 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map2[PO1, PO2 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], fn func(PO1, PO2) (MO, error),
) gomme.Parser[MO] {
	return MapN[PO1, PO2, interface{}, interface{}, interface{}](
		"Map2",
		parse1, parse2, nil, nil, nil, 2, nil, fn, nil, nil, nil)
}

// Map3 applies a function to the successful result of 3 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map3[PO1, PO2, PO3 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3],
	fn func(PO1, PO2, PO3) (MO, error),
) gomme.Parser[MO] {
	return MapN[PO1, PO2, PO3, interface{}, interface{}](
		"Map3",
		parse1, parse2, parse3, nil, nil, 3, nil, nil, fn, nil, nil)
}

// Map4 applies a function to the successful result of 4 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map4[PO1, PO2, PO3, PO4 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3], parse4 gomme.Parser[PO4],
	fn func(PO1, PO2, PO3, PO4) (MO, error),
) gomme.Parser[MO] {
	return MapN[PO1, PO2, PO3, PO4, interface{}](
		"Map4",
		parse1, parse2, parse3, parse4, nil, 4, nil, nil, nil, fn, nil)
}

// Map5 applies a function to the successful result of 5 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map5[PO1, PO2, PO3, PO4, PO5 any, MO any](
	parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3], parse4 gomme.Parser[PO4], parse5 gomme.Parser[PO5],
	fn func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) gomme.Parser[MO] {
	return MapN("Map5", parse1, parse2, parse3, parse4, parse5, 5, nil, nil, nil, nil, fn)
}
