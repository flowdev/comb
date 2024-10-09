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
	consumptionCount := uint(0)
	consumptionSum := uint(0)

	avgConsumption := func() uint {
		if consumptionCount == 0 {
			return gomme.DefaultConsumption
		}
		return (consumptionSum + consumptionCount/2) / consumptionCount
	}

	parseMany := func(state gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, min(32, atMost))
		remaining := state
		count := uint(0)
		for {
			if count >= atMost {
				consumptionCount++
				consumptionSum += uint(state.ByteCount(remaining))
				return remaining, outputs
			}
			newState, output := parse.It(remaining)
			if newState.Failed() && newState.NoWayBack() {
				newState, output = gomme.HandleAllErrors(remaining.Failure(newState), parse) // this will force it through
			} else if newState.Failed() {
				if count < atLeast {
					// TODO: In error handling mode Insert we should "insert" missing results to reach atLeast
					// TODO: Think this case better through
					newState, output = gomme.HandleCurrentError(remaining.Failure(newState), parse)
					if newState.Failed() {
						return state.Failure(newState), []Output{}
					}
				} else {
					consumptionCount++
					consumptionSum += uint(state.ByteCount(remaining))
					return remaining, outputs
				}
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if !newState.Moved(remaining) {
				avgc := avgConsumption()
				return state.NewError(fmt.Sprintf("%s (found empty element)", parse.Expected()),
					remaining, avgc*(atLeast-count)/atLeast,
				), []Output{}
			}

			outputs = append(outputs, output)
			remaining = newState
			count++
		}
	}

	return gomme.NewParser[[]Output]("ManyMN", avgConsumption, parseMany)
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

	consumptionCount := uint(0)
	consumptionSum := uint(0)

	avgConsumption := func() uint {
		if consumptionCount == 0 {
			return gomme.DefaultConsumption
		}
		return (consumptionSum + consumptionCount/2) / consumptionCount
	}

	parseSep := func(state gomme.State) (gomme.State, []Output) {
		if atMost == 0 {
			consumptionCount++
			return state, []Output{}
		}

		firstState, firstOutput := parse.It(state)
		firstMoved := firstState.Moved(state)
		if firstState.Failed() && firstState.NoWayBack() {
			firstState, firstOutput = gomme.HandleAllErrors(state.Failure(firstState), parse) // this will force it through
		} else if firstState.Failed() {
			if atLeast > 0 {
				// TODO: Is this correct? Not in handle error mode "Insert"!
				firstState, firstOutput = gomme.HandleCurrentError(state.Failure(firstState), parse)
				if firstState.Failed() {
					return state.Failure(firstState), []Output{}
				}
				return state.Failure(firstState), []Output{}
			}
			consumptionCount++
			return state, []Output{} // still success
		}

		newState, outputs := parseMany.It(firstState)
		if newState.Failed() {
			return state.Failure(newState), []Output{} // parseMany handled errors already
		}

		if parseSeparatorAtEnd {
			separatorState, _ := separator.It(newState)
			if !separatorState.Failed() {
				newState = separatorState
			}
		}

		consumptionCount++
		consumptionSum += uint(len(state.BytesTo(newState)))
		finalOutputs := make([]Output, 0, len(outputs)+1)
		if firstMoved {
			finalOutputs = append(finalOutputs, firstOutput)
		}
		return newState, append(finalOutputs, outputs...)
	}

	return gomme.NewParser[[]Output]("SeperatedMN", avgConsumption, parseSep)
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
