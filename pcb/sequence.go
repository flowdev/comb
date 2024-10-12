package pcb

import "github.com/oleiade/gomme"

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any](prefix gomme.Parser[OP], parse gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return Map3(prefix, parse, suffix, func(output1 OP, output2 O, output3 OS) (O, error) {
		return output2, nil
	})
}

// Preceded parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
func Preceded[OP, O any](prefix gomme.Parser[OP], parse gomme.Parser[O]) gomme.Parser[O] {
	return Map2(prefix, parse, func(output1 OP, output2 O) (O, error) {
		return output2, nil
	})
}

// Sequence applies a sequence of parsers of the same type and
// returns either a slice of results or an error if any parser fails.
// Use one of the MapX parsers for differently typed parsers.
func Sequence[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[[]Output] {
	containsNoWayBack := parsers[0].ContainsRefuge()
	for i := 1; i < len(parsers); i++ {
		containsNoWayBack = max(containsNoWayBack, parsers[i].ContainsRefuge())
	}

	// Construct myRefugeRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, len(parsers))
	for i, parser := range parsers {
		if parser.ContainsRefuge() > gomme.TernaryNo {
			subRecoverers[i] = parser.RefugeRecoverer
		}
	}
	myRefugeRecoverer := gomme.NewCombiningRecoverer(subRecoverers...)

	parseSeq := func(state gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, len(parsers))
		remaining := state
		for _, parse := range parsers {
			newState, output := parse.It(remaining)
			if newState.Failed() {
				newState, output = gomme.HandleCurrentError(remaining.Failure(newState), parse)
				if newState.Failed() {
					return state.Failure(newState), gomme.ZeroOf[[]Output]()
				}
			}

			outputs = append(outputs, output)
			remaining = newState
		}

		return remaining, outputs
	}

	return gomme.NewParser[[]Output](
		"Sequence",
		parseSeq,
		BasicRecovererFunc(parseSeq),
		containsNoWayBack,
		myRefugeRecoverer.Recover,
	)
}

// Terminated parses a result from the main parser, it then
// parses the result from the suffix parser and discards it; only
// returning the result of the main parser.
func Terminated[O, OS any](parse gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return Map2(parse, suffix, func(output1 O, output2 OS) (O, error) {
		return output1, nil
	})
}
