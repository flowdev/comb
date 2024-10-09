// Package hexcolor implements a parser for hexadecimal color strings.
// It demonstrates how to use gomme to build a parser for a simple string format.
package hexcolor

import (
	"github.com/oleiade/gomme/pcb"
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
	parse := pcb.Preceded(
		pcb.Char('#'),
		pcb.Map(
			pcb.Count(HexColorComponent(), 3),
			func(components []uint8) (RGBColor, error) {
				return RGBColor{components[0], components[1], components[2]}, nil
			},
		),
	)

	output, err := gomme.RunOnString(0, input, parse)
	if err != nil {
		return RGBColor{}, err
	}

	return output, nil
}

// HexColorComponent produces a parser that parses a single hex color component,
// which is a two digit hexadecimal number.
func HexColorComponent() gomme.Parser[uint8] {
	return pcb.Map(
		pcb.SatisfyMN("hex digit", 2, 2, pcb.IsHexDigit),
		fromHex,
	)
}

// fromHex converts a two digits hexadecimal number to its decimal value.
func fromHex(input string) (uint8, error) {
	res, err := strconv.ParseUint(input, 16, 8)
	if err != nil {
		return 0, err
	}

	return uint8(res), nil
}
