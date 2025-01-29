package pcb

import (
	"github.com/oleiade/gomme"
	"math"
)

// Count runs the provided parser `count` times.
//
// If the provided parser cannot be successfully applied `count` times, the operation
// fails and the Result will contain an error.
func Count[Output any](parse gomme.Parser[Output], count int) gomme.Parser[[]Output] {
	if count < 0 {
		panic("Count is unable to handle negative `count`")
	}

	return ManyMN(parse, count, count)
}

// Many0 applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output.
//
// Note that Many0 will succeed even if the parser fails to match at all. It will
// however fail if the provided parser accepts empty inputs (such as `Digit0`, or
// `Alpha0`) in order to prevent infinite loops.
func Many0[Output any](parse gomme.Parser[Output]) gomme.Parser[[]Output] {
	return ManyMN(parse, 0, math.MaxInt)
}

// Many1 applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output. Many1 will fail if the parser fails to
// match at least once.
//
// Note that Many1 will fail if the provided parser accepts empty
// inputs (such as `Digit0`, or `Alpha0`) in order to prevent infinite loops.
func Many1[Output any](parse gomme.Parser[Output]) gomme.Parser[[]Output] {
	return ManyMN(parse, 1, math.MaxInt)
}

// ManyMN applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output.
//
// Note that ManyMN fails if the provided parser accepts empty inputs (such as
// `Digit0`, or `Alpha0`) in order to prevent infinite loops.
func ManyMN[Output any](parse gomme.Parser[Output], atLeast, atMost int) gomme.Parser[[]Output] {
	return SeparatedMN[Output, string](parse, nil, atLeast, atMost, false)
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
	return SeparatedMN(parse, separator, 0, math.MaxInt, parseSeparatorAtEnd)
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
	return SeparatedMN(parse, separator, 1, math.MaxInt, parseSeparatorAtEnd)
}
