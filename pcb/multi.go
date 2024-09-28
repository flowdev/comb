package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
	"math"
)

// Count runs the provided parser `count` times.
//
// If the provided parser cannot be successfully applied `count` times, the operation
// fails and the Result will contain an error.
func Count[Output any](parse gomme.Parser[Output], count uint) gomme.Parser[[]Output] {
	return ManyMN(parse, count, count)
}

// ManyMN applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output.
//
// Note that ManyMN fails if the provided parser accepts empty inputs (such as
// `Digit0`, or `Alpha0`) in order to prevent infinite loops.
func ManyMN[Output any](parse gomme.Parser[Output], atLeast, atMost uint) gomme.Parser[[]Output] {
	expected := "ManyMN"

	return func(input gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, min(32, atMost))
		remaining := input
		count := uint(0)
		for {
			if count >= atMost {
				return remaining, outputs
			}
			newState, output := parse(remaining)
			if newState.Failed() {
				if count < atLeast || newState.NoWayBack() {
					return input.Failure(newState), []Output{}
				}
				return remaining, outputs
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if !newState.Moved(remaining) {
				return input.AddError(fmt.Sprintf("%s (got empty element)", expected)), []Output{}
			}

			outputs = append(outputs, output)
			remaining = newState
			count++
		}
	}
}

// Many0 applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output.
//
// Note that Many0 will succeed even if the parser fails to match at all. It will
// however fail if the provided parser accepts empty inputs (such as `Digit0`, or
// `Alpha0`) in order to prevent infinite loops.
func Many0[Output any](parse gomme.Parser[Output]) gomme.Parser[[]Output] {
	return ManyMN(parse, 0, math.MaxUint)
}

// Many1 applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output. Many1 will fail if the parser fails to
// match at least once.
//
// Note that Many1 will fail if the provided parser accepts empty
// inputs (such as `Digit0`, or `Alpha0`) in order to prevent infinite loops.
func Many1[Output any](parse gomme.Parser[Output]) gomme.Parser[[]Output] {
	return ManyMN(parse, 1, math.MaxUint)
}

// SeparatedMN applies an element parser and a separator parser repeatedly in order
// to produce a slice of elements.
//
// Because the `SeparatedListMN` is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed even if the separator parser fails to
// match at the end.
//
// The parser will fail if both parsers together accepted an empty input
// in order to prevent infinite loops.
func SeparatedMN[Output any, S gomme.Separator](
	parse gomme.Parser[Output], separator gomme.Parser[S],
	atLeast, atMost uint,
	parseSeparatorAtEnd bool,
) gomme.Parser[[]Output] {
	parseMany := ManyMN(Preceded(separator, parse), max(atLeast, 1)-1, atMost-1)

	return func(state gomme.State) (gomme.State, []Output) {
		if atMost == 0 {
			return state, []Output{}
		}

		firstState, firstOutput := parse(state)
		firstMoved := firstState.Moved(state)
		if firstState.Failed() {
			if atLeast > 0 || firstState.NoWayBack() {
				return state.Failure(firstState), []Output{}
			}
			return state, []Output{} // still success
		}

		newState, outputs := parseMany(firstState)
		if newState.Failed() {
			return state.Failure(newState), []Output{}
		}

		if parseSeparatorAtEnd {
			separatorState, _ := separator(newState)
			if !separatorState.Failed() {
				newState = separatorState
			}
		}

		finalOutputs := make([]Output, 0, len(outputs)+1)
		if firstMoved {
			finalOutputs = append(finalOutputs, firstOutput)
		}
		return newState, append(finalOutputs, outputs...)
	}
}

// Separated0 applies an element parser and a separator parser repeatedly in order
// to produce a list of elements.
//
// Note that Separated0 will succeed even if the element parser fails to match at all.
//
// Because the `Separated0` is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed even if the separator parser fails to
// match at all.
//
// The parser will fail if the both parsers together accepted an empty input
// in order to prevent infinite loops.
func Separated0[Output any, S gomme.Separator](
	parse gomme.Parser[Output], separator gomme.Parser[S],
	parseSeparatorAtEnd bool,
) gomme.Parser[[]Output] {
	return SeparatedMN(parse, separator, 0, math.MaxUint, parseSeparatorAtEnd)
}

// Separated1 applies an element parser and a separator parser repeatedly in order
// to produce a list of elements.
//
// Note that Separated1 will fail if the element parser fails to match at all.
//
// Because the `SeparatedList1` is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed even if the separator parser fails to
// match at all.
func Separated1[Output any, S gomme.Separator](
	parse gomme.Parser[Output], separator gomme.Parser[S],
	parseSeparatorAtEnd bool,
) gomme.Parser[[]Output] {
	return SeparatedMN(parse, separator, 1, math.MaxUint, parseSeparatorAtEnd)
}
