package pcb

import (
	"github.com/oleiade/gomme"
	"strings"
)

// Optional applies an optional child parser. Will return nil
// if not successful.
//
// NOTE:
// Optional will ignore any parsing error.
func Optional[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	optParse := func(state gomme.State) (gomme.State, Output) {
		newState, output := parse.It(state)
		if newState.Failed() {
			if newState.NoWayBack() {
				return gomme.HandleAllErrors(state.Failure(newState), parse) // this will force it through
			}
			return state.Success(newState), gomme.ZeroOf[Output]()
		}
		return newState, output
	}

	return gomme.NewParser[Output](
		"Optional",
		optParse,
		Forbidden("Optional"),
		parse.ContainsRefuge(),
		parse.RefugeRecoverer,
	)
}

// Peek tries to apply the provided parser without consuming any input.
// It effectively allows to look ahead in the input.
// NoWayBack isn't honored here because we aren't officially parsing anything.
func Peek[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	peekParse := func(state gomme.State) (gomme.State, Output) {
		newState, output := parse.It(state)
		if newState.Failed() {
			// avoid NoWayBack because we only peek; error message and consumption don't matter anyway
			return state.NewError("Peek"), output
		}

		return state, output
	}
	return gomme.NewParser[Output]("Peek", peekParse, Forbidden("Peek"), gomme.TernaryNo, nil)
}

// Not tries to apply the provided parser without consuming any input.
// Not succeeds if the parser fails and succeeds if the parser fails.
// It effectively allows to look ahead in the input.
// The returned boolean value indicates its own success and not the given parsers.
// NoWayBack isn't honored here because we aren't officially parsing anything.
func Not[Output any](parse gomme.Parser[Output]) gomme.Parser[bool] {
	expected := "not " + parse.Expected()
	parseNot := func(state gomme.State) (gomme.State, bool) {
		newState, _ := parse.It(state)
		if newState.Failed() {
			return state, true
		}

		// avoid NoWayBack because we only peek; error message and consumption don't matter either
		return state.NewError(expected), false
	}
	return gomme.NewParser[bool](expected, parseNot, Forbidden("Not"), gomme.TernaryNo, nil)
}

// Recognize returns the consumed input (instead of the original parsers output)
// as the produced value when the provided parser succeeds.
//
// Note:
// Using this parser is a code smell as it effectively removes type safety.
// Rather use one of the MapX functions instead.
func Recognize[Output any](parse gomme.Parser[Output]) gomme.Parser[[]byte] {
	recParse := func(state gomme.State) (gomme.State, []byte) {
		newState, _ := parse.It(state)
		if newState.Failed() {
			newState, _ = gomme.HandleCurrentError(state.Failure(newState), parse)
			if newState.Failed() {
				return state.Failure(newState), []byte{}
			}
		}

		return newState, state.BytesTo(newState)
	}
	return gomme.NewParser[[]byte](
		"Recognize",
		recParse,
		parse.MyRecoverer(),
		parse.ContainsRefuge(),
		parse.RefugeRecoverer,
	)
}

// Assign returns the provided value if the parser succeeds, otherwise
// it returns an error result.
func Assign[Output1, Output2 any](value Output1, parse gomme.Parser[Output2]) gomme.Parser[Output1] {
	asgnParse := func(state gomme.State) (gomme.State, Output1) {
		newState, _ := parse.It(state)
		if newState.Failed() {
			newState, _ = gomme.HandleCurrentError(state.Failure(newState), parse)
			if newState.Failed() {
				return state.Failure(newState), gomme.ZeroOf[Output1]()
			}
		}

		return newState, value
	}
	return gomme.NewParser[Output1](
		parse.Expected(),
		asgnParse,
		parse.MyRecoverer(),
		parse.ContainsRefuge(),
		parse.RefugeRecoverer,
	)
}

// Map applies a function to the successful result of 1 parser.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map[PO1 any, MO any](parse gomme.Parser[PO1], fn func(PO1) (MO, error)) gomme.Parser[MO] {
	mapParse := func(state gomme.State) (gomme.State, MO) {
		newState, output := parse.It(state)
		if newState.Failed() {
			newState, output = gomme.HandleCurrentError(state.Failure(newState), parse)
			if newState.Failed() {
				return state.Failure(newState), gomme.ZeroOf[MO]()
			}
		}

		mapped, err := fn(output)
		if err != nil {
			return state.NewSemanticError(err.Error()), gomme.ZeroOf[MO]()
		}

		return newState, mapped
	}

	return gomme.NewParser[MO](
		parse.Expected(),
		mapParse,
		parse.MyRecoverer(),
		parse.ContainsRefuge(),
		parse.RefugeRecoverer,
	)
}

// Map2 applies a function to the successful result of 2 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map2[PO1, PO2 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], fn func(PO1, PO2) (MO, error),
) gomme.Parser[MO] {
	expected := strings.Builder{}
	expected.WriteString(parse1.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse2.Expected())

	containsNoWayBack := max(parse1.ContainsRefuge(), parse2.ContainsRefuge())

	// Construct myRefugeRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, 0, 2)
	if parse1.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse1.RefugeRecoverer)
	}
	if parse2.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse2.RefugeRecoverer)
	}
	myRefugeRecoverer := gomme.NewCombiningRecoverer(subRecoverers...)

	mapParse := func(state gomme.State) (gomme.State, MO) {
		newState1, output1 := parse1.It(state)
		if newState1.Failed() {
			newState1, output1 = gomme.HandleCurrentError(state.Failure(newState1), parse1)
			if newState1.Failed() {
				return state.Failure(newState1), gomme.ZeroOf[MO]()
			}
		}

		newState2, output2 := parse2.It(newState1)
		if newState2.Failed() {
			newState2, output2 = gomme.HandleCurrentError(state.Failure(newState2), parse2)
			if newState2.Failed() {
				return state.Failure(newState2), gomme.ZeroOf[MO]()
			}
		}

		mapped, err := fn(output1, output2)
		if err != nil {
			return state.NewSemanticError(err.Error()), gomme.ZeroOf[MO]()
		}

		return newState2, mapped
	}

	return gomme.NewParser[MO](
		expected.String(),
		mapParse,
		BasicRecovererFunc(mapParse),
		containsNoWayBack,
		myRefugeRecoverer.Recover,
	)
}

// Map3 applies a function to the successful result of 3 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map3[PO1, PO2, PO3 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3],
	fn func(PO1, PO2, PO3) (MO, error),
) gomme.Parser[MO] {
	expected := strings.Builder{}
	expected.WriteString(parse1.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse2.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse3.Expected())

	containsNoWayBack := max(parse1.ContainsRefuge(), parse2.ContainsRefuge(), parse3.ContainsRefuge())

	// Construct myRefugeRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, 0, 3)
	if parse1.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse1.RefugeRecoverer)
	}
	if parse2.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse2.RefugeRecoverer)
	}
	if parse3.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse3.RefugeRecoverer)
	}
	myRefugeRecoverer := gomme.NewCombiningRecoverer(subRecoverers...)

	mapParse := func(state gomme.State) (gomme.State, MO) {
		newState1, output1 := parse1.It(state)
		if newState1.Failed() {
			newState1, output1 = gomme.HandleCurrentError(state.Failure(newState1), parse1)
			if newState1.Failed() {
				return state.Failure(newState1), gomme.ZeroOf[MO]()
			}
		}

		newState2, output2 := parse2.It(newState1)
		if newState2.Failed() {
			newState2, output2 = gomme.HandleCurrentError(newState1.Failure(newState2), parse2)
			if newState2.Failed() {
				return state.Failure(newState2), gomme.ZeroOf[MO]()
			}
		}

		newState3, output3 := parse3.It(newState2)
		if newState3.Failed() {
			newState3, output3 = gomme.HandleCurrentError(newState2.Failure(newState3), parse3)
			if newState3.Failed() {
				return state.Failure(newState3), gomme.ZeroOf[MO]()
			}
		}

		mapped, err := fn(output1, output2, output3)
		if err != nil {
			return state.NewSemanticError(err.Error()), gomme.ZeroOf[MO]()
		}

		return newState3, mapped
	}

	return gomme.NewParser[MO](
		expected.String(),
		mapParse,
		BasicRecovererFunc(mapParse),
		containsNoWayBack,
		myRefugeRecoverer.Recover,
	)
}

// Map4 applies a function to the successful result of 4 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map4[PO1, PO2, PO3, PO4 any, MO any](parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3], parse4 gomme.Parser[PO4],
	fn func(PO1, PO2, PO3, PO4) (MO, error),
) gomme.Parser[MO] {
	expected := strings.Builder{}
	expected.WriteString(parse1.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse2.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse3.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse4.Expected())

	containsNoWayBack := max(parse1.ContainsRefuge(), parse2.ContainsRefuge(),
		parse3.ContainsRefuge(), parse4.ContainsRefuge())

	// Construct myRefugeRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, 0, 3)
	if parse1.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse1.RefugeRecoverer)
	}
	if parse2.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse2.RefugeRecoverer)
	}
	if parse3.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse3.RefugeRecoverer)
	}
	if parse4.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse4.RefugeRecoverer)
	}
	myRefugeRecoverer := gomme.NewCombiningRecoverer(subRecoverers...)

	mapParse := func(state gomme.State) (gomme.State, MO) {
		newState1, output1 := parse1.It(state)
		if newState1.Failed() {
			newState1, output1 = gomme.HandleCurrentError(state.Failure(newState1), parse1)
			if newState1.Failed() {
				return state.Failure(newState1), gomme.ZeroOf[MO]()
			}
		}

		newState2, output2 := parse2.It(newState1)
		if newState2.Failed() {
			newState2, output2 = gomme.HandleCurrentError(newState1.Failure(newState2), parse2)
			if newState2.Failed() {
				return state.Failure(newState2), gomme.ZeroOf[MO]()
			}
		}

		newState3, output3 := parse3.It(newState2)
		if newState3.Failed() {
			newState3, output3 = gomme.HandleCurrentError(newState2.Failure(newState3), parse3)
			if newState3.Failed() {
				return state.Failure(newState3), gomme.ZeroOf[MO]()
			}
		}

		newState4, output4 := parse4.It(newState3)
		if newState4.Failed() {
			newState4, output4 = gomme.HandleCurrentError(newState3.Failure(newState4), parse4)
			if newState4.Failed() {
				return state.Failure(newState4), gomme.ZeroOf[MO]()
			}
		}

		mapped, err := fn(output1, output2, output3, output4)
		if err != nil {
			return state.NewSemanticError(err.Error()), gomme.ZeroOf[MO]()
		}

		return newState4, mapped
	}

	return gomme.NewParser[MO](
		expected.String(),
		mapParse,
		BasicRecovererFunc(mapParse),
		containsNoWayBack,
		myRefugeRecoverer.Recover,
	)
}

// Map5 applies a function to the successful result of 5 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map5[PO1, PO2, PO3, PO4, PO5 any, MO any](
	parse1 gomme.Parser[PO1], parse2 gomme.Parser[PO2], parse3 gomme.Parser[PO3], parse4 gomme.Parser[PO4], parse5 gomme.Parser[PO5],
	fn func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) gomme.Parser[MO] {
	expected := strings.Builder{}
	expected.WriteString(parse1.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse2.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse3.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse4.Expected())
	expected.WriteString(" + ")
	expected.WriteString(parse5.Expected())

	containsNoWayBack := max(parse1.ContainsRefuge(), parse2.ContainsRefuge(), parse3.ContainsRefuge(),
		parse4.ContainsRefuge(), parse5.ContainsRefuge())

	// Construct myRefugeRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, 0, 3)
	if parse1.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse1.RefugeRecoverer)
	}
	if parse2.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse2.RefugeRecoverer)
	}
	if parse3.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse3.RefugeRecoverer)
	}
	if parse4.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse4.RefugeRecoverer)
	}
	if parse5.ContainsRefuge() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, parse5.RefugeRecoverer)
	}
	myRefugeRecoverer := gomme.NewCombiningRecoverer(subRecoverers...)

	mapParse := func(state gomme.State) (gomme.State, MO) {
		newState1, output1 := parse1.It(state)
		if newState1.Failed() {
			newState1, output1 = gomme.HandleCurrentError(state.Failure(newState1), parse1)
			if newState1.Failed() {
				return state.Failure(newState1), gomme.ZeroOf[MO]()
			}
		}

		newState2, output2 := parse2.It(newState1)
		if newState2.Failed() {
			newState2, output2 = gomme.HandleCurrentError(newState1.Failure(newState2), parse2)
			if newState2.Failed() {
				return state.Failure(newState2), gomme.ZeroOf[MO]()
			}
		}

		newState3, output3 := parse3.It(newState2)
		if newState3.Failed() {
			newState3, output3 = gomme.HandleCurrentError(newState2.Failure(newState3), parse3)
			if newState3.Failed() {
				return state.Failure(newState3), gomme.ZeroOf[MO]()
			}
		}

		newState4, output4 := parse4.It(newState3)
		if newState4.Failed() {
			newState4, output4 = gomme.HandleCurrentError(newState3.Failure(newState4), parse4)
			if newState4.Failed() {
				return state.Failure(newState4), gomme.ZeroOf[MO]()
			}
		}

		newState5, output5 := parse5.It(newState4)
		if newState5.Failed() {
			newState5, output5 = gomme.HandleCurrentError(newState4.Failure(newState5), parse5)
			if newState5.Failed() {
				return state.Failure(newState5), gomme.ZeroOf[MO]()
			}
		}

		mapped, err := fn(output1, output2, output3, output4, output5)
		if err != nil {
			return state.NewSemanticError(err.Error()), gomme.ZeroOf[MO]()
		}

		return newState5, mapped
	}

	return gomme.NewParser[MO](
		expected.String(),
		mapParse,
		BasicRecovererFunc(mapParse),
		containsNoWayBack,
		myRefugeRecoverer.Recover,
	)
}
