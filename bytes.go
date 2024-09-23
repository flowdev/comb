package gomme

import (
	"fmt"
	"strings"
)

// Take returns a subset of the input of size `count`.
func Take(count uint) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		if uint(len(input.Bytes))-input.Pos < count {
			return Failure[[]byte](NewError(input, fmt.Sprintf("Take(%d)", count)), input)
		}

		oldPos := input.Pos
		input.Pos += count
		return Success(input.Bytes[oldPos:input.Pos], input)
	}
}

// TakeUntil parses any number of characters until the provided parser is successful.
// If the provided parser is not successful, the parser fails, and the entire input is
// returned as the Result's Remaining.
func TakeUntil[Output any](parse Parser[Output]) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		if input.Pos >= uint(len(input.Bytes)) {
			return Failure[[]byte](NewError(input, "TakeUntil"), input)
		}

		if input.Text {
			for i, _ := range string(input.Bytes[input.Pos:]) { // this will loop over runes of a string
				current := input
				current.Pos += uint(i)
				res := parse(current)
				if res.Err == nil {
					oldPos := input.Pos
					input.Pos += uint(i)
					return Success(input.Bytes[oldPos:input.Pos], input)
				}
			}

		} else {
			for i := input.Pos; i < uint(len(input.Bytes)); i++ { // this will loop over bytes
				current := input
				current.Pos += i
				res := parse(current)
				if res.Err == nil {
					oldPos := input.Pos
					input.Pos += i
					return Success(input.Bytes[oldPos:input.Pos], input)
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
		if input.Pos >= uint(len(input.Bytes)) {
			return Failure[[]byte](NewError(input, "TakeWhileMN"), input)
		}

		// Input is shorter than the minimum expected matching length,
		// it is thus not possible to match it within the established
		// constraints.
		if uint(len(input.Bytes))-input.Pos < atLeast {
			return Failure[[]byte](NewError(input, fmt.Sprintf("TakeWhileMN(%d, ...)", atLeast)), input)
		}

		lastValidPos := 0
		count := uint(0)
		for i, r := range string(input.Bytes[input.Pos:]) { // this will loop over runes of a string
			if count >= atMost {
				break
			}
			matched := predicate(r)
			if !matched {
				if count < atLeast {
					return Failure[[]byte](NewError(input, "TakeWhileMN"), input)
				}

				oldPos := input.Pos
				input.Pos += uint(i)
				return Success(input.Bytes[oldPos:input.Pos], input)
			}

			count++
			lastValidPos = i
		}

		oldPos := input.Pos
		input.Pos += uint(lastValidPos)
		return Success(input.Bytes[oldPos:input.Pos], input)
	}
}

// Token parses a token from the input, and returns the part of the input that
// matched the token.
// If the token could not be found, the parser returns an error result.
func Token(token string) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		if !strings.HasPrefix(string(input.Bytes[input.Pos:]), token) {
			return Failure[[]byte](NewError(input, fmt.Sprintf("Token(%s)", token)), input)
		}

		oldPos := input.Pos
		input.Pos += uint(len(token))
		return Success(input.Bytes[oldPos:input.Pos], input)
	}
}
