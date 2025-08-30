package cmb

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/flowdev/comb"
)

// ============================================================================
// Parse Integer Numbers
//

// Integer parses any kind of integer number.
// `signAllowed` can be false to parse only unsigned integers.
// `radix` can be 0 to honor prefixes "0x", "0X", "0b", "0B", "0o", "0O" and "0"
// according to the Go language specification.
// `underscoreAllowed` can be true to allow '_' characters.
// No check on position or number of (consecutive) underscores is done.
// The Go parse functions will do more checks on this.
func Integer(signAllowed bool, base int, underscoreAllowed bool) comb.Parser[string] {
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

	parser := func(state comb.State) (comb.State, string, *comb.ParserError) {
		fullInput := state.CurrentString()
		input := fullInput
		if input == "" {
			return state, "", state.NewSyntaxError(expected + " at EOF")
		}

		n := 0 // number of bytes read from input

		// Pick off the leading sign.
		if signAllowed {
			if input[0] == '+' || input[0] == '-' {
				input = input[1:]
				n = 1
				if input == "" {
					return state, "", state.NewSyntaxError(expected + " at EOF")
				}
			}
		}

		input, base, n = rebaseInt(input, base, n)
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
	return comb.NewParser[string](expected, parser, IndexOfAny(allRunes[:recovererBase]...))
}

func rebaseInt(input string, base, n int) (string, int, int) {
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
func Int64(signAllowed bool, base int) comb.Parser[int64] {
	var p comb.Parser[int64]

	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	intParser := Integer(signAllowed, base, underscoreAllowed)

	parser := func(state comb.State) (comb.State, int64, *comb.ParserError) {
		nState, out, pErr := intParser.ParseAny(p.ID(), state)
		str, _ := out.(string)
		if pErr != nil {
			return state, 0, comb.ClaimError(pErr)
		}
		i, err := strconv.ParseInt(str, base, 64)
		if err != nil {
			return nState, i, state.NewSemanticError(err.Error())
		}
		return nState, i, nil
	}
	p = comb.NewParser[int64](intParser.Expected(), parser, intParser.Recover)
	return p
}

// UInt64 parses an integer from the input using `strconv.ParseUint`.
func UInt64(signAllowed bool, base int) comb.Parser[uint64] {
	var p comb.Parser[uint64]

	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	intParser := Integer(signAllowed, base, underscoreAllowed)

	parser := func(state comb.State) (comb.State, uint64, *comb.ParserError) {
		nState, out, pErr := intParser.ParseAny(p.ID(), state)
		str, _ := out.(string)
		if pErr != nil {
			return state, 0, comb.ClaimError(pErr)
		}
		ui, err := strconv.ParseUint(str, base, 64)
		if err != nil {
			return nState, ui, state.NewSemanticError(err.Error())
		}
		return nState, ui, nil
	}
	p = comb.NewParser[uint64](intParser.Expected(), parser, intParser.Recover)
	return p
}

// ============================================================================
// Parse Floating Point Numbers
//

// Float parses any kind of floating point number.
// `signAllowed` can be false to parse only unsigned numbers.
// `radix` can be 0 to honor prefixes "0x" and "0X"
// according to the Go language specification.
// `underscoreAllowed` can be true to allow '_' characters.
// No check on position or number of (consecutive) underscores is done.
// The Go parse functions will do more checks on this.
func Float(signAllowed bool, base int, underscoreAllowed bool) comb.Parser[string] {
	if base != 0 && base != 10 && base != 16 {
		panic(fmt.Sprintf("The base has to be 0, 10 or 16, but is: %d", base))
	}
	expected := ""
	switch base {
	case 0:
		expected = "Go float"
	case 10:
		expected = "decimal float"
	default:
		expected = "hexadecimal float"
	}

	const allDigits = "0123456789abcdef"

	parser := func(state comb.State) (comb.State, string, *comb.ParserError) {
		input := state.CurrentString()
		if input == "" {
			return state, "", state.NewSyntaxError(expected + " at EOF")
		}

		n := 0 // number of bytes read from input

		// Pick off the leading sign.
		if signAllowed {
			if input[0] == '+' || input[0] == '-' {
				n = 1
				if len(input) <= 1 {
					return state, "", state.NewSyntaxError(expected + " at EOF")
				}
			}
		}

		m := 0
		base, m = rebaseFloat(input[n:], base)
		n += m
		digits := allDigits[:base]
		good := false
		digit := ' '

		digit, m, good = readDigits(input[n:], underscoreAllowed, digits)
		if !good && digit != '.' {
			return state, "", state.NewSyntaxError("%s found '%c'", expected, digit)
		}
		n += m
		hasDigits := good

		if digit == '.' {
			n++
			digit, m, good = readDigits(input[n:], underscoreAllowed, digits)
			if !good && !hasDigits {
				return state, "", state.NewSyntaxError("%s found '%c'", expected, digit)
			}
			n += m
		}

		if (base == 10 && (digit == 'e' || digit == 'E')) ||
			(base == 16 && (digit == 'p' || digit == 'P')) {

			n++
			digit, m, good = readDigits(input[n:], underscoreAllowed, allDigits[:10])
			if !good {
				return state, "", state.NewSyntaxError("%s found '%c'", expected, digit)
			}
			n += m
		}

		return state.MoveBy(n), input[:n], nil
	}

	recovererBase := base
	if base == 0 {
		recovererBase = 10 // best guess
	}
	return comb.NewParser[string](expected, parser, indexOfFloat(allDigits[:recovererBase]))
}
func rebaseFloat(input string, base int) (int, int) {
	if base != 0 {
		return base, 0
	}
	baseChar := ' ' // set to impossible value
	if len(input) >= 3 {
		baseChar = rune(input[1]) // only ASCII digits are supported
	}
	if input[0] == '0' && len(input) >= 3 && (baseChar == 'x' || baseChar == 'X') {
		return 16, 2
	}
	return 10, 0
}
func readDigits(input string, underscoreAllowed bool, digits string) (int32, int, bool) {
	digit := ' '
	good := false
	n := 0

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
	return digit, n, good
}

func indexOfFloat(digits string) func(comb.State, interface{}) (int, interface{}) {
	dotDigits := "." + digits
	return func(state comb.State, data interface{}) (int, interface{}) {
		input := state.CurrentString()
		i := strings.IndexAny(input, dotDigits)
		if i < 0 || strings.ContainsRune(digits, rune(input[i])) {
			return i, nil
		}
		if len(input) > i+1 && strings.ContainsRune(digits, rune(input[i+1])) {
			return i, nil
		}
		return comb.RecoverWasteTooMuch, nil
	}
}

// Float64 parses a floating point number from the input using `strconv.ParseFloat`.
func Float64(signAllowed bool, base int) comb.Parser[float64] {
	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	floatParser := Float(signAllowed, base, underscoreAllowed)

	parser := func(state comb.State) (comb.State, float64, *comb.ParserError) {
		nState, out, pErr := floatParser.ParseAny(0, state)
		str, _ := out.(string)
		if pErr != nil {
			return state, 0, comb.ClaimError(pErr)
		}

		if base == 16 {
			if str[0] == '-' || str[0] == '+' {
				str = string(str[0]) + "0x" + str[1:]
			} else {
				str = "0x" + str
			}
		}
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nState, f, state.NewSemanticError(err.Error())
		}
		return nState, f, nil
	}
	return comb.NewParser[float64](floatParser.Expected(), parser, floatParser.Recover)
}
