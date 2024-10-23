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
	atLeast, atMost int,
	parseSeparatorAtEnd bool,
) gomme.Parser[[]Output] {
	if atLeast < 0 {
		panic("SeparatedMN is unable to handle negative `atLeast`")
	}
	if atMost < 0 {
		panic("SeparatedMN is unable to handle negative `atMost`")
	}

	var sep1N gomme.Parser[[]Output]

	parseManySP := ManyMN(Prefixed(separator, parse), max(atLeast-1, 0), max(atMost-1, 0))
	if parseSeparatorAtEnd {
		sep1N = Map3(parse, parseManySP, Optional(separator), func(out1 Output, outMany []Output, _ S) ([]Output, error) {
			outAll := make([]Output, 0, len(outMany)+1)
			outAll = append(outAll, out1)
			return append(outAll, outMany...), nil
		})
	} else {
		sep1N = Map2(parse, parseManySP, func(out1 Output, outMany []Output) ([]Output, error) {
			outAll := make([]Output, 0, len(outMany)+1)
			outAll = append(outAll, out1)
			return append(outAll, outMany...), nil
		})
	}

	if atLeast > 0 {
		return sep1N
	}
	return Optional(sep1N)
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
