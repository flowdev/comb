package gomme

// Alternative tests a list of parsers in order, one by one, until one
// succeeds.
//
// If none of the parsers succeed, this combinator produces an error Result.
func Alternative[Output any](parsers ...Parser[Output]) Parser[Output] {
	if len(parsers) == 0 {
		panic("Alternative(missing parsers)")
	}
	return func(state State) (State, Output) {
		var zeroOutput Output

		bestState := state
		for i, parse := range parsers {
			newState, output := parse(state)
			if !newState.Failed() {
				return newState, output
			}

			// may the best error(s) win:
			if i == 0 {
				bestState = newState
			} else {
				bestState = BetterOf(bestState, newState)
			}
		}

		return state.Failure(bestState), zeroOutput
	}
}
