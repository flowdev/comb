package parsify

import (
	"fmt"
	"github.com/oleiade/gomme"
	"strconv"
	"unicode/utf8"
)

// Parser is a simple function here.
// An interface wouldn't work at all.
type Parser[Output any] func(gomme.State) (gomme.State, Output, *gomme.ParserError)

// Parserish types are any type that can be turned into a Parser by Parsify
type Parserish[Output any] interface {
	~rune | ~func(gomme.State) (gomme.State, Output)
}

// Parsify turns p of type Parserish into a real Parser.
func Parsify[Output any, Parsish Parserish[Output]](p Parsish) Parser[Output] {
	var zeroOutput Output
	ip := interface{}(p) // convert p to an interface so we can type switch over it

	// ip: interface p; ap: asserted p; op: Output p
	switch ap := ip.(type) {
	case Parser[Output]:
		return ap
	case rune:
		if op, ok := ip.(Output); ok { // Output == rune
			iruneErr := interface{}(utf8.RuneError)
			oruneErr, _ := iruneErr.(Output)
			expected := strconv.QuoteRune(ap)
			return func(state gomme.State) (gomme.State, Output, *gomme.ParserError) {
				r, size := utf8.DecodeRuneInString(state.CurrentString())
				if r == utf8.RuneError {
					if size == 0 {
						return state, oruneErr, state.NewSyntaxError("%q (at EOF)", expected)
					}
					return state, oruneErr, state.NewSyntaxError("%q (got UTF-8 error)", expected)
				}
				if r != ap {
					return state, oruneErr, state.NewSyntaxError("%q (got %q)", expected, r)
				}

				return state.MoveBy(size), op, nil
			}
		}
		panic(fmt.Errorf("can't turn a rune into a parser of type `%T`", zeroOutput))
	default:
		panic(fmt.Errorf("can't turn a `%T` into a parser of type `%T`", p, zeroOutput))
	}
}
