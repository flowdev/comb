package json

import (
	"strconv"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
	. "github.com/flowdev/comb/cute"
)

// break initialization cycle:
func init() {
	element = comb.LazyBranchParser(elementParser)
	member = comb.LazyBranchParser(memberParser)
	members = comb.LazyBranchParser(membersParser)
	elements = comb.LazyBranchParser(elementsParser)
	objectp = comb.LazyBranchParser(parseObject)
	arrayp = parseArray()
	valuep = parseValue()
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

// parseValue is a parser that attempts to parse different types of
// JSON values (object, array, string, etc.).
func parseValue() comb.Parser[JSONValue] {
	return cmb.FirstSuccessful(
		objectp,
		arrayp,
		stringp,
		numberp,
		truep,
		falsep,
		nullp,
	)
}

var valuep comb.Parser[JSONValue]

// parseObject parses a JSON object, which starts and ends with
// curly braces and contains key-value pairs.
func parseObject() comb.Parser[JSONValue] {
	return cmb.Map(
		cmb.Delimited[rune, map[string]JSONValue, rune](
			cmb.Char('{'),
			cmb.Optional[map[string]JSONValue](
				cmb.Prefixed(
					ws,
					cmb.Suffixed[map[string]JSONValue](
						members,
						ws,
					),
				),
			),
			cmb.Char('}'),
		),
		func(members map[string]JSONValue) (JSONValue, error) {
			return JSONObject(members), nil
		},
	)
}

var objectp comb.Parser[JSONValue]

// parseArray parses a JSON array, which starts and ends with
// square brackets and contains a list of values.
func parseArray() comb.Parser[JSONValue] {
	return cmb.Map(
		cmb.Delimited[rune, []JSONValue, rune](
			cmb.Char('['),
			cmb.FirstSuccessful(
				elements,
				cmb.Map(ws, func(s string) ([]JSONValue, error) { return []JSONValue{}, nil }),
			),
			cmb.Char(']'),
		),
		func(elements []JSONValue) (JSONValue, error) {
			return JSONArray(elements), nil
		},
	)
}

var arrayp comb.Parser[JSONValue]

// parseNumber parses a JSON number.
func parseNumber() comb.Parser[JSONValue] {
	return cmb.Map3[string, string, string, JSONValue](
		integer,
		cmb.Optional(fraction),
		cmb.Optional(exponent),
		func(part1, part2, part3 string) (JSONValue, error) {
			// Construct the float string from parts
			var floatStr string

			// Integer part
			floatStr = part1

			// Fraction part
			if part2 != "" {
				floatStr += "." + part2
			}

			// Exponent part
			if part3 != "" {
				floatStr += "e" + part3
			}

			f, err := strconv.ParseFloat(floatStr, 64)
			if err != nil {
				return JSONNumber(0.0), err
			}

			return JSONNumber(f), nil
		},
	)
}

var numberp = parseNumber()

// parseString parses a JSON string.
func parseString() comb.Parser[JSONValue] {
	return cmb.Map(
		pstring,
		func(s string) (JSONValue, error) {
			return JSONString(s), nil
		},
	)
}

var stringp = parseString()

// parseFalse parses the JSON boolean value 'false'.
func parseFalse() comb.Parser[JSONValue] {
	return cmb.Map(
		cmb.String("false"),
		func(_ string) (JSONValue, error) { return JSONBool(false), nil },
	)
}

var falsep = parseFalse()

// parseTrue parses the JSON boolean value 'true'.
func parseTrue() comb.Parser[JSONValue] {
	return cmb.Map(
		cmb.String("true"),
		func(_ string) (JSONValue, error) { return JSONBool(true), nil },
	)
}

var truep = parseTrue()

// parseNull parses the JSON 'null' value.
func parseNull() comb.Parser[JSONValue] {
	return cmb.Map(
		cmb.String("null"),
		func(_ string) (JSONValue, error) { return nil, nil },
	)
}

var nullp = parseNull()

// elementsParser parses the elements of a JSON array.
func elementsParser() comb.Parser[[]JSONValue] {
	return cmb.Map(
		cmb.Separated0[JSONValue, string](
			element,
			cmb.String(","),
			false,
		),
		func(elems []JSONValue) ([]JSONValue, error) {
			return elems, nil
		},
	)
}

var elements comb.Parser[[]JSONValue]

// membersParser parses a single element of a JSON array.
func membersParser() comb.Parser[map[string]JSONValue] {
	return cmb.Map(
		cmb.Separated0[kv, rune](
			member,
			cmb.Char(','),
			false,
		),
		func(kvs []kv) (map[string]JSONValue, error) {
			obj := make(JSONObject)
			for _, kv := range kvs {
				obj[kv.key] = kv.value
			}
			return obj, nil
		},
	)
}

var members comb.Parser[map[string]JSONValue]

// member creates a parser for a single key-value pair in a JSON object.
//
// It expects a string followed by a colon and then a JSON value.
// The result is a kv struct with the parsed key and value.
func memberParser() comb.Parser[kv] {
	mapFunc := func(o1 string, o2 rune, o3 JSONValue) (kv, error) {
		return kv{key: o1, value: o3}, nil
	}

	return cmb.Map3(
		cmb.Delimited(ws, pstring, ws),
		cmb.Char(':'),
		element,
		mapFunc,
	)
}

var member comb.Parser[kv]

// element creates a parser for a single element in a JSON array.
//
// It wraps the element with optional whitespace on either side.
func elementParser() comb.Parser[JSONValue] {
	return cmb.Delimited(ws, valuep, ws)
}

var element comb.Parser[JSONValue]

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
func stringParser() comb.Parser[string] {
	return cmb.Delimited[rune, string, rune](
		cmb.Char('"'),
		characters,
		cmb.Char('"'),
	)
}

var pstring = stringParser()

// integerParser creates a parser for a JSON number's integer part.
//
// It handles negative and positive integers including zero.
func integerParser() comb.Parser[string] {
	return cmb.FirstSuccessful(
		// "-" onenine digits
		cmb.Prefixed(
			cmb.Char('-'),
			cmb.Map2(
				onenine, digits,
				func(first rune, rest string) (string, error) {
					return "-" + string(first) + rest, nil
				},
			),
		),

		// onenine digits
		cmb.Map2(
			onenine, digits,
			func(first rune, rest string) (string, error) {
				return string(first) + rest, nil
			},
		),

		// "-" digit
		cmb.Prefixed(
			cmb.Char('-'),
			cmb.Map(
				digit,
				func(r rune) (string, error) {
					return "-" + string(r), nil
				},
			),
		),

		// digit
		cmb.Map(
			digit,
			func(r rune) (string, error) {
				return string(r), nil
			},
		),
	)
}

var integer = integerParser()

// digits creates a parser for a sequence of digits.
//
// It concatenates the sequence into a single string.
func digitsParser() comb.Parser[string] {
	return cmb.Map(cmb.Many1(digit), func(digits []rune) (string, error) {
		return string(digits), nil
	})
}

var digits = digitsParser()

// digit creates a parser for a single digit.
//
// It distinguishes between '0' and non-zero digits.
func digitParser() comb.Parser[rune] {
	return cmb.FirstSuccessful(
		cmb.Char('0'),
		onenine,
	)
}

var digit = digitParser()

// onenine creates a parser for digits from 1 to 9.
func onenineParser() comb.Parser[rune] {
	return cmb.OneOfRunes('1', '2', '3', '4', '5', '6', '7', '8', '9')
}

var onenine = onenineParser()

// fraction creates a parser for the fractional part of a JSON number.
//
// It expects a dot followed by at least one digit.
func fractionParser() comb.Parser[string] {
	return cmb.Prefixed(
		cmb.String("."),
		cmb.Digit1(),
	)
}

var fraction = fractionParser()

// exponent creates a parser for the exponent part of a JSON number.
//
// It handles the exponent sign and the exponent digits.
func exponentParser() comb.Parser[string] {
	return cmb.Prefixed(
		cmb.String("e"),
		cmb.Map2(
			sign, digits,
			func(sign rune, digits string) (string, error) {
				return string(sign) + digits, nil
			},
		),
	)
}

var exponent = exponentParser()

// sign creates a parser for the sign part of a number's exponent.
//
// It can parse both positive ('+') and negative ('-') signs.
func signParser() comb.Parser[rune] {
	return cmb.Optional(
		OneOfRunes('-', '+'),
	)
}

var sign = signParser()

// characters creates a parser for a sequence of JSON string characters.
//
// It handles regular characters and escaped sequences.
func charactersParser() comb.Parser[string] {
	return cmb.Optional(
		cmb.Map(
			cmb.Many1[rune](character),
			func(chars []rune) (string, error) {
				return string(chars), nil
			},
		),
	)
}

var characters = charactersParser()

// character creates a parser for a single JSON string character.
//
// It distinguishes between regular characters and escape sequences.
func characterParser() comb.Parser[rune] {
	return cmb.FirstSuccessful(
		cmb.Satisfy("normal character", func(c rune) bool {
			return c != '"' && c != '\\' && c >= 0x20 && c <= 0x10FFFF
		}),

		// escape
		escape,
	)
}

var character = characterParser()

// escape creates a parser for escaped characters in a JSON string.
//
// It handles common escape sequences like '\n', '\t', etc., and unicode escapes.
func escapeParser() comb.Parser[rune] {
	mapFunc := func(char1, char2 rune) (rune, error) {
		// char1 will always be '\\'
		switch char2 {
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
			return char2, nil
		}
	}

	return cmb.Map2(
		cmb.Char('\\'),
		cmb.FirstSuccessful(
			cmb.OneOfRunes('"', '\\', '/', 'b', 'f', 'n', 'r', 't'),
			unicodeEscape,
		),
		mapFunc,
	)
}

var escape = escapeParser()

// unicodeEscape creates a parser for a Unicode escape sequence in a JSON string.
//
// It expects a sequence starting with 'u' followed by four hexadecimal digits and
// converts them to the corresponding rune.
func unicodeEscapeParser() comb.Parser[rune] {
	mapFunc := func(_ rune, hex string) (rune, error) {
		// char will always be 'u'
		codePoint, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return 0, err
		}
		return rune(codePoint), nil
	}

	return cmb.Map2(
		cmb.Char('u'),
		cmb.SatisfyMN("hex digit", 4, 4, cmb.IsHexDigit),
		mapFunc,
	)
}

var unicodeEscape = unicodeEscapeParser()

// hex creates a parser for a single hexadecimal digit.
//
// It can parse digits ('0'-'9') as well as
// letters ('a'-'f', 'A'-'F') used in hexadecimal numbers.
func hexParser() comb.Parser[rune] {
	return cmb.Satisfy("hex digit", func(r rune) bool {
		return ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F')
	})
}

var hex = hexParser()

// ws creates a parser for whitespace in JSON.
//
// It can handle spaces, tabs, newlines, and carriage returns.
// The parser accumulates all whitespace characters and returns them as a single string.
func wsParser() comb.Parser[string] {
	mapFunc := func(runes []rune) (string, error) {
		return string(runes), nil
	}

	return cmb.Map(cmb.Many0(
		cmb.OneOfRunes(' ', '\t', '\n', '\r'),
	), mapFunc)
}

var ws = wsParser()
