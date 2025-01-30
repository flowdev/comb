package pcb

import (
	"bytes"
	"fmt"
	"github.com/oleiade/gomme"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Char parses a single rune and matches it with
// a provided candidate.
// If the rune could not be found at the current position,
// the parser returns an error result.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func Char(char rune) gomme.Parser[rune] {
	expected := strconv.QuoteRune(char)

	parse := func(state gomme.State) (gomme.State, rune, *gomme.ParserError) {
		r, size := utf8.DecodeRuneInString(state.CurrentString())
		if r == utf8.RuneError {
			if size == 0 {
				return state, utf8.RuneError, state.NewSyntaxError("%s (at EOF)", expected)
			}
			return state, utf8.RuneError, state.NewSyntaxError("%s (got UTF-8 error)", expected)
		}
		if r != char {
			return state, utf8.RuneError, state.NewSyntaxError("%s (got %q)", expected, r)
		}

		return state.MoveBy(size), r, nil
	}

	return gomme.NewParser[rune](expected, parse, IndexOf(char))
}

// Byte parses a single byte and matches it with
// a provided candidate.
// If the byte could not be found at the current position,
// the parser returns an error result.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func Byte(byt byte) gomme.Parser[byte] {
	expected := "0x" + strconv.FormatUint(uint64(byt), 16)

	parse := func(state gomme.State) (gomme.State, byte, *gomme.ParserError) {
		buf := state.CurrentBytes()
		if len(buf) == 0 {
			return state, 0, state.NewSyntaxError("%s (at EOF)", expected)
		}
		b := buf[0]
		if b != byt {
			return state, 0, state.NewSyntaxError("%s (got 0x%x)", expected, b)
		}

		return state.MoveBy(1), b, nil
	}

	return gomme.NewParser[byte](expected, parse, IndexOf(byt))
}

// Satisfy parses a single character, and ensures that it satisfies the given predicate.
// `expected` is used in error messages to tell the user what is expected at the current position.
//
// This parser is a good candidate for SafeSpot and has an optimized Recoverer.
// An even more specialized Recoverer can be used later with `parser.SwapRecoverer(newRecoverer)`.
func Satisfy(expected string, predicate func(rune) bool) gomme.Parser[rune] {
	parse := func(state gomme.State) (gomme.State, rune, *gomme.ParserError) {
		r, size := utf8.DecodeRuneInString(state.CurrentString())
		if r == utf8.RuneError {
			if size == 0 {
				return state, utf8.RuneError, state.NewSyntaxError("%s (at EOF)", expected)
			}
			return state, utf8.RuneError, state.NewSyntaxError("%s (got UTF-8 error)", expected)
		}
		if !predicate(r) {
			return state, utf8.RuneError, state.NewSyntaxError("%s (got %q)", expected, r)
		}

		return state.MoveBy(size), r, nil
	}

	recoverer := func(state gomme.State) int {
		return strings.IndexFunc(state.CurrentString(), predicate)
	}

	return gomme.NewParser[rune](expected, parse, recoverer)
}

// String parses a token from the input, and returns the part of the input that
// matched the token.
// If the token could not be found at the current position,
// the parser returns an error result.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func String(token string) gomme.Parser[string] {
	expected := strconv.Quote(token)

	parse := func(state gomme.State) (gomme.State, string, *gomme.ParserError) {
		if !strings.HasPrefix(state.CurrentString(), token) {
			return state, "", state.NewSyntaxError(expected)
		}

		newState := state.MoveBy(len(token))
		return newState, token, nil
	}

	return gomme.NewParser[string](expected, parse, IndexOf(token))
}

// Bytes parses a token from the input, and returns the part of the input that
// matched the token.
// If the token could not be found at the current position,
// the parser returns an error result.
func Bytes(token []byte) gomme.Parser[[]byte] {
	expected := fmt.Sprintf("0x%x", token)

	parse := func(state gomme.State) (gomme.State, []byte, *gomme.ParserError) {
		if !bytes.HasPrefix(state.CurrentBytes(), token) {
			return state, []byte{}, state.NewSyntaxError(expected)
		}

		newState := state.MoveBy(len(token))
		return newState, token, nil
	}

	return gomme.NewParser[[]byte](expected, parse, IndexOf(token))
}

// UntilString parses until it finds a token in the input, and returns
// the part of the input that preceded the token.
// If found the parser moves beyond the stop string.
// If the token could not be found, the parser returns an error result.
//
// NOTE:
//   - This function panics if `stop` is empty.
//   - UntilString is rather dangerous especially in case of error recovery
//     because it potentially consumes much more input than expected.
//     In error cases it will usually start earlier because other parsers are skipped.
//     Especially using it as a `SafeSpot` parser is a bad idea!
func UntilString(stop string) gomme.Parser[string] {
	expected := fmt.Sprintf("... %q", stop)

	if stop == "" {
		panic("stop is empty")
	}

	parse := func(state gomme.State) (gomme.State, string, *gomme.ParserError) {
		input := state.CurrentString()
		i := strings.Index(input, stop)
		if i == -1 {
			return state, "", state.NewSyntaxError(expected)
		}

		newState := state.MoveBy(i + len(stop))
		return newState, input[:i], nil
	}

	return gomme.NewParser[string](
		expected,
		parse,
		func(state gomme.State) int {
			if strings.Contains(state.CurrentString(), stop) {
				return 0 // this is probably not what the user wants but the only correct value :(
			}
			return -1
		},
	)
}

// SatisfyMN returns the longest input subset that matches the predicate,
// within the boundaries of `atLeast` <= number of runes found <= `atMost`.
//
// If the provided parser is not successful or the predicate doesn't match
// `atLeast` times, the parser fails and goes back to the start.
//
// This parser is a good candidate for SafeSpot and has an optimized Recoverer.
// An even more specialized Recoverer can be used later with `parser.SwapRecoverer(newRecoverer) Parser`.
func SatisfyMN(expected string, atLeast, atMost int, predicate func(rune) bool) gomme.Parser[string] {
	if atLeast < 0 {
		panic("SatisfyMN is unable to handle negative `atLeast` argument")
	}
	if atMost < 0 {
		panic("SatisfyMN is unable to handle negative `atMost` argument")
	}
	parse := func(state gomme.State) (gomme.State, string, *gomme.ParserError) {
		current := state
		count := 0
		for atMost > count {
			r, size := utf8.DecodeRuneInString(current.CurrentString())
			if r == utf8.RuneError {
				if count >= atLeast {
					output := state.StringTo(current)
					return current, output, nil
				}
				if size == 0 {
					return state, "", state.NewSyntaxError("%s (need %d, found %d at EOF)",
						expected, atLeast, count)
				}
				return state, "", state.NewSyntaxError("%s (need %d, found %d, got UTF-8 error)",
					expected, atLeast, count)
			}

			if !predicate(r) {
				if count >= atLeast {
					output := state.StringTo(current)
					return current, output, nil
				}
				return state, "", state.NewSyntaxError("%s (need %d, found %d, got %q)",
					expected, atLeast, count, r)
			}

			current = current.MoveBy(size)
			count++
		}

		output := state.StringTo(current)
		return current, output, nil
	}

	return gomme.NewParser[string](expected, parse, satisfyMNRecoverer(atLeast, predicate))
}

func satisfyMNRecoverer(atLeast int, predicate func(rune) bool) gomme.Recoverer {
	return func(state gomme.State) int {
		count := 0
		start := 0
		for i, r := range state.CurrentString() {
			if predicate(r) {
				if count == 0 {
					start = i
				}
				count++
				if count >= atLeast {
					return start
				}
			} else {
				count = 0
			}
		}
		return -1
	}
}

// AlphaMN parses at least `atLeast` and at most `atMost` Unicode letters.
func AlphaMN(atLeast, atMost int) gomme.Parser[string] {
	return SatisfyMN("letter", atLeast, atMost, unicode.IsLetter)
}

// Alpha0 parses a zero or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input is empty, or no character is found, the parser
// returns the input as is.
func Alpha0() gomme.Parser[string] {
	return SatisfyMN("letter", 0, math.MaxInt, unicode.IsLetter)
}

// Alpha1 parses one or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alpha1() gomme.Parser[string] {
	return SatisfyMN("letter", 1, math.MaxInt, unicode.IsLetter)
}

// Alphanumeric0 parses zero or more alphabetical or numerical Unicode characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Alphanumeric0() gomme.Parser[string] {
	return SatisfyMN("letter or numeral", 0, math.MaxInt, IsAlphanumeric)
}

// Alphanumeric1 parses one or more alphabetical or numerical Unicode characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alphanumeric1() gomme.Parser[string] {
	return SatisfyMN("letter or numeral", 1, math.MaxInt, IsAlphanumeric)
}

// Digit0 parses zero or more ASCII numerical characters: 0-9.
// In the cases where the input is empty, or no digit character is found, the parser
// returns the input as is.
func Digit0() gomme.Parser[string] {
	return SatisfyMN("digit", 0, math.MaxInt, IsDigit)
}

// Digit1 parses one or more numerical characters: 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Digit1() gomme.Parser[string] {
	return SatisfyMN("digit", 1, math.MaxInt, IsDigit)
}

// HexDigit0 parses zero or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input is empty, or no terminating character is found, the parser
// returns the input as is.
func HexDigit0() gomme.Parser[string] {
	return SatisfyMN("hexadecimal digit", 0, math.MaxInt, IsHexDigit)
}

// HexDigit1 parses one or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func HexDigit1() gomme.Parser[string] {
	return SatisfyMN("hexadecimal digit", 1, math.MaxInt, IsHexDigit)
}

// Whitespace0 parses zero or more Unicode whitespace characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Whitespace0() gomme.Parser[string] {
	return SatisfyMN("whitespace", 0, math.MaxInt, unicode.IsSpace)
}

// Whitespace1 parses one or more Unicode whitespace characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Whitespace1() gomme.Parser[string] {
	return SatisfyMN("whitespace", 1, math.MaxInt, unicode.IsSpace)
}

// OneOfRunes parses a single character from the given set of characters.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func OneOfRunes(collection ...rune) gomme.Parser[rune] {
	n := len(collection)
	if n == 0 {
		panic("OneOfRunes has no characters to match")
	}
	expected := fmt.Sprintf("one of %q", collection)

	parser := Satisfy(expected, func(r rune) bool {
		return slices.Contains(collection, r)
	})
	parser.SwapRecoverer(func(state gomme.State) int {
		return strings.IndexAny(state.CurrentString(), string(collection))
	})
	return parser
}

// OneOf parses a single character from the given set of characters.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func OneOf(collection ...string) gomme.Parser[string] {
	n := len(collection)
	if n == 0 {
		panic("OneOf has no characters to match")
	}
	expected := fmt.Sprintf("one of %q", collection)

	parse := func(state gomme.State) (gomme.State, string, *gomme.ParserError) {
		input := state.CurrentString()
		for _, token := range collection {
			if strings.HasPrefix(input, token) {
				return state.MoveBy(len(token)), token, nil
			}
		}

		return state, "", state.NewSyntaxError(expected)
	}

	return gomme.NewParser[string](expected, parse, IndexOfAny(collection...))
}

// LF parses a line feed `\n` character.
func LF() gomme.Parser[rune] {
	return Char('\n')
}

// CR parses a carriage return `\r` character.
func CR() gomme.Parser[rune] {
	return Char('\r')
}

// CRLF parses the string `\r\n`.
func CRLF() gomme.Parser[string] {
	return String("\r\n")
}

// Space parses an ASCII space character (' ').
func Space() gomme.Parser[rune] {
	return Char(' ')
}

// Tab parses an ASCII tab character ('\t').
func Tab() gomme.Parser[rune] {
	return Char('\t')
}

// IsAlphanumeric returns true if the rune is a Unicode letter,
// a Unicode number or '_'.
func IsAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

// IsDigit returns true if the rune is a digit.
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsHexDigit returns true if the rune is a hexadecimal digit.
func IsHexDigit(r rune) bool {
	return IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
