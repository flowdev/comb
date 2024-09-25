package gomme

import (
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Char parses a single character and matches it with
// a provided candidate.
func Char(character rune) Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() {
			return Failure[rune](NewError(input, string(character)), input)
		}
		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[rune](NewError(input, string(character)), input)
		}
		if r != character {
			return Failure[rune](NewError(input, string(character)), input)
		}

		return Success(r, input.MoveBy(uint(size)))
	}
}

// AnyChar parses any single character.
func AnyChar() Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() {
			return Failure[rune](NewError(input, "AnyChar"), input)
		}
		r, size := utf8.DecodeRune(input.Bytes[input.Pos:])
		if r == utf8.RuneError {
			return Failure[rune](NewError(input, "AnyChar"), input)
		}

		return Success(r, input.MoveBy(uint(size)))
	}
}

// Alpha0 parses a zero or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input is empty, or no character is found, the parser
// returns the input as is.
func Alpha0() Parser[string] {
	return func(input Input) Result[string] {
		current := input
		for !current.AtEnd() { // loop over runes of the input
			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Alpha0: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !unicode.IsLetter(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// Alpha1 parses one or more lowercase or uppercase alphabetic characters: a-z, A-Z.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alpha1() Parser[string] {
	return func(input Input) Result[string] {
		if input.AtEnd() {
			return Failure[string](NewError(input, "Alpha1"), input)
		}

		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[string](NewError(input, "Alpha1: UTF-8 error"), input)
		}
		if !unicode.IsLetter(r) {
			return Failure[string](NewError(input, "Alpha1"), input)
		}
		current := input.MoveBy(uint(size))

		for !current.AtEnd() { // loop over runes of the input
			r, size = utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Alpha1: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !unicode.IsLetter(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// Alphanumeric0 parses zero or more alphabetical or numerical Unicode characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Alphanumeric0() Parser[string] {
	return func(input Input) Result[string] {
		current := input
		for !current.AtEnd() { // loop over runes of the input
			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Alphanumeric0: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !IsAlphanumeric(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// Alphanumeric1 parses one or more alphabetical or numerical Unicode characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Alphanumeric1() Parser[string] {
	return func(input Input) Result[string] {
		if input.AtEnd() {
			return Failure[string](NewError(input, "Alphanumeric1"), input)
		}

		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[string](NewError(input, "Alphanumeric1: UTF-8 error"), input)
		}
		if !IsAlphanumeric(r) {
			return Failure[string](NewError(input, "Alphanumeric1"), input)
		}
		current := input.MoveBy(uint(size))

		for !current.AtEnd() { // loop over runes of the input
			r, size = utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Alphanumeric1: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !IsAlphanumeric(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// Digit0 parses zero or more ASCII numerical characters: 0-9.
// In the cases where the input is empty, or no digit character is found, the parser
// returns the input as is.
func Digit0() Parser[string] {
	return func(input Input) Result[string] {
		current := input
		for !current.AtEnd() { // loop over runes of the input
			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Digit0: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !IsDigit(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// Digit1 parses one or more numerical characters: 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Digit1() Parser[string] {
	return func(input Input) Result[string] {
		if input.AtEnd() {
			return Failure[string](NewError(input, "Digit1"), input)
		}

		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[string](NewError(input, "Digit1: UTF-8 error"), input)
		}
		if !IsDigit(r) {
			return Failure[string](NewError(input, "Digit1"), input)
		}
		current := input.MoveBy(uint(size))

		for !current.AtEnd() { // loop over runes of the input
			r, size = utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Digit1: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !IsDigit(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// HexDigit0 parses zero or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input is empty, or no terminating character is found, the parser
// returns the input as is.
func HexDigit0() Parser[string] {
	return func(input Input) Result[string] {
		current := input
		for !current.AtEnd() { // loop over runes of the input
			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Digit0: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !IsHexDigit(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// HexDigit1 parses one or more ASCII hexadecimal characters: a-f, A-F, 0-9.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func HexDigit1() Parser[string] {
	return func(input Input) Result[string] {
		if input.AtEnd() {
			return Failure[string](NewError(input, "HexDigit1"), input)
		}

		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[string](NewError(input, "HexDigit1: UTF-8 error"), input)
		}
		if !IsHexDigit(r) {
			return Failure[string](NewError(input, "HexDigit1"), input)
		}
		current := input.MoveBy(uint(size))

		for !current.AtEnd() { // loop over runes of the input
			r, size = utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "HexDigit1: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !IsHexDigit(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// Whitespace0 parses zero or more Unicode whitespace characters.
// In the cases where the input is empty, or no matching character is found, the parser
// returns the input as is.
func Whitespace0() Parser[string] {
	return func(input Input) Result[string] {
		current := input
		for !current.AtEnd() { // loop over runes of the input
			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Whitespace0: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !unicode.IsSpace(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// Whitespace1 parses one or more Unicode whitespace characters.
// In the cases where the input doesn't hold enough data, or a terminating character
// is found before any matching ones were, the parser returns an error result.
func Whitespace1() Parser[string] {
	return func(input Input) Result[string] {
		if input.AtEnd() {
			return Failure[string](NewError(input, "Whitespace1"), input)
		}

		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[string](NewError(input, "Whitespace1: UTF-8 error"), input)
		}
		if !unicode.IsSpace(r) {
			return Failure[string](NewError(input, "Whitespace1"), input)
		}
		current := input.MoveBy(uint(size))

		for !current.AtEnd() { // loop over runes of the input
			r, size = utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[string](NewError(current, "Whitespace1: UTF-8 error"), input)
				}
				return Success(input.StringTo(current), current)
			}
			if !unicode.IsSpace(r) {
				return Success(input.StringTo(current), current)
			}
			current = current.MoveBy(uint(size))
		}

		return Success(input.StringTo(current), current)
	}
}

// LF parses a line feed `\n` character.
func LF() Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() || input.CurrentBytes()[0] != '\n' {
			return Failure[rune](NewError(input, "LF"), input)
		}

		return Success('\n', input.MoveBy(1))
	}
}

// CR parses a carriage return `\r` character.
func CR() Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() || input.CurrentBytes()[0] != '\r' {
			return Failure[rune](NewError(input, "CR"), input)
		}

		return Success('\r', input.MoveBy(1))
	}
}

// CRLF parses the string `\r\n`.
func CRLF() Parser[string] {
	return func(input Input) Result[string] {
		bytes := input.CurrentBytes()
		if len(bytes) < 2 || (bytes[0] != '\r' || bytes[1] != '\n') {
			return Failure[string](NewError(input, "CRLF"), input)
		}

		return Success(string(bytes[:2]), input.MoveBy(2))
	}
}

// OneOf parses a single character from the given set of characters.
func OneOf(collection ...rune) Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() {
			return Failure[rune](NewError(input, "OneOf"), input)
		}

		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[rune](NewError(input, "OneOf: UTF-8 error"), input)
		}
		for _, c := range collection {
			if r == c {
				return Success(r, input.MoveBy(uint(size)))
			}
		}

		return Failure[rune](NewError(input, "OneOf"), input)
	}
}

// Satisfy parses a single character, and ensures that it satisfies the given predicate.
func Satisfy(predicate func(rune) bool) Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() {
			return Failure[rune](NewError(input, "Satisfy"), input)
		}

		r, size := utf8.DecodeRune(input.CurrentBytes())
		if r == utf8.RuneError {
			return Failure[rune](NewError(input, "Satisfy: UTF-8 error"), input)
		}

		if !predicate(r) {
			return Failure[rune](NewError(input, "Satisfy"), input)
		}

		return Success(r, input.MoveBy(uint(size)))
	}
}

// Space parses an ASCII space character (' ').
func Space() Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() || input.CurrentBytes()[0] != ' ' {
			return Failure[rune](NewError(input, "Space"), input)
		}

		return Success(' ', input.MoveBy(1))
	}
}

// Tab parses an ASCII tab character ('\t').
func Tab() Parser[rune] {
	return func(input Input) Result[rune] {
		if input.AtEnd() || input.CurrentBytes()[0] != '\t' {
			return Failure[rune](NewError(input, "Tab"), input)
		}

		return Success('\t', input.MoveBy(1))
	}
}

// Int64 parses an integer from the input, and returns it plus the remaining input.
func Int64() Parser[int64] {
	return func(input Input) Result[int64] {
		parser := Sequence(Recognize(Optional(OneOf('-', '+'))), Recognize(Digit1()))

		result := parser(input)
		if result.Err != nil {
			return Failure[int64](NewError(input, "Int64"), input)
		}

		n, err := strconv.ParseInt(input.StringTo(result.Remaining), 10, 64)
		if err != nil {
			return Failure[int64](NewError(input, "Int64: strconv error"), input)
		}

		return Success(n, result.Remaining)
	}
}

// Int8 parses an 8-bit integer from the input,
// and returns the part of the input that matched the integer.
func Int8() Parser[int8] {
	return func(input Input) Result[int8] {
		parser := Sequence(Recognize(Optional(OneOf('-', '+'))), Recognize(Digit1()))

		result := parser(input)
		if result.Err != nil {
			return Failure[int8](NewError(input, "Int8"), input)
		}

		n, err := strconv.ParseInt(input.StringTo(result.Remaining), 10, 8)
		if err != nil {
			return Failure[int8](NewError(input, "Int8: strconv error"), input)
		}

		return Success(int8(n), result.Remaining)
	}
}

// UInt8 parses an 8-bit integer from the input,
// and returns the part of the input that matched the integer.
func UInt8() Parser[uint8] {
	return func(input Input) Result[uint8] {
		parser := Sequence(Recognize(Optional(Char('+'))), Recognize(Digit1()))

		result := parser(input)
		if result.Err != nil {
			return Failure[uint8](NewError(input, "Int8"), input)
		}

		n, err := strconv.ParseUint(string(result.Output[1]), 10, 8)
		if err != nil {
			return Failure[uint8](NewError(input, "Int8: strconv error"), input)
		}

		return Success(uint8(n), result.Remaining)
	}
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
