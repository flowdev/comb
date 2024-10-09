package pcb

import (
	"bytes"
	"fmt"
	"github.com/oleiade/gomme"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Char parses a single rune and matches it with
// a provided candidate.
// If the rune could not be found at the current position,
// the parser returns an error result.
func Char(char rune) gomme.Parser[rune] {
	expected := strconv.QuoteRune(char)
	consumption := uint(utf8.RuneCountInString(string(char)))

	parse := func(state gomme.State) (gomme.State, rune) {
		r, size := utf8.DecodeRune(state.CurrentBytes())
		if r == utf8.RuneError {
			if size == 0 {
				return state.NewError(fmt.Sprintf("%s (at EOF)", expected), state, consumption), utf8.RuneError
			}
			return state.NewError(fmt.Sprintf("%s (got UTF-8 error)", expected), state, consumption), utf8.RuneError
		}
		if r != char {
			return state.NewError(fmt.Sprintf("%s (got %q)", expected, r), state, consumption), utf8.RuneError
		}

		return state.MoveBy(uint(size)), r
	}

	return gomme.NewParser[rune](expected, gomme.ConstantConsumption(consumption), parse)
}

// Byte parses a single byte and matches it with
// a provided candidate.
// If the byte could not be found at the current position,
// the parser returns an error result.
func Byte(char byte) gomme.Parser[byte] {
	expected := "0x" + strconv.FormatUint(uint64(char), 16)
	consumption := uint(utf8.RuneCountInString(string(char)))

	parse := func(state gomme.State) (gomme.State, byte) {
		buf := state.CurrentBytes()
		if len(buf) == 0 {
			return state.NewError(fmt.Sprintf("%s (at EOF)", expected), state, consumption), 0
		}
		b := buf[0]
		if b != char {
			return state.NewError(fmt.Sprintf("%s (got 0x%x)", expected, b), state, consumption), 0
		}

		return state.MoveBy(1), b
	}

	return gomme.NewParser[byte](expected, gomme.ConstantConsumption(consumption), parse)
}

// Satisfy parses a single character, and ensures that it satisfies the given predicate.
// `expected` is used in error messages to tell the user what is expected at the current position.
func Satisfy(expected string, predicate func(rune) bool) gomme.Parser[rune] {
	count := 0
	sum := 0

	avgConsumption := func() uint {
		if count == 0 {
			return 2 // one rune can never be 5 bytes long
		}
		return uint((sum + count/2) / count)
	}

	parse := func(state gomme.State) (gomme.State, rune) {
		r, size := utf8.DecodeRune(state.CurrentBytes())
		if r == utf8.RuneError {
			if size == 0 {
				return state.NewError(fmt.Sprintf("%s (at EOF)", expected), state, avgConsumption()), utf8.RuneError
			}
			return state.NewError(fmt.Sprintf("%s (got UTF-8 error)", expected), state, avgConsumption()), utf8.RuneError
		}
		if !predicate(r) {
			return state.NewError(fmt.Sprintf("%s (got %q)", expected, r), state, avgConsumption()), utf8.RuneError
		}

		sum += size
		count++
		return state.MoveBy(uint(size)), r
	}

	return gomme.NewParser[rune](expected, avgConsumption, parse)
}

// String parses a token from the input, and returns the part of the input that
// matched the token.
// If the token could not be found at the current position,
// the parser returns an error result.
func String(token string) gomme.Parser[string] {
	expected := strconv.Quote(token)

	parse := func(state gomme.State) (gomme.State, string) {
		if !strings.HasPrefix(state.CurrentString(), token) {
			return state.NewError(expected, state, uint(len(token))), ""
		}

		newState := state.MoveBy(uint(len(token)))
		return newState, token
	}

	return gomme.NewParser[string](expected, gomme.ConstantConsumption(uint(len(token))), parse)
}

// Bytes parses a token from the input, and returns the part of the input that
// matched the token.
// If the token could not be found at the current position,
// the parser returns an error result.
func Bytes(token []byte) gomme.Parser[[]byte] {
	expected := fmt.Sprintf("0x%x", token)

	parse := func(state gomme.State) (gomme.State, []byte) {
		if !bytes.HasPrefix(state.CurrentBytes(), token) {
			return state.NewError(expected, state, uint(len(token))), []byte{}
		}

		newState := state.MoveBy(uint(len(token)))
		return newState, token
	}

	return gomme.NewParser[[]byte](expected, gomme.ConstantConsumption(uint(len(token))), parse)
}

// UntilString parses until it finds a token in the input, and returns
// the part of the input that preceded the token.
// If found the parser moves beyond the stop string.
// If the token could not be found, the parser returns an error result.
// This function panics if `stop` is empty.
func UntilString(stop string) gomme.Parser[string] {
	expected := fmt.Sprintf("... %q", stop)
	count := uint(0)
	sum := uint(0)

	avgConsumption := func() uint {
		if count == 0 {
			return uint(len(stop) * 2)
		}
		return (sum + count/2) / count
	}

	if stop == "" {
		panic("stop is empty")
	}

	parse := func(state gomme.State) (gomme.State, string) {
		input := state.CurrentString()
		i := strings.Index(input, stop)
		if i == -1 {
			return state.NewError(expected, state, avgConsumption()), ""
		}

		consumption := uint(i + len(stop))
		sum += consumption
		count++
		newState := state.MoveBy(consumption)
		return newState, input[:i]
	}

	return gomme.NewParser[string](expected, avgConsumption, parse)
}

// SatisfyMN returns the longest input subset that matches the predicate,
// within the boundaries of `atLeast` <= number of runes found <= `atMost`.
//
// If the provided parser is not successful or the predicate doesn't match
// `atLeast` times, the parser fails and goes back to the start.
func SatisfyMN(expected string, atMost, atLeast uint, predicate func(rune) bool) gomme.Parser[string] {
	consumptionCount := 0
	consumptionSum := 0

	avgConsumption := func() uint {
		if consumptionCount == 0 {
			return min(gomme.DefaultConsumption, atLeast)
		}
		return uint((consumptionSum + consumptionCount/2) / consumptionCount)
	}

	parse := func(state gomme.State) (gomme.State, string) {
		current := state
		count := uint(0)
		for atMost > count {
			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if count >= atLeast {
					output := state.StringTo(current)
					consumptionCount++
					consumptionSum += len(output)
					return current, output
				}
				avgc := avgConsumption()
				if size == 0 {
					return state.NewError(
						fmt.Sprintf("%s (need %d, found %d at EOF)", expected, atLeast, count),
						current, avgc*(atLeast-count)/atLeast,
					), ""
				}
				return state.NewError(
					fmt.Sprintf("%s (need %d, found %d, got UTF-8 error)", expected, atLeast, count),
					current, avgc*(atLeast-count)/atLeast,
				), ""
			}

			if !predicate(r) {
				if count >= atLeast {
					output := state.StringTo(current)
					consumptionCount++
					consumptionSum += len(output)
					return current, output
				}
				avgc := avgConsumption()
				return state.NewError(
					fmt.Sprintf("%s (need %d, found %d, got %q)", expected, atLeast, count, r),
					current, avgc*(atLeast-count)/atLeast,
				), ""
			}

			current = current.MoveBy(uint(size))
			count++
		}

		output := state.StringTo(current)
		consumptionCount++
		consumptionSum += len(output)
		return current, output
	}

	return gomme.NewParser[string](expected, avgConsumption, parse)
}

// AlphaMN parses at least `atLeast` and at most `atMost` Unicode letters.
func AlphaMN(atLeast, atMost uint) gomme.Parser[string] {
	return SatisfyMN("letter", atMost, atLeast, unicode.IsLetter)
}

// Alpha0 parses a zero or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input is empty, or no character is found, the parser
// returns the input as is.
func Alpha0() gomme.Parser[string] {
	return SatisfyMN("letter", math.MaxUint, 0, unicode.IsLetter)
}

// Alpha1 parses one or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alpha1() gomme.Parser[string] {
	return SatisfyMN("letter", math.MaxUint, 1, unicode.IsLetter)
}

// Alphanumeric0 parses zero or more alphabetical or numerical Unicode characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Alphanumeric0() gomme.Parser[string] {
	return SatisfyMN("letter or numeral", math.MaxUint, 0, IsAlphanumeric)
}

// Alphanumeric1 parses one or more alphabetical or numerical Unicode characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alphanumeric1() gomme.Parser[string] {
	return SatisfyMN("letter or numeral", math.MaxUint, 1, IsAlphanumeric)
}

// Digit0 parses zero or more ASCII numerical characters: 0-9.
// In the cases where the input is empty, or no digit character is found, the parser
// returns the input as is.
func Digit0() gomme.Parser[string] {
	return SatisfyMN("digit", math.MaxUint, 0, IsDigit)
}

// Digit1 parses one or more numerical characters: 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Digit1() gomme.Parser[string] {
	return SatisfyMN("digit", math.MaxUint, 1, IsDigit)
}

// HexDigit0 parses zero or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input is empty, or no terminating character is found, the parser
// returns the input as is.
func HexDigit0() gomme.Parser[string] {
	return SatisfyMN("hexadecimal digit", math.MaxUint, 0, IsHexDigit)
}

// HexDigit1 parses one or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func HexDigit1() gomme.Parser[string] {
	return SatisfyMN("hexadecimal digit", math.MaxUint, 1, IsHexDigit)
}

// Whitespace0 parses zero or more Unicode whitespace characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Whitespace0() gomme.Parser[string] {
	return SatisfyMN("whitespace", math.MaxUint, 0, unicode.IsSpace)
}

// Whitespace1 parses one or more Unicode whitespace characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Whitespace1() gomme.Parser[string] {
	return SatisfyMN("whitespace", math.MaxUint, 1, unicode.IsSpace)
}

// OneOf parses a single character from the given set of characters.
func OneOf(collection ...rune) gomme.Parser[rune] {
	n := len(collection)
	if n == 0 {
		panic("OneOf has no characters to match")
	}
	expected := fmt.Sprintf("one of %q", collection)

	return Satisfy(expected, func(r rune) bool {
		for _, c := range collection {
			if r == c {
				return true
			}
		}

		return false
	})
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

// Int64 parses an integer from the input, and returns it plus the remaining input.
// Only decimal integers are supported. It may start with a 0.
func Int64() gomme.Parser[int64] {
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
// Only decimal integers are supported. It may start with a 0.
func Int8() gomme.Parser[int8] {
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
// Only decimal integers are supported. It may start with a 0.
func UInt8() gomme.Parser[uint8] {
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
