// Package cmb contains all the standard parsers and all recoverers.
// Everything in this package could be done by a user of the project, too.
// So you are welcome to copy something and adapt it to your needs. 😀
package cmb

import (
	"github.com/flowdev/comb"
)

// EOF parses the end of the input.
// If there is still input left to parse, an error is returned.
// This IS already a `SafeSpot` parser (its recoverer consumes the rest of the input).
func EOF() comb.Parser[interface{}] {
	var p comb.Parser[interface{}]

	expected := "end of the input"

	parse := func(state comb.State) (comb.State, interface{}, *comb.ParserError) {
		remaining := state.BytesRemaining()
		if remaining > 0 {
			return state, nil, state.NewSyntaxError(p.ID(), "%s (still %d bytes of input left)", expected, remaining)
		}

		return state, nil, nil
	}

	p = comb.SafeSpot(
		comb.NewParser[interface{}](expected, parse, func(_ *comb.ParserError, state comb.State) int {
			return state.BytesRemaining()
		}),
	)
	return p
}
