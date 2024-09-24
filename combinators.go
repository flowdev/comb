// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

type InputBytes struct {
	// Go is fundamentally working with bytes and can interpret them as strings or as containing runes.
	// There are no standard library functions for handling []rune or the like.
	Bytes []byte
	Text  bool
	Pos   uint // position in the sequence a.k.a. the *byte* index
}

func NewFromString(input string) InputBytes {
	return InputBytes{
		Bytes: []byte(input),
		Text:  true,
		Pos:   0,
	}
}

func NewFromBytes(input []byte) InputBytes {
	return InputBytes{
		Bytes: input,
		Text:  false,
		Pos:   0,
	}
}

func (i InputBytes) AtEnd() bool {
	return i.Pos >= uint(len(i.Bytes))
}

func (i InputBytes) BytesRemaining() uint {
	return uint(len(i.Bytes)) - i.Pos
}

func (i InputBytes) CurrentString() string {
	return string(i.Bytes[i.Pos:])
}

func (i InputBytes) CurrentBytes() []byte {
	return i.Bytes[i.Pos:]
}

func (i InputBytes) StringTo(remaining InputBytes) string {
	return string(i.BytesTo(remaining))
}

func (i InputBytes) BytesTo(remaining InputBytes) []byte {
	if remaining.Pos < i.Pos {
		return []byte{}
	}
	if remaining.Pos > uint(len(i.Bytes)) {
		return i.Bytes[i.Pos:]
	}
	return i.Bytes[i.Pos:remaining.Pos]
}

func (i InputBytes) MoveBy(countBytes uint) InputBytes {
	i2 := i
	i2.Pos += countBytes
	ulen := uint(len(i.Bytes))
	if i2.Pos > ulen { // prevent overrun
		i2.Pos = ulen
	}
	return i2
}

// Separator is a generic type alias for separators (byte, rune, []byte or string)
type Separator interface {
	rune | byte | string | []byte
}

// Result is a generic parser result
type Result[Output any] struct {
	Output    Output
	Err       *Error
	Remaining InputBytes
}

// Parser defines the type of a generic Parser function
type Parser[Output any] func(input InputBytes) Result[Output]

// Success creates a Result with an output set from
// the result of successful parsing.
func Success[Output any](output Output, r InputBytes) Result[Output] {
	return Result[Output]{Output: output, Err: nil, Remaining: r}
}

// Failure creates a Result with an error set from
// the result of failed parsing.
func Failure[Output any](err *Error, input InputBytes) Result[Output] {
	var output Output
	return Result[Output]{Output: output, Err: err, Remaining: input}
}

// Map applies a function to the result of a parser.
func Map[ParserOutput any, MapperOutput any](parse Parser[ParserOutput], fn func(ParserOutput) (MapperOutput, error)) Parser[MapperOutput] {
	return func(input InputBytes) Result[MapperOutput] {
		res := parse(input)
		if res.Err != nil {
			return Failure[MapperOutput](NewError(input, "Map"), input)
		}

		output, err := fn(res.Output)
		if err != nil {
			return Failure[MapperOutput](NewError(input, err.Error()), input)
		}

		return Success(output, res.Remaining)
	}
}

// Optional applies an optional child parser. Will return nil
// if not successful.
//
// N.B: unless a FatalError or NoWayBack is encountered, Optional will ignore
// any parsing failures and errors.
func Optional[Output any](parse Parser[Output]) Parser[Output] {
	return func(input InputBytes) Result[Output] {
		result := parse(input)
		if result.Err != nil {
			if result.Err.IsFatal() || result.Err.NoWayBack {
				return Failure[Output](result.Err, input)
			}
			result.Err = nil // ignore normal errors
		}

		return Success(result.Output, result.Remaining)
	}
}

// Peek tries to apply the provided parser without consuming any input.
// It effectively allows to look ahead in the input.
func Peek[Output any](parse Parser[Output]) Parser[Output] {
	return func(input InputBytes) Result[Output] {
		oldPos := input.Pos
		result := parse(input)
		if result.Err != nil {
			return Failure[Output](result.Err, input)
		}

		input.Pos = oldPos // don't consume any input
		return Success(result.Output, input)
	}
}

// Recognize returns the consumed input (instead of the original parsers output)
// as the produced value when the provided parser succeeds.
func Recognize[Output any](parse Parser[Output]) Parser[[]byte] {
	return func(input InputBytes) Result[[]byte] {
		result := parse(input)
		if result.Err != nil {
			return Failure[[]byte](result.Err, input)
		}

		return Success(input.BytesTo(result.Remaining), result.Remaining)
	}
}

// Assign returns the provided value if the parser succeeds, otherwise
// it returns an error result.
func Assign[Output1, Output2 any](value Output1, parse Parser[Output2]) Parser[Output1] {
	return func(input InputBytes) Result[Output1] {
		result := parse(input)
		if result.Err != nil {
			return Failure[Output1](result.Err, input)
		}

		return Success(value, result.Remaining)
	}
}
