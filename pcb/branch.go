package pcb

import "github.com/oleiade/gomme"

// NoWayBack applies a child parser and marks the state with NoWayBack if successful.
func NoWayBack[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	return func(state gomme.State) (gomme.State, Output) {
		newState, output := parse(state)
		if newState.Failed() {
			return newState, output
		}

		return newState.SignalNoWayBack(), output
	}
}

// Alternative tests a list of parsers in order, one by one,
// until one succeeds.
// All parsers have to be of the same type.
//
// If no parser succeeds, this combinator produces an error Result.
func Alternative[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[Output] {
	if len(parsers) == 0 {
		panic("Alternative(missing parsers)")
	}

	return func(state gomme.State) (gomme.State, Output) {
		bestState := state
		for i, parse := range parsers {
			newState, output := parse(state)
			if !newState.Failed() {
				return newState, output
			}
			if newState.NoWayBack() {
				return state.Failure(newState), gomme.ZeroOf[Output]()
			}

			// may the best error(s) win:
			if i == 0 {
				bestState = newState
			} else {
				bestState = gomme.BetterOf(bestState, newState)
			}
		}

		return state.Failure(bestState), gomme.ZeroOf[Output]()
	}
}
