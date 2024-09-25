package gomme

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Take returns the next `count` bytes of the input.
func Take(count uint) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		if input.BytesRemaining() < count {
			return Failure[[]byte](NewError(input, fmt.Sprintf("Take(%d)", count)), input)
		}

		newInput := input.MoveBy(count)
		return Success(input.BytesTo(newInput), newInput)
	}
}

// TakeUntil parses any number of characters until the provided parser is successful.
// If the provided parser is not successful, the parser fails, and the entire input is
// returned as the Result's Remaining.
func TakeUntil[Output any](parse Parser[Output]) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		if input.AtEnd() {
			return Failure[[]byte](NewError(input, "TakeUntil"), input)
		}

		if input.Text {
			for i := range input.CurrentString() { // this will loop over runes of a string
				current := input
				current.Pos += uint(i)
				res := parse(current)
				if res.Err == nil {
					return Success(input.BytesTo(current), res.Remaining)
				}
			}
		} else {
			for i := range input.CurrentBytes() { // this will loop over bytes
				current := input
				current.Pos += uint(i)
				res := parse(current)
				if res.Err == nil {
					return Success(input.BytesTo(current), res.Remaining)
				}
			}
		}

		return Failure[[]byte](NewError(input, "TakeUntil"), input)
	}
}

// TakeWhileMN returns the longest input subset that matches the predicates, within
// the boundaries of `atLeast` <= len(input) <= `atMost`.
//
// If the provided parser is not successful or the pattern is out of the
// `atLeast` <= len(input) <= `atMost` range, the parser fails, and the entire
// input is returned as the Result's Remaining.
func TakeWhileMN(atLeast, atMost uint, predicate func(rune) bool) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		if input.AtEnd() {
			return Failure[[]byte](NewError(input, "TakeWhileMN"), input)
		}

		// Input is shorter than the minimum expected matching length,
		// it is thus not possible to match it within the established
		// constraints.
		if input.BytesRemaining() < atLeast {
			return Failure[[]byte](NewError(input, fmt.Sprintf("TakeWhileMN(%d, ...)", atLeast)), input)
		}

		count := uint(0)
		current := input
		for !current.AtEnd() { // loop over runes of the input
			if count >= atMost {
				break
			}

			r, size := utf8.DecodeRune(current.CurrentBytes())
			if r == utf8.RuneError {
				if input.Text {
					return Failure[[]byte](NewError(current, "TakeWhileMN: UTF-8 error"), input)
				}
				return Success(input.BytesTo(current), current)
			}

			matched := predicate(r)
			if !matched {
				if count < atLeast {
					return Failure[[]byte](NewError(input, "TakeWhileMN"), input)
				}
				return Success(input.BytesTo(current), current)
			}

			count++
			current = current.MoveBy(uint(size))
		}

		return Success(input.BytesTo(current), current)
	}
}

// Token parses a token from the input, and returns the part of the input that
// matched the token.
// If the token could not be found, the parser returns an error result.
func Token(token string) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		if !strings.HasPrefix(input.CurrentString(), token) {
			return Failure[[]byte](NewError(input, fmt.Sprintf("Token(%s)", token)), input)
		}

		newInput := input.MoveBy(uint(len(token)))
		return Success(input.BytesTo(newInput), newInput)
	}
}

// BytesToString convertes a parser that has an output of []byte to a parser that outputs a string.
// Errors are not changed. This allowes parsers from this file to be used in text parsers.
func BytesToString(parse Parser[[]byte]) Parser[string] {
	return func(input InputBytes) Result[string] {
		res := parse(input)
		return Result[string]{
			Output:    string(res.Output),
			Err:       res.Err,
			Remaining: res.Remaining,
		}
	}
}
