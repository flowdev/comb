package pcb

import (
	"github.com/oleiade/gomme"
)

// Optional applies an optional child parser. Will return a zero value
// if not successful.
// Optional will ignore any parsing error except if a NoWayBack is active.
func Optional[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	optParse := func(state gomme.State) (gomme.State, Output) {
		newState, output := parse.It(state)
		if newState.Failed() && !state.NoWayBackMoved(newState) {
			return state.Succeed(newState), gomme.ZeroOf[Output]()
		}
		return newState, output
	}
	optParser := gomme.NewParser[Output](
		"Optional",
		optParse,
		true,
		Forbidden("Optional"),
		parse.ContainsNoWayBack(),
		parse.NoWayBackRecoverer,
	)
	return MapN[Output, interface{}, interface{}, interface{}, interface{}](
		optParser, nil, nil, nil, nil, 1, func(o Output) (Output, error) {
			return o, nil
		}, nil, nil, nil, nil)
}

// Peek tries to apply the provided parser without consuming any input.
// It effectively allows to look ahead in the input.
//
// NOTE:
//   - NoWayBack isn't honored here because we aren't officially parsing anything.
//   - Even though Peek accepts a parser as argument it behaves like a leaf parser
//     to the outside. So it doesn't need to use MapN or the like.
func Peek[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	peekParse := func(state gomme.State) (gomme.State, Output) {
		newState, output := parse.It(state)
		if newState.Failed() {
			// avoid NoWayBack because we only peek; error message and consumption don't matter anyway
			return state.NewError("Peek"), output
		}

		return state, output
	}
	return gomme.NewParser[Output]("Peek", peekParse, false, Forbidden("Peek"), gomme.TernaryNo, nil)
}

// Not tries to apply the provided parser without consuming any input.
// Not succeeds if the parser fails and succeeds if the parser fails.
// It effectively allows to look ahead in the input.
// The returned boolean value indicates its own success and not the given parsers.
//
// NOTE:
//   - NoWayBack isn't honored here because we aren't officially parsing anything.
//   - Even though Not accepts a parser as argument it behaves like a leaf parser
//     to the outside. So it doesn't need to use MapN or the like.
func Not[Output any](parse gomme.Parser[Output]) gomme.Parser[bool] {
	expected := "not " + parse.Expected()
	notParse := func(state gomme.State) (gomme.State, bool) {
		newState, _ := parse.It(state)
		if newState.Failed() {
			return state, true
		}

		// avoid NoWayBack because we only peek; error message and consumption don't matter either
		return state.NewError(expected), false
	}
	return gomme.NewParser[bool](expected, notParse, false, Forbidden("Not"),
		gomme.TernaryNo, nil)
}

// Recognize returns the consumed input (instead of the original parsers output)
// as the produced value when the provided parser succeeds.
//
// Note:
//   - Using this parser is a code smell as it effectively removes type safety.
//   - Rather use one of the MapX functions instead.
func Recognize[Output any](parse gomme.Parser[Output]) gomme.Parser[[]byte] {
	recParse := func(state gomme.State) (gomme.State, []byte) {
		newState, _ := parse.It(state)
		if newState.Failed() {
			return state.Preserve(newState), nil
		}
		return newState, state.BytesTo(newState)
	}
	recParser := gomme.NewParser[[]byte](
		"Recognize",
		recParse,
		parse.PossibleWitness(),
		parse.MyRecoverer(),
		parse.ContainsNoWayBack(),
		parse.NoWayBackRecoverer,
	)
	return MapN[[]byte, interface{}, interface{}, interface{}, interface{}](recParser, nil, nil, nil, nil, 1, func(bs []byte) ([]byte, error) {
		return bs, nil
	}, nil, nil, nil, nil)
}

// Assign returns the provided value if the parser succeeds, otherwise
// it returns an error result.
func Assign[Output1, Output2 any](value Output1, parse gomme.Parser[Output2]) gomme.Parser[Output1] {
	return MapN[Output2, interface{}, interface{}, interface{}, interface{}](
		parse, nil, nil, nil, nil, 1, func(_ Output2) (Output1, error) {
			return value, nil
		}, nil, nil, nil, nil)
}

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any](prefix gomme.Parser[OP], parse gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return MapN[OP, O, OS, interface{}, interface{}](
		prefix, parse, suffix, nil, nil, 3, nil, nil,
		func(output1 OP, output2 O, output3 OS) (O, error) {
			return output2, nil
		}, nil, nil)
}

// Prefixed parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
func Prefixed[OP, O any](prefix gomme.Parser[OP], parse gomme.Parser[O]) gomme.Parser[O] {
	return MapN[OP, O, interface{}, interface{}, interface{}](
		prefix, parse, nil, nil, nil, 2, nil,
		func(output1 OP, output2 O) (O, error) {
			return output2, nil
		}, nil, nil, nil)
}

// Suffixed parses a result from the main parser, it then
// parses the result from the suffix parser and discards it; only
// returning the result of the main parser.
func Suffixed[O, OS any](parse gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return MapN[O, OS, interface{}, interface{}, interface{}](
		parse, suffix, nil, nil, nil, 2, nil,
		func(output1 O, output2 OS) (O, error) {
			return output1, nil
		}, nil, nil, nil)
}

// Map applies a function to the successful result of 1 parser.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map[PO1 any, MO any](parse gomme.Parser[PO1], fn func(PO1) (MO, error)) gomme.Parser[MO] {
	return MapN[PO1, interface{}, interface{}, interface{}, interface{}](
		parse, nil, nil, nil, nil, 1, fn, nil, nil, nil, nil)
}

// Map2 applies a function to the successful result of 2 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map2[PO1, PO2 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], fn func(PO1, PO2) (MO, error),
) gomme.Parser[MO] {
	return MapN[PO1, PO2, interface{}, interface{}, interface{}](
		parse1, parse2, nil, nil, nil, 2, nil, fn, nil, nil, nil)
}

// Map3 applies a function to the successful result of 3 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map3[PO1, PO2, PO3 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3],
	fn func(PO1, PO2, PO3) (MO, error),
) gomme.Parser[MO] {
	return MapN[PO1, PO2, PO3, interface{}, interface{}](
		parse1, parse2, parse3, nil, nil, 3, nil, nil, fn, nil, nil)
}

// Map4 applies a function to the successful result of 4 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map4[PO1, PO2, PO3, PO4 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3], parse4 gomme.Parser[PO4],
	fn func(PO1, PO2, PO3, PO4) (MO, error),
) gomme.Parser[MO] {
	return MapN[PO1, PO2, PO3, PO4, interface{}](
		parse1, parse2, parse3, parse4, nil, 4, nil, nil, nil, fn, nil)
}

// Map5 applies a function to the successful result of 5 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map5[PO1, PO2, PO3, PO4, PO5 any, MO any](
	parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3], parse4 gomme.Parser[PO4], parse5 gomme.Parser[PO5],
	fn func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) gomme.Parser[MO] {
	return MapN(parse1, parse2, parse3, parse4, parse5, 5, nil, nil, nil, nil, fn)
}
