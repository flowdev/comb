package main

import (
	_ "embed"
	"fmt"
	"github.com/oleiade/gomme"
	"github.com/oleiade/gomme/pcb"
	"log"
	"strconv"
)

//go:embed test.json
var testJSON string

// break initialization cycle:
func init() {
	element = gomme.LazyParser(elementParser)
	member = gomme.LazyParser(memberParser)
	members = gomme.LazyParser(membersParser)
	elements = gomme.LazyParser(elementsParser)
	objectp = gomme.LazyParser(parseObject)
	arrayp = parseArray()
	valuep = parseValue()
}

func main() {
	output, err := gomme.RunOnString(testJSON, valuep)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	fmt.Println(output)
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
func parseValue() gomme.Parser[JSONValue] {
	return pcb.Alternative(
		objectp,
		arrayp,
		stringp,
		numberp,
		truep,
		falsep,
		nullp,
	)
}

var valuep gomme.Parser[JSONValue]

// parseObject parses a JSON object, which starts and ends with
// curly braces and contains key-value pairs.
func parseObject() gomme.Parser[JSONValue] {
	return pcb.Map(
		pcb.Delimited[rune, map[string]JSONValue, rune](
			pcb.Char('{'),
			pcb.Optional[map[string]JSONValue](
				pcb.Preceded(
					ws,
					pcb.Terminated[map[string]JSONValue](
						members,
						ws,
					),
				),
			),
			pcb.Char('}'),
		),
		func(members map[string]JSONValue) (JSONValue, error) {
			return JSONObject(members), nil
		},
	)
}

var objectp gomme.Parser[JSONValue]

// parseArray parses a JSON array, which starts and ends with
// square brackets and contains a list of values.
func parseArray() gomme.Parser[JSONValue] {
	return pcb.Map(
		pcb.Delimited[rune, []JSONValue, rune](
			pcb.Char('['),
			pcb.Alternative(
				elements,
				pcb.Map(ws, func(s string) ([]JSONValue, error) { return []JSONValue{}, nil }),
			),
			pcb.Char(']'),
		),
		func(elements []JSONValue) (JSONValue, error) {
			return JSONArray(elements), nil
		},
	)
}

var arrayp gomme.Parser[JSONValue]

// parseNumber parses a JSON number.
func parseNumber() gomme.Parser[JSONValue] {
	return pcb.Map[[]string, JSONValue](
		pcb.Sequence(
			pcb.Map(integer, func(i int) (string, error) { return strconv.Itoa(i), nil }),
			pcb.Optional(fraction),
			pcb.Optional(exponent),
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
	)
}

var numberp = parseNumber()

// parseString parses a JSON string.
func parseString() gomme.Parser[JSONValue] {
	return pcb.Map(
		pstring,
		func(s string) (JSONValue, error) {
			return JSONString(s), nil
		},
	)
}

var stringp = parseString()

// parseFalse parses the JSON boolean value 'false'.
func parseFalse() gomme.Parser[JSONValue] {
	return pcb.Map(
		pcb.String("false"),
		func(_ string) (JSONValue, error) { return JSONBool(false), nil },
	)
}

var falsep = parseFalse()

// parseTrue parses the JSON boolean value 'true'.
func parseTrue() gomme.Parser[JSONValue] {
	return pcb.Map(
		pcb.String("true"),
		func(_ string) (JSONValue, error) { return JSONBool(true), nil },
	)
}

var truep = parseTrue()

// parseNull parses the JSON 'null' value.
func parseNull() gomme.Parser[JSONValue] {
	return pcb.Map(
		pcb.String("null"),
		func(_ string) (JSONValue, error) { return nil, nil },
	)
}

var nullp = parseNull()

// elementsParser parses the elements of a JSON array.
func elementsParser() gomme.Parser[[]JSONValue] {
	return pcb.Map(
		pcb.Separated0[JSONValue, string](
			element,
			pcb.String(","),
			false,
		),
		func(elems []JSONValue) ([]JSONValue, error) {
			return elems, nil
		},
	)
}

var elements gomme.Parser[[]JSONValue]

// membersParser parses a single element of a JSON array.
func membersParser() gomme.Parser[map[string]JSONValue] {
	return pcb.Map(
		pcb.Separated0[kv, rune](
			member,
			pcb.Char(','),
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

var members gomme.Parser[map[string]JSONValue]

// member creates a parser for a single key-value pair in a JSON object.
//
// It expects a string followed by a colon and then a JSON value.
// The result is a kv struct with the parsed key and value.
func memberParser() gomme.Parser[kv] {
	mapFunc := func(o1 string, o2 rune, o3 JSONValue) (kv, error) {
		return kv{key: o1, value: o3}, nil
	}

	return pcb.Map3(
		pcb.Delimited(ws, pstring, ws),
		pcb.Char(':'),
		element,
		mapFunc,
	)
}

var member gomme.Parser[kv]

// element creates a parser for a single element in a JSON array.
//
// It wraps the element with optional whitespace on either side.
func elementParser() gomme.Parser[JSONValue] {
	return pcb.Delimited(ws, valuep, ws)
}

var element gomme.Parser[JSONValue]

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
	return pcb.Delimited[rune, string, rune](
		pcb.Char('"'),
		characters,
		pcb.Char('"'),
	)
}

var pstring = stringParser()

// integer creates a parser for a JSON number's integer part.
//
// It handles negative and positive integers including zero.
func integerParser() gomme.Parser[int] {
	return pcb.Alternative(
		// "-" onenine digits
		pcb.Preceded(
			pcb.Char('-'),
			pcb.Map2(
				onenine, digits,
				func(first rune, rest string) (int, error) {
					return strconv.Atoi("-" + string(first) + rest)
				},
			),
		),

		// onenine digits
		pcb.Map2(
			onenine, digits,
			func(first rune, rest string) (int, error) {
				return strconv.Atoi(string(first) + rest)
			},
		),

		// "-" digit
		pcb.Preceded(
			pcb.Char('-'),
			pcb.Map(
				digit,
				func(r rune) (int, error) {
					return strconv.Atoi("-" + string(r))
				},
			),
		),

		// digit
		pcb.Map(
			digit,
			func(r rune) (int, error) {
				return strconv.Atoi(string(r))
			},
		),
	)
}

var integer = integerParser()

// digits creates a parser for a sequence of digits.
//
// It concatenates the sequence into a single string.
func digitsParser() gomme.Parser[string] {
	return pcb.Map(pcb.Many1(digit), func(digits []rune) (string, error) {
		return string(digits), nil
	})
}

var digits = digitsParser()

// digit creates a parser for a single digit.
//
// It distinguishes between '0' and non-zero digits.
func digitParser() gomme.Parser[rune] {
	return pcb.Alternative(
		pcb.Char('0'),
		onenine,
	)
}

var digit = digitParser()

// onenine creates a parser for digits from 1 to 9.
func onenineParser() gomme.Parser[rune] {
	return pcb.OneOf('1', '2', '3', '4', '5', '6', '7', '8', '9')
}

var onenine = onenineParser()

// fraction creates a parser for the fractional part of a JSON number.
//
// It expects a dot followed by at least one digit.
func fractionParser() gomme.Parser[string] {
	return pcb.Preceded(
		pcb.String("."),
		pcb.Digit1(),
	)
}

var fraction = fractionParser()

// exponent creates a parser for the exponent part of a JSON number.
//
// It handles the exponent sign and the exponent digits.
func exponentParser() gomme.Parser[string] {
	return pcb.Preceded(
		pcb.String("e"),
		pcb.Map2(
			sign, digits,
			func(sign string, digits string) (string, error) {
				return sign + digits, nil
			},
		),
	)
}

var exponent = exponentParser()

// sign creates a parser for the sign part of a number's exponent.
//
// It can parse both positive ('+') and negative ('-') signs.
func signParser() gomme.Parser[string] {
	return pcb.Optional(
		pcb.Alternative(
			pcb.String("-"),
			pcb.String("+"),
		),
	)
}

var sign = signParser()

// characters creates a parser for a sequence of JSON string characters.
//
// It handles regular characters and escaped sequences.
func charactersParser() gomme.Parser[string] {
	return pcb.Optional(
		pcb.Map(
			pcb.Many1[rune](character),
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
func characterParser() gomme.Parser[rune] {
	return pcb.Alternative(
		pcb.Satisfy("normal character", func(c rune) bool {
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
func escapeParser() gomme.Parser[rune] {
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

	return pcb.Map(
		pcb.Sequence(
			pcb.Char('\\'),
			pcb.Alternative(
				pcb.OneOf('"', '\\', '/', 'b', 'f', 'n', 'r', 't'),
				unicodeEscape,
			),
		),
		mapFunc,
	)
}

var escape = escapeParser()

// unicodeEscape creates a parser for a Unicode escape sequence in a JSON string.
//
// It expects a sequence starting with 'u' followed by four hexadecimal digits and
// converts them to the corresponding rune.
func unicodeEscapeParser() gomme.Parser[rune] {
	mapFunc := func(chars []rune) (rune, error) {
		// chars[0] will always be 'u'
		hex := string(chars[1:5])
		codePoint, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return 0, err
		}
		return rune(codePoint), nil
	}

	return pcb.Map(
		pcb.Sequence(
			pcb.Char('u'),
			hex,
			hex,
			hex,
			hex,
		),
		mapFunc,
	)
}

var unicodeEscape = unicodeEscapeParser()

// hex creates a parser for a single hexadecimal digit.
//
// It can parse digits ('0'-'9') as well as
// letters ('a'-'f', 'A'-'F') used in hexadecimal numbers.
func hexParser() gomme.Parser[rune] {
	return pcb.Satisfy("hex digit", func(r rune) bool {
		return ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F')
	})
}

var hex = hexParser()

// ws creates a parser for whitespace in JSON.
//
// It can handle spaces, tabs, newlines, and carriage returns.
// The parser accumulates all whitespace characters and returns them as a single string.
func wsParser() gomme.Parser[string] {
	mapFunc := func(runes []rune) (string, error) {
		return string(runes), nil
	}

	return pcb.Map(pcb.Many0(
		pcb.OneOf(' ', '\t', '\n', '\r'),
	), mapFunc)
}

var ws = wsParser()
