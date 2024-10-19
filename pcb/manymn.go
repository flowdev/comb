package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
)

// ManyMN applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output.
//
// Note that ManyMN fails if the provided parser accepts empty inputs (such as
// `Digit0`, or `Alpha0`) in order to prevent infinite loops.
func ManyMN[Output any](parse gomme.Parser[Output], atLeast, atMost int) gomme.Parser[[]Output] {
	id := gomme.NewBranchParserID()

	if atLeast < 0 {
		panic("ManyMN is unable to handle negative `atLeast` argument")
	}
	if atMost < 0 {
		panic("ManyMN is unable to handle negative `atMost` argument")
	}

	parseMany := func(state gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, min(32, atMost))
		remaining := state
		count := 0
		for {
			if count >= atMost {
				return remaining, outputs
			}
			newState, output := parse.It(remaining)
			if newState.Failed() && newState.NoWayBack() {
				// TODO: handle error!!!
			} else if newState.Failed() {
				if count < atLeast {
					// TODO: Add more error handling!
					return gomme.IWitnessed(state, id, count, newState), []Output{}
				} else {
					return remaining, outputs
				}
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if !newState.Moved(remaining) {
				return state.NewError(fmt.Sprintf("%s (found empty element)", parse.Expected())), []Output{}
			}

			outputs = append(outputs, output)
			remaining = newState
			count++
		}
	}

	recoverer := Forbidden("Many(atLeast=0)")
	containsNoWayBack := gomme.TernaryNo
	if atLeast > 0 {
		recoverer = BasicRecovererFunc(parseMany)
		containsNoWayBack = parse.ContainsNoWayBack()
	}
	return gomme.NewParser[[]Output]("ManyMN", parseMany, true, recoverer,
		containsNoWayBack, parse.NoWayBackRecoverer)
}
