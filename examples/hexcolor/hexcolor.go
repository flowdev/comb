// Package hexcolor implements a parser for hexadecimal color strings.
// It demonstrates how to use gomme to build a parser for a simple string format.
package hexcolor

import (
	"strconv"

	"github.com/oleiade/gomme"
)

// RGBColor stores the three bytes describing a color in the RGB space.
type RGBColor struct {
	red   uint8
	green uint8
	blue  uint8
}

// ParseRGBColor creates a new RGBColor from a hexadecimal color string.
// The string must be a six digit hexadecimal number, prefixed with a "#".
func ParseRGBColor(input string) (RGBColor, error) {
	parser := gomme.Preceded(
		gomme.String("#"),
		gomme.Map1(
			gomme.Count(HexColorComponent(), 3),
			func(components []uint8) (RGBColor, error) {
				return RGBColor{components[0], components[1], components[2]}, nil
			},
		),
	)

	newState, output := parser(gomme.NewInputFromString(input))
	if newState.Failed() {
		return RGBColor{}, result.Err
	}

	return output, nil
}

// HexColorComponent produces a parser that parses a single hex color component,
// which is a two digit hexadecimal number.
func HexColorComponent() gomme.Parser[uint8] {
	return func(input gomme.State) gomme.Result[uint8] {
		return gomme.Map1(
			gomme.SatisfyMN(2, 2, gomme.IsHexDigit),
			fromHex,
		)(input)
	}
}

// fromHex converts a two digits hexadecimal number to its decimal value.
func fromHex(input []byte) (uint8, error) {
	res, err := strconv.ParseUint(string(input), 16, 8)
	if err != nil {
		return 0, err
	}

	return uint8(res), nil
}
