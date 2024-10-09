package pcb

import (
	"github.com/oleiade/gomme"
)

// NoWayBack applies a child parser and marks the state with NoWayBack if successful.
// It tests the optional recoverer of the parser during the construction phase
// to get an early panic.
// This way we won't have a panic at the runtime of the parser.
func NoWayBack[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	// call Recoverer to make a Forbidden recoverer panic during the construction phase
	recoverer := parse.Recoverer()
	if recoverer != nil {
		recoverer(gomme.NewState(0, []byte{}))
	}

	newParse := func(state gomme.State) (gomme.State, Output) {
		newState, output := parse.It(state)
		if newState.Failed() {
			return newState, output
		}

		return newState.SignalNoWayBack(), output
	}

	return gomme.NewParser[Output]("NoWayBack", newParse, parse.Recoverer())
}

// FirstSuccessfulOf tests a list of parsers in order, one by one,
// until one succeeds.
// All parsers have to be of the same type.
//
// If no parser succeeds, this combinator produces an error Result.
func FirstSuccessfulOf[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[Output] {
	if len(parsers) == 0 {
		panic("pcb.FirstSuccessfulOf(missing parsers)")
	}

	newParse := func(state gomme.State) (gomme.State, Output) {
		bestState := state
		for i, parse := range parsers {
			newState, output := parse.It(state)
			if !newState.Failed() {
				return newState, output
			}
			failState := state.Failure(newState)
			if failState.NoWayBack() {
				return gomme.HandleAllErrors(failState, parse) // this will force it through
			}
			newState, output = gomme.HandleCurrentError(failState, parse)
			if !newState.Failed() {
				return newState, output
			}

			// may the best error win:
			if i == 0 {
				bestState = newState
			} else {
				bestState = gomme.BetterOf(bestState, newState)
			}
		}

		return state.Failure(bestState), gomme.ZeroOf[Output]()
	}

	return gomme.NewParser[Output]("FirstSuccessfulOf", newParse, nil)
}
