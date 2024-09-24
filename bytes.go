package gomme

import (
	"fmt"
	"strings"
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
			for i, _ := range input.CurrentString() { // this will loop over runes of a string
				current := input
				current.Pos += uint(i)
				res := parse(current)
				if res.Err == nil {
					return Success(input.BytesTo(current), res.Remaining)
				}
			}
		} else {
			for i, _ := range input.CurrentBytes() { // this will loop over bytes
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

		lastValidPos := 0
		count := uint(0)
		for i, r := range input.CurrentString() { // this will loop over runes of a string
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

		newInput := input.MoveBy(uint(lastValidPos))
		return Success(input.BytesTo(newInput), newInput)
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
