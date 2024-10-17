package parsify

import (
	"fmt"
	"github.com/oleiade/gomme"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any, ParserOP Parserish[OP], ParserO Parserish[O], ParserOS Parserish[OS]](
	prefix ParserOP, parse ParserO, suffix ParserOS) Parser[O] {
	return Map3(prefix, parse, suffix, func(output1 OP, output2 O, output3 OS) (O, error) {
		return output2, nil
	})
}

// Map3 applies a function to the successful result of 3 parsers.
// Arbitrary complex data structures can be built with Map and Map2 alone.
// The other MapX parsers are provided for convenience.
func Map3[PO1, PO2, PO3, MO any, ParserPO1 Parserish[PO1], ParserPO2 Parserish[PO2], ParserPO3 Parserish[PO3]](
	parse1 ParserPO1, parse2 ParserPO2, parse3 ParserPO3, fn func(PO1, PO2, PO3) (MO, error),
) Parser[MO] {
	pparse1 := Parsify[PO1, ParserPO1](parse1)
	pparse2 := Parsify[PO2, ParserPO2](parse2)
	pparse3 := Parsify[PO3, ParserPO3](parse3)

	return func(state gomme.State) (gomme.State, MO) {
		newState1, output1 := pparse1(state)
		if newState1.Failed() {
			return state.Preserve(newState1), gomme.ZeroOf[MO]()
		}

		newState2, output2 := pparse2(newState1)
		if newState2.Failed() {
			return state.Preserve(newState2), gomme.ZeroOf[MO]()
		}

		newState3, output3 := pparse3(newState2)
		if newState3.Failed() {
			return state.Preserve(newState3), gomme.ZeroOf[MO]()
		}

		mapped, err := fn(output1, output2, output3)
		if err != nil {
			return state.NewError(err.Error()), gomme.ZeroOf[MO]()
		}

		return newState3, mapped
	}
}

// Char parses a single character and matches it with
// a provided candidate.
func Char(char rune) Parser[rune] {
	expected := strconv.QuoteRune(char)

	return func(state gomme.State) (gomme.State, rune) {
		r, size := utf8.DecodeRune(state.CurrentBytes())
		if r == utf8.RuneError {
			if size == 0 {
				return state.NewError(fmt.Sprintf("%q (at EOF)", expected)), utf8.RuneError
			}
			return state.NewError(fmt.Sprintf("%q (got UTF-8 error)", expected)), utf8.RuneError
		}
		if r != char {
			return state.NewError(fmt.Sprintf("%q (got %q)", expected, r)), utf8.RuneError
		}

		return state.MoveBy(size), r
	}
}

// Char2 parses a single character and matches it with
// a provided candidate.
func Char2[Output rune](char rune) Parser[Output] {
	expected := strconv.QuoteRune(char)

	return func(state gomme.State) (gomme.State, Output) {
		r, size := utf8.DecodeRune(state.CurrentBytes())
		if r == utf8.RuneError {
			if size == 0 {
				return state.NewError(fmt.Sprintf("%q (at EOF)", expected)), utf8.RuneError
			}
			return state.NewError(fmt.Sprintf("%q (got UTF-8 error)", expected)), utf8.RuneError
		}
		if r != char {
			return state.NewError(fmt.Sprintf("%q (got %q)", expected, r)), utf8.RuneError
		}

		return state.MoveBy(size), Output(r)
	}
}

// UntilString parses until it finds a token in the input, and returns
// the part of the input that preceded the token.
// If found the parser moves beyond the stop string.
// If the token could not be found, the parser returns an error result.
func UntilString(stop string) Parser[string] {
	return func(state gomme.State) (gomme.State, string) {
		input := state.CurrentString()
		i := strings.Index(input, stop)
		if i == -1 {
			return state.NewError(fmt.Sprintf("... %q", stop)), ""
		}

		newState := state.MoveBy(i + len(stop))
		return newState, input[:i]
	}
}
