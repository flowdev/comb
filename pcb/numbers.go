package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
	"strconv"
	"strings"
	"unicode"
)

// import "math"

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

	parse := func(state gomme.State) (gomme.State, string) {
		fullInput := state.CurrentString()
		input := fullInput
		if input == "" {
			return state.NewError(expected + " at EOF"), ""
		}

		n := 0 // number of bytes read from input

		// Pick off leading sign.
		if signAllowed {
			if input[0] == '+' || input[0] == '-' {
				input = input[1:]
				n = 1
				if input == "" {
					return state.NewError(expected + " at EOF"), ""
				}
			}
		}

		input, base, n = rebaseInput(input, base, n)
		digits := allDigits[:base]
		good := false

	ForLoop:
		for _, digit := range input {
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
			return state.NewError(expected), ""
		}
		return state.MoveBy(n), fullInput[:n]

	}

	recovererBase := base
	if base == 0 {
		recovererBase = 10
	}
	allRunes := digitsToRunes(allDigits)
	return gomme.NewParser[string](expected, parse, false,
		IndexOfAny(allRunes[:recovererBase]...), nil)
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
	return Map(Integer(signAllowed, base, underscoreAllowed), func(integer string) (int64, error) {
		return strconv.ParseInt(integer, base, 64)
	})
}

// Int8 parses an integer from the input using `strconv.ParseInt`.
func Int8(signAllowed bool, base int) gomme.Parser[int8] {
	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	return Map(Integer(signAllowed, base, underscoreAllowed), func(integer string) (int8, error) {
		i, err := strconv.ParseInt(integer, base, 8)
		return int8(i), err
	})
}

// UInt8 parses an integer from the input using `strconv.ParseUint`.
func UInt8(signAllowed bool, base int) gomme.Parser[uint8] {
	underscoreAllowed := false
	if base == 0 {
		underscoreAllowed = true
	}
	return Map(Integer(signAllowed, base, underscoreAllowed), func(integer string) (uint8, error) {
		if integer[0] == '+' {
			integer = integer[1:]
		}
		ui, err := strconv.ParseUint(integer, base, 8)
		return uint8(ui), err
	})
}

// Float parses a sequence of numerical characters into a float64.
// The '.' character is used as the optional decimal delimiter. Any
// number without a decimal part will still be parsed as a float64.
//
// N.B: it is not the parser's role to make sure the floating point
// number you're attempting to parse fits into a 64 bits float.

// func Float[I bytes]() Parser[I, float64] {
// 	digitsParser := TakeWhileOneOf[I]([]rune("0123456789")...)
// 	minusParser := Char[I]('-')
// 	dotParser := Char[I]('.')

// 	return func(input I) Result[float64, I] {
// 		var negative bool

// 		minusresult := minusParser(input)
// 		if !result.Failed() {
// 			negative = true
// 		}

// 		result = digitsParser(result.Remaining)
// 		// result = Expect(digitsParser, "digits")(result.Remaining)
// 		// if result.Failed() {
// 		// 	return result
// 		// }

// 		parsed, ok := result.Output.(string)
// 		if !ok {
// 			err := fmt.Errorf("failed parsing floating point value; " +
// 				"reason: converting Float() parser result's output to string failed",
// 			)
// 			return Preserve(NewFatalError(input, err, "float"), input)
// 		}
// 		if resultTest := dotParser(result.Remaining); resultTest.Err == nil {
// 			if resultTest = digitsParser(resultTest.Remaining); resultTest.Err == nil {
// 				parsed = parsed + "." + resultTest.Output.(string)
// 				result = resultTest
// 			}
// 		}

// 		floatingPointValue, err := strconv.ParseFloat(parsed, 64)
// 		if err != nil {
// 			err = fmt.Errorf("failed to parse '%v' as float; reason: %w", parsed, err)
// 			return Preserve(NewFatalError(input, err), input)
// 		}

// 		if negative {
// 			floatingPointValue = -floatingPointValue
// 		}

// 		result.Output = floatingPointValue

// 		return result
// 	}
// }
