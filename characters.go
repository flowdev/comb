package gomme

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Char parses a single character and matches it with
// a provided candidate.
func Char(char rune) Parser[rune] {
	return func(state State) (State, rune) {
		r, size := utf8.DecodeRune(state.CurrentBytes())
		if r == utf8.RuneError {
			if size == 0 {
				return state.AddError(fmt.Sprintf("%q (at EOF)", char)), utf8.RuneError
			}
			return state.AddError(fmt.Sprintf("%q (got UTF-8 error)", char)), utf8.RuneError
		}
		if r != char {
			return state.AddError(fmt.Sprintf("%q (got %q)", char, r)), utf8.RuneError
		}

		return state.MoveBy(uint(size)), r
	}
}

// Satisfy parses a single character, and ensures that it satisfies the given predicate.
func Satisfy(predicate func(rune) bool) Parser[rune] {
	return func(state State) (State, rune) {
		r, size := utf8.DecodeRune(state.CurrentBytes())
		if r == utf8.RuneError {
			if size == 0 {
				return state.AddError(fmt.Sprintf("%s (at EOF)", "Satisfy")), utf8.RuneError
			}
			return state.AddError(fmt.Sprintf("%s (got UTF-8 error)", "Satisfy")), utf8.RuneError
		}
		if !predicate(r) {
			return state.AddError(fmt.Sprintf("%s (got %q)", "Satisfy", r)), utf8.RuneError
		}

		return state.MoveBy(uint(size)), r
	}
}

// String parses a token from the input, and returns the part of the input that
// matched the token.
// If the token could not be found, the parser returns an error result.
func String(token string) Parser[string] {
	return func(state State) (State, string) {
		if !strings.HasPrefix(state.CurrentString(), token) {
			return state.AddError(fmt.Sprintf("%q", token)), ""
		}

		newState := state.MoveBy(uint(len(token)))
		return newState, state.StringTo(newState)
	}
}

// SatisfyMN returns the longest input subset that matches the predicate,
// within the boundaries of `atLeast` <= number of runes found <= `atMost`.
//
// If the provided parser is not successful or the predicate doesn't match
// `atLeast` times, the parser fails and goes back to the start.
func SatisfyMN(atLeast, atMost uint, predicate func(rune) bool) Parser[string] {
	return func(state State) (State, string) {
		current := state
		count := uint(0)
		for atMost > count {
			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if count >= atLeast {
					return current, state.StringTo(current)
				}
				if size == 0 {
					return state.Failure(current.AddError(
						fmt.Sprintf("<character> (need %d, found %d at EOF)", atLeast, count),
					)), ""
				}
				return state.Failure(current.AddError(
					fmt.Sprintf("<character> (need %d, found %d, got UTF-8 error)", atLeast, count),
				)), ""
			}

			if !predicate(r) {
				if count >= atLeast {
					return current, state.StringTo(current)
				}
				return state.Failure(current.AddError(
					fmt.Sprintf("<character> (need %d, found %d, got %q)", atLeast, count, r),
				)), ""
			}

			current = current.MoveBy(uint(size))
		}

		return current, state.StringTo(current)
	}
}

// AlphaMN parses at least `atLeast` and at most `atMost` Unicode letters.
func AlphaMN(atLeast, atMost uint) Parser[string] {
	return SatisfyMN(atLeast, atMost, unicode.IsLetter)
}

// Alpha0 parses a zero or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input is empty, or no character is found, the parser
// returns the input as is.
func Alpha0() Parser[string] {
	return SatisfyMN(0, math.MaxUint, unicode.IsLetter)
}

// Alpha1 parses one or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alpha1() Parser[string] {
	return SatisfyMN(1, math.MaxUint, unicode.IsLetter)
}

// Alphanumeric0 parses zero or more alphabetical or numerical Unicode characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Alphanumeric0() Parser[string] {
	return SatisfyMN(0, math.MaxUint, IsAlphanumeric)
}

// Alphanumeric1 parses one or more alphabetical or numerical Unicode characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alphanumeric1() Parser[string] {
	return SatisfyMN(1, math.MaxUint, IsAlphanumeric)
}

// Digit0 parses zero or more ASCII numerical characters: 0-9.
// In the cases where the input is empty, or no digit character is found, the parser
// returns the input as is.
func Digit0() Parser[string] {
	return SatisfyMN(0, math.MaxUint, IsDigit)
}

// Digit1 parses one or more numerical characters: 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Digit1() Parser[string] {
	return SatisfyMN(1, math.MaxUint, IsDigit)
}

// HexDigit0 parses zero or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input is empty, or no terminating character is found, the parser
// returns the input as is.
func HexDigit0() Parser[string] {
	return SatisfyMN(0, math.MaxUint, IsHexDigit)
}

// HexDigit1 parses one or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func HexDigit1() Parser[string] {
	return SatisfyMN(1, math.MaxUint, IsHexDigit)
}

// Whitespace0 parses zero or more Unicode whitespace characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Whitespace0() Parser[string] {
	return SatisfyMN(0, math.MaxUint, unicode.IsSpace)
}

// Whitespace1 parses one or more Unicode whitespace characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Whitespace1() Parser[string] {
	return SatisfyMN(1, math.MaxUint, unicode.IsSpace)
}

// LF parses a line feed `\n` character.
func LF() Parser[rune] {
	return Char('\n')
}

// CR parses a carriage return `\r` character.
func CR() Parser[rune] {
	return Char('\r')
}

// CRLF parses the string `\r\n`.
func CRLF() Parser[string] {
	return String("\r\n")
}

// OneOf parses a single character from the given set of characters.
func OneOf(collection ...rune) Parser[rune] {
	return Satisfy(func(r rune) bool {
		for _, c := range collection {
			if r == c {
				return true
			}
		}

		return false
	})
}

// Space parses an ASCII space character (' ').
func Space() Parser[rune] {
	return Char(' ')
}

// Tab parses an ASCII tab character ('\t').
func Tab() Parser[rune] {
	return Char('\t')
}

// Int64 parses an integer from the input, and returns it plus the remaining input.
func Int64() Parser[int64] {
	return Map2(Optional(OneOf('-', '+')), Digit1(), func(optSign rune, digits string) (int64, error) {
		i, err := strconv.ParseInt(digits, 10, 64)
		if err != nil {
			return 0, err
		}
		if optSign == '-' {
			i = -i
		}
		return i, nil
	})
}

// Int8 parses an 8-bit integer from the input,
// and returns the part of the input that matched the integer.
func Int8() Parser[int8] {
	return Map2(Optional(OneOf('-', '+')), Digit1(), func(optSign rune, digits string) (int8, error) {
		i, err := strconv.ParseInt(digits, 10, 8)
		if err != nil {
			return 0, err
		}
		if optSign == '-' {
			i = -i
		}
		return int8(i), nil
	})
}

// UInt8 parses an 8-bit integer from the input,
// and returns the part of the input that matched the integer.
func UInt8() Parser[uint8] {
	return Map2(Optional(Char('+')), Digit1(), func(optSign rune, digits string) (uint8, error) {
		ui, err := strconv.ParseUint(digits, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(ui), nil
	})
}

// IsAlphanumeric returns true if the rune is an alphanumeric character.
func IsAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// IsDigit returns true if the rune is a digit.
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsHexDigit returns true if the rune is a hexadecimal digit.
func IsHexDigit(r rune) bool {
	return IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
