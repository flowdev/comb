// Package hexcolor implements a parser for hexadecimal color strings.
// It demonstrates how to use gomme to build a parser for a simple string format.
package hexcolor

import (
	"strconv"

	"github.com/oleiade/gomme"
	. "github.com/oleiade/gomme/cute"
	"github.com/oleiade/gomme/pcb"
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
	parse := pcb.Map4(
		NoWayBack(C('#')),
		HexColorComponent("red hex color"),
		HexColorComponent("green hex color"),
		HexColorComponent("blue hex color"),
		func(_ rune, r, g, b string) (RGBColor, error) {
			return RGBColor{fromHex(r), fromHex(g), fromHex(b)}, nil
		},
	)

	output, err := gomme.RunOnString(-1, nil, -1, input, parse)
	if err != nil {
		return RGBColor{}, err
	}

	return output, nil
}

// HexColorComponent produces a parser that parses a single hex color component,
// which is a two digit hexadecimal number.
func HexColorComponent(expected string) gomme.Parser[string] {
	return NoWayBack(pcb.SatisfyMN(expected, 2, 2, pcb.IsHexDigit))
}

// fromHex converts a two digits hexadecimal number to its decimal value.
func fromHex(input string) uint8 {
	res, _ := strconv.ParseUint(input, 16, 8) // errors have been caught by the parser
	return uint8(res)
}
