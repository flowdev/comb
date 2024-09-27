package gomme

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any](prefix Parser[OP], parse Parser[O], suffix Parser[OS]) Parser[O] {
	return Map3(prefix, parse, suffix, func(output1 OP, output2 O, output3 OS) (O, error) {
		return output2, nil
	})
}

// Preceded parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
func Preceded[OP, O any](prefix Parser[OP], parse Parser[O]) Parser[O] {
	return Map2(prefix, parse, func(output1 OP, output2 O) (O, error) {
		return output2, nil
	})
}

// Sequence applies a sequence of parsers and returns either a
// slice of results or an error if any parser fails.
// All parsers in the sequence have to produce the same result type.
// Use one of the Map* parsers for different result types.
func Sequence[O any](parsers ...Parser[O]) Parser[[]O] {
	return func(state State) (State, []O) {
		outputs := make([]O, 0, len(parsers))
		remaining := state
		for _, parse := range parsers {
			newState, output := parse(remaining)
			if newState.Failed() {
				return state.Failure(newState), ZeroOf[[]O]()
			}

			outputs = append(outputs, output)
			remaining = newState
		}

		return remaining, outputs
	}
}

// Terminated parses a result from the main parser, it then
// parses the result from the suffix parser and discards it; only
// returning the result of the main parser.
func Terminated[O, OS any](parse Parser[O], suffix Parser[OS]) Parser[O] {
	return Map2(parse, suffix, func(output1 O, output2 OS) (O, error) {
		return output1, nil
	})
}
