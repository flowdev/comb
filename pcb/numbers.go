package pcb

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/oleiade/gomme"
)

// Integer parses any kind of integer number.
// `signAllowed` can be false to parse only unsigned integers.
// `radix` can be 0 to honor prefixes "0x", "0X", "0b", "0B", "0o", "0O" and "0"
// according to the Go language specification.
// `underscoreAllowed` can be true to allow '_' characters.
// No check on position or number of (consecutive) underscores is done.
// The Go parse functions will do more checks on this.
func Integer(signAllowed bool, base int, underscoreAllowed bool) gomme.Parser[string] {
	if base != 0 && (base < 2 || base > 36) {
		panic(fmt.Sprintf(
			"The base has to be 0 or between 2 and 36, but is: %d", base,
		))
	}
	expected := ""
	switch base {
	case 0:
		expected = "Go integer"
	case 2:
		expected = "binary integer"
	case 8:
		expected = "octal integer"
	case 10:
		expected = "decimal integer"
	case 16:
		expected = "hexadecimal integer"
	default:
		expected = fmt.Sprintf("integer of base %d", base)
	}

	const allDigits = "0123456789abcdefghijklmnopqrstuvwxyz"

	parser := func(state gomme.State) (gomme.State, string, *gomme.ParserError) {
		fullInput := state.CurrentString()
		input := fullInput
		if input == "" {
			return state, "", state.NewSyntaxError(expected + " at EOF")
		}

		n := 0 // number of bytes read from input

		// Pick off leading sign.
		if signAllowed {
			if input[0] == '+' || input[0] == '-' {
				input = input[1:]
				n = 1
				if input == "" {
					return state, "", state.NewSyntaxError(expected + " at EOF")
				}
			}
		}

		input, base, n = rebaseInput(input, base, n)
		digits := allDigits[:base]
		good := false
		digit := ' '

	ForLoop:
		for _, digit = range input {
			switch {
			case digit == '_':
				if !underscoreAllowed {
					break ForLoop // don't break switch but for
				}
				n++
			case strings.IndexRune(digits, unicode.ToLower(digit)) >= 0:
				n++
				good = true
			default:
				break ForLoop // don't break switch but for
			}
		}

		if !good {
			return state, "", state.NewSyntaxError("%s found '%c'", expected, digit)
		}
		return state.MoveBy(n), fullInput[:n], nil
	}

	recovererBase := base
	if base == 0 {
		recovererBase = 10
	}
	allRunes := digitsToRunes(allDigits)
	return gomme.NewParser[string](expected, parser, IndexOfAny(allRunes[:recovererBase]...))
}

func rebaseInput(input string, base, n int) (string, int, int) {
	if base != 0 {
		return input, base, n
	}
	baseChar := ' ' // set to impossible value
	if len(input) >= 3 {
		baseChar = rune(input[1])
	}
	base = 10
	if input[0] == '0' { // Look for prefix.
		switch {
		case len(input) >= 3 && (baseChar == 'b' || baseChar == 'B'):
			base = 2
			input = input[2:]
			n += 2
		case len(input) >= 3 && (baseChar == 'o' || baseChar == 'O'):
			base = 8
			input = input[2:]
			n += 2
		case len(input) >= 3 && (baseChar == 'x' || baseChar == 'X'):
			base = 16
			input = input[2:]
			n += 2
		default:
			base = 8
			input = input[1:]
			n++
		}
	}
	return input, base, n
}

func digitsToRunes(digits string) []rune {
	runes := make([]rune, len(digits))
	for i, d := range []byte(digits) { // it's all ASCII
		runes[i] = rune(d)
	}
	return runes
}

// Int64 parses an integer from the input using `strconv.ParseInt`.
func Int64(signAllowed bool, base int) gomme.Parser[int64] {
	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	intParser := Integer(signAllowed, base, underscoreAllowed)

	parser := func(state gomme.State) (gomme.State, int64, *gomme.ParserError) {
		nState, str, pErr := intParser.Parse(state)
		if pErr != nil {
			return state, 0, gomme.ClaimError(pErr)
		}
		i, err := strconv.ParseInt(str, base, 64)
		if err != nil {
			return nState, i, state.NewSemanticError(err.Error())
		}
		return nState, i, nil
	}
	return gomme.NewParser[int64](intParser.Expected(), parser, intParser.Recover)
}

// Int8 parses an integer from the input using `strconv.ParseInt`.
func Int8(signAllowed bool, base int) gomme.Parser[int8] {
	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	intParser := Integer(signAllowed, base, underscoreAllowed)

	parser := func(state gomme.State) (gomme.State, int8, *gomme.ParserError) {
		nState, str, pErr := intParser.Parse(state)
		if pErr != nil {
			return state, 0, gomme.ClaimError(pErr)
		}
		i, err := strconv.ParseInt(str, base, 8)
		if err != nil {
			return nState, int8(i), state.NewSemanticError(err.Error())
		}
		return nState, int8(i), nil
	}
	return gomme.NewParser[int8](intParser.Expected(), parser, intParser.Recover)
}

// UInt8 parses an integer from the input using `strconv.ParseUint`.
func UInt8(signAllowed bool, base int) gomme.Parser[uint8] {
	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	intParser := Integer(signAllowed, base, underscoreAllowed)

	parser := func(state gomme.State) (gomme.State, uint8, *gomme.ParserError) {
		nState, str, pErr := intParser.Parse(state)
		if pErr != nil {
			return state, 0, gomme.ClaimError(pErr)
		}
		ui, err := strconv.ParseUint(str, base, 8)
		if err != nil {
			return nState, uint8(ui), state.NewSemanticError(err.Error())
		}
		return nState, uint8(ui), nil
	}
	return gomme.NewParser[uint8](intParser.Expected(), parser, intParser.Recover)
}
