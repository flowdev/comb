package parsify

import (
	"fmt"
	"github.com/oleiade/gomme"
	"unicode/utf8"
)

// Parserish types are any type that can be turned into a Parser by Parsify
type Parserish[Output any] interface {
	~rune | ~func(gomme.State) (gomme.State, Output)
}

// Parsify turns p of type Parserish into a real Parser.
func Parsify[Output any, Parsish Parserish[Output]](p Parsish) gomme.Parser[Output] {
	var zeroOutput Output
	ip := interface{}(p) // convert p to an interface so we can type switch over it

	// ip: interface p; ap: asserted p; op: Output p
	switch ap := ip.(type) {
	case gomme.Parser[Output]:
		return ap
	case rune:
		if op, ok := ip.(Output); ok { // Output == rune
			iruneErr := interface{}(utf8.RuneError)
			oruneErr, _ := iruneErr.(Output)
			return func(state gomme.State) (gomme.State, Output) {
				r, size := utf8.DecodeRune(state.CurrentBytes())
				if r == utf8.RuneError {
					if size == 0 {
						return state.AddError(fmt.Sprintf("%q (at EOF)", ap)), oruneErr
					}
					return state.AddError(fmt.Sprintf("%q (got UTF-8 error)", ap)), oruneErr
				}
				if r != ap {
					return state.AddError(fmt.Sprintf("%q (got %q)", ap, r)), oruneErr
				}

				return state.MoveBy(uint(size)), op
			}
		}
		panic(fmt.Errorf("can't turn a rune into a parser of type `%T`", zeroOutput))
	default:
		panic(fmt.Errorf("can't turn a `%T` into a parser of type `%T`", p, zeroOutput))
	}
}
