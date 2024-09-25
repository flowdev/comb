package main

import (
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/oleiade/gomme"
)

//go:embed test.json
var testJSON string

func main() {
	result := parseJSON(gomme.NewInputFromString(testJSON))
	if result.Err != nil {
		log.Fatal(result.Err)
		return
	}

	fmt.Println(result.Output)
}

type (
	// JSONValue represents any value that can be encountered in
	// JSON, including complex types like objects and arrays.
	JSONValue interface{}

	// JSONString represents a JSON string value.
	JSONString string

	// JSONNumber represents a JSON number value, which internally is treated as float64.
	JSONNumber float64

	// JSONObject represents a JSON object, which is a collection of key-value pairs.
	JSONObject map[string]JSONValue

	// JSONArray represents a JSON array, which is a list of JSON values.
	JSONArray []JSONValue

	// JSONBool represents a JSON boolean value.
	JSONBool bool

	// JSONNull represents the JSON null value.
	JSONNull struct{}
)

// parseJSON is a convenience function to start parsing JSON from the given input string.
func parseJSON(input gomme.Input) gomme.Result[JSONValue] {
	return parseValue(input)
}

// parseValue is a parser that attempts to parse different types of
// JSON values (object, array, string, etc.).
func parseValue(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Alternative(
		parseObject,
		parseArray,
		parseString,
		parseNumber,
		parseTrue,
		parseFalse,
		parseNull,
	)(input)
}

// parseObject parses a JSON object, which starts and ends with
// curly braces and contains key-value pairs.
func parseObject(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1(
		gomme.Delimited[rune, map[string]JSONValue, rune](
			gomme.Char('{'),
			gomme.Optional[map[string]JSONValue](
				gomme.Preceded(
					ws(),
					gomme.Terminated[map[string]JSONValue](
						parseMembers,
						ws(),
					),
				),
			),
			gomme.Char('}'),
		),
		func(members map[string]JSONValue) (JSONValue, error) {
			return JSONObject(members), nil
		},
	)(input)
}

// Ensure parseObject is a Parser[JSONValue]
var _ gomme.Parser[JSONValue] = parseObject

// parseArray parses a JSON array, which starts and ends with
// square brackets and contains a list of values.
func parseArray(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1(
		gomme.Delimited[rune, []JSONValue, rune](
			gomme.Char('['),
			gomme.Alternative(
				parseElements,
				gomme.Map1(ws(), func(s string) ([]JSONValue, error) { return []JSONValue{}, nil }),
			),
			gomme.Char(']'),
		),
		func(elements []JSONValue) (JSONValue, error) {
			return JSONArray(elements), nil
		},
	)(input)
}

// Ensure parseArray is a Parser[string, JSONValue]
var _ gomme.Parser[JSONValue] = parseArray

func parseElement(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1(
		gomme.Delimited[string, JSONValue, string](ws(), parseValue, ws()),
		func(v JSONValue) (JSONValue, error) { return v, nil },
	)(input)
}

// Ensure parseElement is a Parser[JSONValue]
var _ gomme.Parser[JSONValue] = parseElement

// parseNumber parses a JSON number.
func parseNumber(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1[[]string, JSONValue](
		gomme.Sequence(
			gomme.Map1(integer(), func(i int) (string, error) { return strconv.Itoa(i), nil }),
			gomme.Optional(fraction()),
			gomme.Optional(exponent()),
		),
		func(parts []string) (JSONValue, error) {
			// Construct the float string from parts
			var floatStr string

			// Integer part
			floatStr += parts[0]

			// Fraction part
			if parts[1] != "" {
				fractionPart, err := strconv.Atoi(parts[1])
				if err != nil {
					return 0, err
				}

				if fractionPart != 0 {
					floatStr += "." + parts[1]
				}
			}

			// Exponent part
			if parts[2] != "" {
				floatStr += "e" + parts[2]
			}

			f, err := strconv.ParseFloat(floatStr, 64)
			if err != nil {
				return JSONNumber(0.0), err
			}

			return JSONNumber(f), nil
		},
	)(input)
}

// Ensure parseNumber is a Parser[JSONValue]
var _ gomme.Parser[JSONValue] = parseNumber

// parseString parses a JSON string.
func parseString(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1(
		stringParser(),
		func(s string) (JSONValue, error) {
			return JSONString(s), nil
		},
	)(input)
}

// Ensure parseString is a Parser[JSONValue]
var _ gomme.Parser[JSONValue] = parseString

// parseFalse parses the JSON boolean value 'false'.
func parseFalse(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1(
		gomme.Token("false"),
		func(_ []byte) (JSONValue, error) { return JSONBool(false), nil },
	)(input)
}

// Ensure parseFalse is a Parser[JSONValue]
var _ gomme.Parser[JSONValue] = parseFalse

// parseTrue parses the JSON boolean value 'true'.
func parseTrue(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1(
		gomme.Token("true"),
		func(_ []byte) (JSONValue, error) { return JSONBool(true), nil },
	)(input)
}

// Ensure parseTrue is a Parser[JSONValue]
var _ gomme.Parser[JSONValue] = parseTrue

// parseNull parses the JSON 'null' value.
func parseNull(input gomme.Input) gomme.Result[JSONValue] {
	return gomme.Map1(
		gomme.Token("null"),
		func(_ []byte) (JSONValue, error) { return nil, nil },
	)(input)
}

// Ensure parseNull is a Parser[string, JSONValue]
var _ gomme.Parser[JSONValue] = parseNull

// parseElements parses the elements of a JSON array.
func parseElements(input gomme.Input) gomme.Result[[]JSONValue] {
	return gomme.Map1(
		gomme.SeparatedList0[JSONValue, []byte](
			parseElement,
			gomme.Token(","),
			false,
		),
		func(elems []JSONValue) ([]JSONValue, error) {
			return elems, nil
		},
	)(input)
}

// Ensure parseElements is a Parser[[]JSONValue]
var _ gomme.Parser[[]JSONValue] = parseElements

// parseElement parses a single element of a JSON array.
func parseMembers(input gomme.Input) gomme.Result[map[string]JSONValue] {
	return gomme.Map1(
		gomme.SeparatedList0[kv, []byte](
			parseMember,
			gomme.Token(","),
			false,
		),
		func(kvs []kv) (map[string]JSONValue, error) {
			obj := make(JSONObject)
			for _, kv := range kvs {
				obj[kv.key] = kv.value
			}
			return obj, nil
		},
	)(input)
}

// Ensure parseMembers is a Parser[map[string]JSONValue]
var _ gomme.Parser[map[string]JSONValue] = parseMembers

// parseMember parses a single member (key-value pair) of a JSON object.
func parseMember(input gomme.Input) gomme.Result[kv] {
	return member()(input)
}

// Ensure parseMember is a Parser[kv]
var _ gomme.Parser[kv] = parseMember

// member creates a parser for a single key-value pair in a JSON object.
//
// It expects a string followed by a colon and then a JSON value.
// The result is a kv struct with the parsed key and value.
func member() gomme.Parser[kv] {
	mapFunc := func(left string, right JSONValue) (kv, error) {
		return kv{left, right}, nil
	}

	return gomme.Map2(
		gomme.Delimited(ws(), stringParser(), ws()),
		gomme.Preceded(
			gomme.Token(":"),
			element()),
		mapFunc,
	)
}

// element creates a parser for a single element in a JSON array.
//
// It wraps the element with optional whitespace on either side.
func element() gomme.Parser[JSONValue] {
	return gomme.Delimited(ws(), parseValue, ws())
}

// kv is a struct representing a key-value pair in a JSON object.
//
// 'key' holds the string key, and 'value' holds the corresponding JSON value.
type kv struct {
	key   string
	value JSONValue
}

// stringParser creates a parser for a JSON string.
//
// It expects a sequence of characters enclosed in double quotes.
func stringParser() gomme.Parser[string] {
	return gomme.Delimited[rune, string, rune](
		gomme.Char('"'),
		characters(),
		gomme.Char('"'),
	)
}

// integer creates a parser for a JSON number's integer part.
//
// It handles negative and positive integers including zero.
func integer() gomme.Parser[int] {
	return gomme.Alternative(
		// "-" onenine digits
		gomme.Preceded(
			gomme.Token("-"),
			gomme.Map2(
				onenine(), digits(),
				func(first string, rest string) (int, error) {
					return strconv.Atoi(first + rest)
				},
			),
		),

		// onenine digits
		gomme.Map2(
			onenine(), digits(),
			func(first string, rest string) (int, error) {
				return strconv.Atoi(first + rest)
			},
		),

		// "-" digit
		gomme.Preceded(
			gomme.Token("-"),
			gomme.Map1(
				digit(),
				strconv.Atoi,
			),
		),

		// digit
		gomme.Map1(digit(), strconv.Atoi),
	)
}

// digits creates a parser for a sequence of digits.
//
// It concatenates the sequence into a single string.
func digits() gomme.Parser[string] {
	return gomme.Map1(gomme.Many1(digit()), func(digits []string) (string, error) {
		return strings.Join(digits, ""), nil
	})
}

// digit creates a parser for a single digit.
//
// It distinguishes between '0' and non-zero digits.
func digit() gomme.Parser[string] {
	return gomme.Alternative(
		gomme.BytesToString(gomme.Token("0")),
		onenine(),
	)
}

// onenine creates a parser for digits from 1 to 9.
func onenine() gomme.Parser[string] {
	return gomme.BytesToString(gomme.Alternative(
		gomme.Token("1"),
		gomme.Token("2"),
		gomme.Token("3"),
		gomme.Token("4"),
		gomme.Token("5"),
		gomme.Token("6"),
		gomme.Token("7"),
		gomme.Token("8"),
		gomme.Token("9"),
	))
}

// fraction creates a parser for the fractional part of a JSON number.
//
// It expects a dot followed by at least one digit.
func fraction() gomme.Parser[string] {
	return gomme.Preceded(
		gomme.Token("."),
		gomme.Digit1(),
	)
}

// exponent creates a parser for the exponent part of a JSON number.
//
// It handles the exponent sign and the exponent digits.
func exponent() gomme.Parser[string] {
	return gomme.Preceded(
		gomme.Token("e"),
		gomme.Map2(
			sign(), digits(),
			func(sign string, digits string) (string, error) {
				return sign + digits, nil
			},
		),
	)
}

// sign creates a parser for the sign part of a number's exponent.
//
// It can parse both positive ('+') and negative ('-') signs.
func sign() gomme.Parser[string] {
	return gomme.Optional(
		gomme.BytesToString(gomme.Alternative(
			gomme.Token("-"),
			gomme.Token("+"),
		)),
	)
}

// characters creates a parser for a sequence of JSON string characters.
//
// It handles regular characters and escaped sequences.
func characters() gomme.Parser[string] {
	return gomme.Optional(
		gomme.Map1(
			gomme.Many1[rune](character()),
			func(chars []rune) (string, error) {
				return string(chars), nil
			},
		),
	)
}

// character creates a parser for a single JSON string character.
//
// It distinguishes between regular characters and escape sequences.
func character() gomme.Parser[rune] {
	return gomme.Alternative(
		// normal character
		gomme.Satisfy(func(c rune) bool {
			return c != '"' && c != '\\' && c >= 0x20 && c <= 0x10FFFF
		}),

		// escape
		escape(),
	)
}

// escape creates a parser for escaped characters in a JSON string.
//
// It handles common escape sequences like '\n', '\t', etc., and unicode escapes.
func escape() gomme.Parser[rune] {
	mapFunc := func(chars []rune) (rune, error) {
		// chars[0] will always be '\\'
		switch chars[1] {
		case '"':
			return '"', nil
		case '\\':
			return '\\', nil
		case '/':
			return '/', nil
		case 'b':
			return '\b', nil
		case 'f':
			return '\f', nil
		case 'n':
			return '\n', nil
		case 'r':
			return '\r', nil
		case 't':
			return '\t', nil
		default: // for unicode escapes
			return chars[1], nil
		}
	}

	return gomme.Map1(
		gomme.Sequence(
			gomme.Char('\\'),
			gomme.Alternative(
				gomme.Char('"'),
				gomme.Char('\\'),
				gomme.Char('/'),
				gomme.Char('b'),
				gomme.Char('f'),
				gomme.Char('n'),
				gomme.Char('r'),
				gomme.Char('t'),
				unicodeEscape(),
			),
		),
		mapFunc,
	)
}

// unicodeEscape creates a parser for a Unicode escape sequence in a JSON string.
//
// It expects a sequence starting with 'u' followed by four hexadecimal digits and
// converts them to the corresponding rune.
func unicodeEscape() gomme.Parser[rune] {
	mapFunc := func(chars []rune) (rune, error) {
		// chars[0] will always be 'u'
		hex := string(chars[1:5])
		codePoint, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return 0, err
		}
		return rune(codePoint), nil
	}

	return gomme.Map1(
		gomme.Sequence(
			gomme.Char('u'),
			hex(),
			hex(),
			hex(),
			hex(),
		),
		mapFunc,
	)
}

// hex creates a parser for a single hexadecimal digit.
//
// It can parse digits ('0'-'9') as well as
// letters ('a'-'f', 'A'-'F') used in hexadecimal numbers.
func hex() gomme.Parser[rune] {
	return gomme.Satisfy(func(r rune) bool {
		return ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F')
	})
}

// ws creates a parser for whitespace in JSON.
//
// It can handle spaces, tabs, newlines, and carriage returns.
// The parser accumulates all whitespace characters and returns them as a single string.
func ws() gomme.Parser[string] {
	mapFunc := func(runes []rune) (string, error) {
		return string(runes), nil
	}

	return gomme.Map1(gomme.Many0(
		gomme.Satisfy(func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n' || r == '\r'
		}),
	), mapFunc)
}
