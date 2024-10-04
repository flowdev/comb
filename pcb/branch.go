package pcb

import (
	"github.com/oleiade/gomme"
)

// NoWayBack applies a child parser and marks the state with NoWayBack if successful.
func NoWayBack[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	newParse := func(state gomme.State) (gomme.State, Output) {
		newState, output := parse.It(state)
		if newState.Failed() {
			return newState, output
		}

		return newState.SignalNoWayBack(), output
	}

	return gomme.NewParser[Output]("NoWayBack", parse.AvgConsumption, newParse)
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
	count := 0
	sum := 0

	avgConsumption := func() uint {
		if count == 0 {
			return gomme.DefaultConsumption
		}
		return uint((sum + count/2) / count)
	}

	newParse := func(state gomme.State) (gomme.State, Output) {
		bestState := state
		for i, parse := range parsers {
			newState, output := parse.It(state)
			if !newState.Failed() {
				count++
				sum += len(state.BytesTo(newState))
				return newState, output
			}
			failState := state.Failure(newState)
			if failState.NoWayBack() {
				return gomme.HandleAllErrors(failState, parse) // this will force it through
			}
			newState, output = gomme.HandleCurError(failState, parse)
			if !newState.Failed() {
				return newState, output
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

	return gomme.NewParser[Output]("Alternative", avgConsumption, newParse)
}
