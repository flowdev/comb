package cmb

import (
	"github.com/flowdev/comb"
)

// Optional applies an optional child parser. Will return a zero value
// if not successful.
// Optional will ignore any parsing error except if a SafeSpot is active.
func Optional[Output any](parser comb.Parser[Output]) comb.Parser[Output] {
	var p comb.Parser[Output]

	p = comb.NewBranchParser[Output](
		"Optional",
		func() []comb.AnyParser {
			return []comb.AnyParser{parser}
		}, func(
			childID int32,
			childStartState, childState comb.State,
			childOut interface{},
			childErr *comb.ParserError,
			data interface{},
		) (comb.State, Output, *comb.ParserError, interface{}) {
			var out Output
			comb.Debugf("Optional.parseAfterChild - childID=%d, pos=%d", childID, childState.CurrentPos())
			if childID >= 0 { // bottom-up
				out, _ = data.(Output)
			} else { // top-down
				childStartState = childState
				childState, childOut, childErr = parser.ParseAny(p.ID(), childStartState)
				out, _ = childOut.(Output)
			}
			if childErr != nil && childStartState.SafeSpotMoved(childState) { // we can't ignore the error
				return childState, out, childErr, out
			}
			if childErr != nil { // successful result without input consumption
				return childStartState, out, nil, nil
			}
			return childState, out, nil, nil
		},
	)
	return p
}

// Peek tries to apply the provided parser without consuming any input.
// It effectively allows looking ahead in the input.
//
// NOTE:
//   - SafeSpot isn't honored here because we aren't officially parsing anything.
//   - Even though Peek accepts a parser as an argument, it behaves like a leaf parser
//     to the outside world. There will be no error recovery as we don't parse anything.
func Peek[Output any](parse comb.Parser[Output]) comb.Parser[Output] {
	var p comb.Parser[Output]
	peekParse := func(state comb.State) (comb.State, Output, *comb.ParserError) {
		_, aOut, err := parse.ParseAny(comb.ParentUnknown, state)
		out, _ := aOut.(Output)
		return state, out, comb.ClaimError(err)
	}
	p = comb.NewParser[Output]("Peek", peekParse, Forbidden())
	return p
}

// Not tries to apply the provided parser without consuming any input.
// Not succeeds if the parser fails and succeeds if the parser fails.
// It effectively allows looking ahead in the input.
// An error returned should be handled (or ignored) by the parent parser.
//
// NOTE:
//   - SafeSpot isn't honored here because we aren't officially parsing anything.
//   - Even though Not accepts a parser as an argument, it behaves like a leaf parser
//     to the outside world. There will be no error recovery as we don't parse anything.
//   - The returned boolean value indicates its own success and not the given parsers.
func Not[Output any](parser comb.Parser[Output]) comb.Parser[bool] {
	var p comb.Parser[bool]

	expected := "not " + parser.Expected()
	notParse := func(state comb.State) (comb.State, bool, *comb.ParserError) {
		_, _, err := parser.ParseAny(comb.ParentUnknown, state)
		if err != nil {
			return state, true, nil
		}
		return state, false, state.NewSyntaxError(expected)
	}
	p = comb.NewParser[bool](expected, notParse, Forbidden())
	return p
}

// Assign returns the provided value if the parser succeeds, otherwise
// it returns an error result.
func Assign[Output1, Output2 any](value Output1, parser comb.Parser[Output2]) comb.Parser[Output1] {
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
func Delimited[OP, O, OS any](prefix comb.Parser[OP], parser comb.Parser[O], suffix comb.Parser[OS]) comb.Parser[O] {
	return MapN[OP, O, OS, interface{}, interface{}](
		"Delimited",
		prefix, parser, suffix, nil, nil, 3, nil, nil,
		func(output1 OP, output2 O, output3 OS) (O, error) {
			return output2, nil
		}, nil, nil)
}

// Prefixed parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
func Prefixed[OP, O any](prefix comb.Parser[OP], parser comb.Parser[O]) comb.Parser[O] {
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
func Suffixed[O, OS any](parser comb.Parser[O], suffix comb.Parser[OS]) comb.Parser[O] {
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
func Map[PO1 any, MO any](parse comb.Parser[PO1], fn func(PO1) (MO, error)) comb.Parser[MO] {
	return MapN[PO1, interface{}, interface{}, interface{}, interface{}](
		"Map",
		parse, nil, nil, nil, nil, 1, fn, nil, nil, nil, nil)
}

// Map2 applies a function to the successful result of 2 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map2[PO1, PO2 any, MO any](parse1 comb.Parser[PO1], parse2 comb.Parser[PO2], fn func(PO1, PO2) (MO, error),
) comb.Parser[MO] {
	return MapN[PO1, PO2, interface{}, interface{}, interface{}](
		"Map2",
		parse1, parse2, nil, nil, nil, 2, nil, fn, nil, nil, nil)
}

// Map3 applies a function to the successful result of 3 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map3[PO1, PO2, PO3 any, MO any](parse1 comb.Parser[PO1], parse2 comb.Parser[PO2], parse3 comb.Parser[PO3],
	fn func(PO1, PO2, PO3) (MO, error),
) comb.Parser[MO] {
	return MapN[PO1, PO2, PO3, interface{}, interface{}](
		"Map3",
		parse1, parse2, parse3, nil, nil, 3, nil, nil, fn, nil, nil)
}

// Map4 applies a function to the successful result of 4 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map4[PO1, PO2, PO3, PO4 any, MO any](parse1 comb.Parser[PO1], parse2 comb.Parser[PO2], parse3 comb.Parser[PO3], parse4 comb.Parser[PO4],
	fn func(PO1, PO2, PO3, PO4) (MO, error),
) comb.Parser[MO] {
	return MapN[PO1, PO2, PO3, PO4, interface{}](
		"Map4",
		parse1, parse2, parse3, parse4, nil, 4, nil, nil, nil, fn, nil)
}

// Map5 applies a function to the successful result of 5 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map5[PO1, PO2, PO3, PO4, PO5 any, MO any](
	parse1 comb.Parser[PO1], parse2 comb.Parser[PO2], parse3 comb.Parser[PO3], parse4 comb.Parser[PO4], parse5 comb.Parser[PO5],
	fn func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) comb.Parser[MO] {
	return MapN("Map5", parse1, parse2, parse3, parse4, parse5, 5, nil, nil, nil, nil, fn)
}
