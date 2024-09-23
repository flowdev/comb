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

// Separator is a generic type alias for separator characters
type Separator interface {
	rune | byte | string
}

// Result is a generic type alias for Result
type Result[Output any] struct {
	Output    Output
	Err       *Error
	Remaining InputBytes
}

// Parser defines the type of a generic Parser function
type Parser[Output any] func(input InputBytes) Result[Output]

// Success creates a Result with an output set from
// the result of a successful parsing.
func Success[Output any](output Output, r InputBytes) Result[Output] {
	return Result[Output]{output, nil, r}
}

// Failure creates a Result with an error set from
// the result of a failed parsing.
func Failure[Output any](err *Error, input InputBytes) Result[Output] {
	var output Output
	return Result[Output]{output, err, input}
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
		pos0 := input.Pos
		result := parse(input)
		if result.Err != nil {
			return Failure[[]byte](result.Err, input)
		}

		return Success(input.Bytes[pos0:input.Pos], result.Remaining)
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
