// Package pcb contains all the standard parsers and all recoverers and deleters.
// Everything in this package could be done by a user of the project, too.
// So you are welcome to copy something and adapt it to your needs. 😀
package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
)

// EOF parses the end of the input.
// If there is still input left to parse, an error result is returned.
// This IS already a `NoWayBack` parser.
func EOF() gomme.Parser[interface{}] {
	expected := "end of the input"

	parse := func(state gomme.State) (gomme.State, interface{}) {
		remaining := state.BytesRemaining()
		if remaining > 0 {
			return state.NewError(fmt.Sprintf("%s (still %d bytes of input left)", expected, remaining)),
				nil
		}

		return state, nil
	}

	return gomme.NoWayBack(
		gomme.NewParser[interface{}](expected, parse, false, func(state gomme.State) int {
			return state.BytesRemaining()
		}, nil),
	)
}
