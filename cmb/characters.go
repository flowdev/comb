package cmb

import (
	"bytes"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/flowdev/comb"
)

// Char parses a single rune and matches it with
// a provided candidate.
// If the rune could not be found at the current position,
// the parser returns an error result.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func Char(char rune) comb.Parser[rune] {
	expected := strconv.QuoteRune(char)

	parse := func(state comb.State) (comb.State, rune, *comb.ParserError) {
		r, size, pErr := decodeRune(state.CurrentString(), expected, state)
		if pErr != nil {
			return state, r, pErr
		}
		if r != char {
			return state, utf8.RuneError, state.NewSyntaxError("%s (got %q)", expected, r)
		}

		return state.MoveBy(size), r, nil
	}

	return comb.NewParser[rune](expected, parse, IndexOf(char))
}

// AnyChar parses a single rune.
// The parser returns an error result only at the end of the input or in case of a UTF-8 error.
// This parser is NOT a good candidate for SafeSpot and has a Forbidden recoverer.
func AnyChar() comb.Parser[rune] {
	expected := "any character"

	parse := func(state comb.State) (comb.State, rune, *comb.ParserError) {
		r, size, pErr := decodeRune(state.CurrentString(), expected, state)
		if pErr != nil {
			return state, r, pErr
		}
		return state.MoveBy(size), r, nil
	}

	return comb.NewParser[rune](expected, parse, Forbidden())
}

// QuotedChar parses and returns a single rune.
// forbiddenChars is a string containing all characters that are forbidden to be used without escaping.
// additionalEscapedChars is a string containing all characters that have to be escaped next to the standard ones.
// In almost all cases forbiddenChars and additionalEscapedChars should be the same.
//
// This parser is NOT a good candidate for SafeSpot and has a Forbidden recoverer.
// But parsers using QuotedChar in a more constrained context are a good candidate for SafeSpot.
//
// The parser returns an error result only at the end of the input or in case of a UTF-8 error.
// The standard escapes are:
// '\\' [abefnrtv\\]  // standard ANSI escapes
// '\\' [0-3][0-7][0-7] // octal escapes (possibly illegal UTF-8)
// '\\' 'x'[0-9a-fA-F][0-9a-fA-F] // hexadecimal escapes (possibly illegal UTF-8)
// '\\' 'u'[0-9a-fA-F]{4,4} // 4 hex digits for a 16 bit Unicode character
// '\\' 'U'[0-9a-fA-F]{8,8} // 8 hex digits for a 32 bit Unicode character
func QuotedChar(expected, forbiddenChars, additionalEscapedChars string) comb.Parser[rune] {
	parse := func(state comb.State) (comb.State, rune, *comb.ParserError) {
		input := state.CurrentString()
		r, _, pErr := decodeRune(input, expected, state)
		if pErr != nil {
			return state, r, pErr
		}
		if strings.ContainsRune(forbiddenChars, r) { // handle forbidden chars
			return state, utf8.RuneError, state.NewSyntaxError("%s found %q", expected, r)
		}
		if r == '\\' { // handle additional escaped chars
			var size int
			backslashLen := len("\\") // should be always 1; but better safe than sorry
			r, size, pErr = decodeRune(input[backslashLen:], expected, state)
			if pErr != nil {
				return state, r, pErr
			}
			if strings.ContainsRune(additionalEscapedChars, r) {
				return state.MoveBy(backslashLen + size), r, nil
			}
		}
		c, _, rest, err := strconv.UnquoteChar(input, 0) // all other escaped chars are handled here
		if err != nil {
			return state, utf8.RuneError, state.NewSyntaxError("%s (%v)", expected, err)
		}
		return state.MoveBy(len(input) - len(rest)), c, nil
	}

	return comb.NewParser[rune](expected, parse, Forbidden())
}

func CharClassChar(expected string, classRunes []rune, classRanges [][]rune) comb.Parser[rune] {
	if len(classRunes) == 0 && len(classRanges) == 0 {
		panic("no class runes and no class ranges provided")
	}
	for i, cr := range classRanges {
		if len(cr) < 2 {
			panic(fmt.Sprintf("class ranges must contain two runes (index %d has only %d)", i, len(cr)))
		}
	}
	parse := func(state comb.State) (comb.State, rune, *comb.ParserError) {
		input := state.CurrentString()
		r, size, pErr := decodeRune(input, expected, state)
		if pErr != nil {
			return state, r, pErr
		}
		if slices.Contains(classRunes, r) {
			return state.MoveBy(size), r, nil
		}
		for _, cr := range classRanges {
			if r >= cr[0] && r <= cr[1] {
				return state.MoveBy(size), r, nil
			}
		}
		return state, utf8.RuneError, state.NewSyntaxError("%s (got %q)", expected, r)
	}

	return comb.NewParser[rune](expected, parse, Forbidden())
}

func decodeRune(input string, expected string, state comb.State) (rune, int, *comb.ParserError) {
	r, size := utf8.DecodeRuneInString(input)
	if r == utf8.RuneError {
		if size == 0 {
			return utf8.RuneError, 0, state.NewSyntaxError(expected + " (at EOF)")
		}
		return utf8.RuneError, 1, state.NewSyntaxError(expected + " (got UTF-8 error)")
	}
	return r, size, nil
}

// Byte parses a single byte and matches it with
// a provided candidate.
// If the byte could not be found at the current position,
// the parser returns an error result.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func Byte(byt byte) comb.Parser[byte] {
	var p comb.Parser[byte]

	expected := "0x" + strconv.FormatUint(uint64(byt), 16)

	parse := func(state comb.State) (comb.State, byte, *comb.ParserError) {
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

	p = comb.NewParser[byte](expected, parse, IndexOf(byt))
	return p
}

// AnyByte parses a single byte and matches it with
// a provided candidate.
// If the byte could not be found at the current position,
// the parser returns an error result.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func AnyByte() comb.Parser[byte] {
	expected := "any byte"

	parse := func(state comb.State) (comb.State, byte, *comb.ParserError) {
		buf := state.CurrentBytes()
		if len(buf) == 0 {
			return state, 0, state.NewSyntaxError(expected + " (at EOF)")
		}
		b := buf[0]

		return state.MoveBy(1), b, nil
	}

	return comb.NewParser[byte](expected, parse, Forbidden())
}

// Satisfy parses a single character and ensures that it satisfies the given predicate.
// `expected` is used in error messages to tell the user what is expected at the current position.
//
// This parser is a good candidate for SafeSpot and has an optimized Recoverer.
// An even more specialized Recoverer can be used later with `parser.SwapRecoverer(newRecoverer)`.
func Satisfy(expected string, predicate func(rune) bool) comb.Parser[rune] {
	var p comb.Parser[rune]

	parse := func(state comb.State) (comb.State, rune, *comb.ParserError) {
		r, size, pErr := decodeRune(state.CurrentString(), expected, state)
		if pErr != nil {
			return state, r, pErr
		}
		if !predicate(r) {
			return state, utf8.RuneError, state.NewSyntaxError("%s (got %q)", expected, r)
		}

		return state.MoveBy(size), r, nil
	}

	recoverer := func(state comb.State, _ interface{}) (int, interface{}) {
		return strings.IndexFunc(state.CurrentString(), predicate), nil
	}

	p = comb.NewParser[rune](expected, parse, recoverer)
	return p
}

// String parses a token from the input and returns the part of the input that
// matched the token.
// If the token could not be found at the current position,
// the parser returns an error result.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func String(token string) comb.Parser[string] {
	var p comb.Parser[string]

	expected := strconv.Quote(token)

	parse := func(state comb.State) (comb.State, string, *comb.ParserError) {
		if !strings.HasPrefix(state.CurrentString(), token) {
			return state, "", state.NewSyntaxError(expected)
		}

		newState := state.MoveBy(len(token))
		return newState, token, nil
	}

	p = comb.NewParser[string](expected, parse, IndexOf(token))
	return p
}

// Bytes parses a token from the input and returns the part of the input that
// matched the token.
// If the token could not be found at the current position,
// the parser returns an error result.
func Bytes(token []byte) comb.Parser[[]byte] {
	var p comb.Parser[[]byte]

	expected := fmt.Sprintf("0x%x", token)

	parse := func(state comb.State) (comb.State, []byte, *comb.ParserError) {
		if !bytes.HasPrefix(state.CurrentBytes(), token) {
			return state, []byte{}, state.NewSyntaxError(expected)
		}

		newState := state.MoveBy(len(token))
		return newState, token, nil
	}

	p = comb.NewParser[[]byte](expected, parse, IndexOf(token))
	return p
}

// SatisfyMN returns the longest input subset that matches the predicate,
// within the boundaries of `atLeast` <= number of runes found <= `atMost`.
//
// If the provided parser is not successful or the predicate doesn't match
// `atLeast` times, the parser fails and goes back to the start.
//
// This parser is a good candidate for SafeSpot and has an optimized Recoverer.
// An even more specialized Recoverer can be used later with `parser.SwapRecoverer(newRecoverer) Parser`.
func SatisfyMN(expected string, atLeast, atMost int, predicate func(rune) bool) comb.Parser[string] {
	var p comb.Parser[string]

	if atLeast < 0 {
		panic("SatisfyMN is unable to handle negative `atLeast` argument")
	}
	if atMost < 0 {
		panic("SatisfyMN is unable to handle negative `atMost` argument")
	}

	parse := func(state comb.State) (comb.State, string, *comb.ParserError) {
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
					return state, "", state.NewSyntaxError("%s (need %d, found %d at EOF)", expected, atLeast, count)
				}
				return state, "", state.NewSyntaxError("%s (need %d, found %d, got UTF-8 error)", expected, atLeast, count)
			}

			if !predicate(r) {
				if count >= atLeast {
					output := state.StringTo(current)
					return current, output, nil
				}
				return state, "", state.NewSyntaxError("%s (need %d, found %d, got %q)", expected, atLeast, count, r)
			}

			current = current.MoveBy(size)
			count++
		}

		output := state.StringTo(current)
		return current, output, nil
	}

	p = comb.NewParser[string](expected, parse, satisfyMNRecoverer(atLeast, predicate))
	return p
}

func satisfyMNRecoverer(atLeast int, predicate func(rune) bool) comb.Recoverer {
	return func(state comb.State, _ interface{}) (int, interface{}) {
		count := 0
		start := 0
		for i, r := range state.CurrentString() {
			if predicate(r) {
				if count == 0 {
					start = i
				}
				count++
				if count >= atLeast {
					return start, nil
				}
			} else {
				count = 0
			}
		}
		return comb.RecoverWasteTooMuch, nil
	}
}

// AlphaMN parses at least `atLeast` and at most `atMost` Unicode letters.
func AlphaMN(atLeast, atMost int) comb.Parser[string] {
	return SatisfyMN("letter", atLeast, atMost, unicode.IsLetter)
}

// Alpha0 parses zero or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input is empty, or no character is found, the parser
// returns the input as is.
func Alpha0() comb.Parser[string] {
	return SatisfyMN("letter", 0, math.MaxInt, unicode.IsLetter)
}

// Alpha1 parses one or more lowercase or uppercase alphabetic Unicode characters (e.g., a-z or A-Z).
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alpha1() comb.Parser[string] {
	return SatisfyMN("letter", 1, math.MaxInt, unicode.IsLetter)
}

// Alphanumeric0 parses zero or more alphabetical or numerical Unicode characters.
// An '_' is considered a valid character, too.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Alphanumeric0() comb.Parser[string] {
	return SatisfyMN("letter or numeral", 0, math.MaxInt, IsAlphanumeric)
}

// Alphanumeric1 parses one or more alphabetical or numerical Unicode characters.
// An '_' is considered a valid character, too.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alphanumeric1() comb.Parser[string] {
	return SatisfyMN("letter or numeral", 1, math.MaxInt, IsAlphanumeric)
}

// Digit0 parses zero or more ASCII numerical characters: 0-9.
// In the cases where the input is empty, or no digit character is found, the parser
// returns the input as is.
func Digit0() comb.Parser[string] {
	return SatisfyMN("digit", 0, math.MaxInt, IsDigit)
}

// Digit1 parses one or more numerical characters: 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Digit1() comb.Parser[string] {
	return SatisfyMN("digit", 1, math.MaxInt, IsDigit)
}

// HexDigit0 parses zero or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input is empty, or no terminating character is found, the parser
// returns the input as is.
func HexDigit0() comb.Parser[string] {
	return SatisfyMN("hexadecimal digit", 0, math.MaxInt, IsHexDigit)
}

// HexDigit1 parses one or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func HexDigit1() comb.Parser[string] {
	return SatisfyMN("hexadecimal digit", 1, math.MaxInt, IsHexDigit)
}

// Whitespace0 parses zero or more Unicode whitespace characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Whitespace0() comb.Parser[string] {
	return SatisfyMN("whitespace", 0, math.MaxInt, unicode.IsSpace)
}

// Whitespace1 parses one or more Unicode whitespace characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Whitespace1() comb.Parser[string] {
	return SatisfyMN("whitespace", 1, math.MaxInt, unicode.IsSpace)
}

// OneOfRunes parses a single character from the given set of characters.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func OneOfRunes(collection ...rune) comb.Parser[rune] {
	n := len(collection)
	if n == 0 {
		panic("OneOfRunes has no characters to match")
	}
	expected := fmt.Sprintf("one of %q", collection)

	parser := Satisfy(expected, func(r rune) bool {
		return slices.Contains(collection, r)
	})
	parser.SwapRecoverer(func(state comb.State, _ interface{}) (int, interface{}) {
		return strings.IndexAny(state.CurrentString(), string(collection)), nil
	})
	return parser
}

// OneOf parses a single string from the given set of strings.
// This parser is a good candidate for SafeSpot and has an optimized recoverer.
func OneOf(collection ...string) comb.Parser[string] {
	var p comb.Parser[string]

	n := len(collection)
	if n == 0 {
		panic("OneOf has no strings to match")
	}
	expected := fmt.Sprintf("one of %q", collection)

	parse := func(state comb.State) (comb.State, string, *comb.ParserError) {
		input := state.CurrentString()
		for _, token := range collection {
			if strings.HasPrefix(input, token) {
				return state.MoveBy(len(token)), token, nil
			}
		}

		return state, "", state.NewSyntaxError(expected)
	}

	p = comb.NewParser[string](expected, parse, IndexOfAny(collection...))
	return p
}

// LF parses a line feed `\n` character.
func LF() comb.Parser[rune] {
	return Char('\n')
}

// CR parses a carriage return `\r` character.
func CR() comb.Parser[rune] {
	return Char('\r')
}

// CRLF parses the string `\r\n`.
func CRLF() comb.Parser[string] {
	return String("\r\n")
}

// Space parses an ASCII space character (' ').
func Space() comb.Parser[rune] {
	return Char(' ')
}

// Tab parses an ASCII tab character ('\t').
func Tab() comb.Parser[rune] {
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
