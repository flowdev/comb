package gomme

// Alternative tests a list of parsers in order, one by one, until one
// succeeds.
//
// If none of the parsers succeed, this combinator produces an error Result.
func Alternative[Output any](parsers ...Parser[Output]) Parser[Output] {
	return func(state State) Result[Output] {
		if len(parsers) == 0 {
			return Failure[Output](NewError(state, "Alternative(no parser given)"), state)
		}

		bestState := state
		for i, parse := range parsers {
			altState := state.Clean()
			result := parse(altState)
			if result.Err == nil {
				return result
			}

			// may the best error(s) win:
			if i == 0 {
				bestState = altState
			} else {
				bestState = bestState.Better(altState)
			}
		}

		err := &Error{Input: bestState}
		return Failure[Output](err, state)
	}
}
