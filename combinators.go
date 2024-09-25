// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

// NoWayBack applies a child parser and marks a successful result with NoWayBack.
func NoWayBack[Output any](parse Parser[Output]) Parser[Output] {
	return func(input Input) Result[Output] {
		result := parse(input)
		if result.Err != nil {
			result.NoWayBack = true
		}

		return result
	}
}

// Optional applies an optional child parser. Will return nil
// if not successful.
//
// N.B: unless a FatalError or NoWayBack is encountered, Optional will ignore
// any parsing failures and errors.
func Optional[Output any](parse Parser[Output]) Parser[Output] {
	return func(input Input) Result[Output] {
		result := parse(input)
		if result.Err != nil {
			if result.Err.IsFatal() || result.NoWayBack {
				return result
			}
			result.Err = nil // ignore normal errors
		}

		return Success(result.Output, result.Remaining)
	}
}

// Peek tries to apply the provided parser without consuming any input.
// It effectively allows to look ahead in the input.
func Peek[Output any](parse Parser[Output]) Parser[Output] {
	return func(input Input) Result[Output] {
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
	return func(input Input) Result[[]byte] {
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
	return func(input Input) Result[Output1] {
		result := parse(input)
		if result.Err != nil {
			return Failure[Output1](result.Err, input)
		}

		return Success(value, result.Remaining)
	}
}

// Map1 applies a function to the successful result of 1 parser.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map1[PO1 any, MO any](parse Parser[PO1], fn func(PO1) (MO, error)) Parser[MO] {
	return func(input Input) Result[MO] {
		res := parse(input)
		if res.Err != nil {
			return Failure[MO](res.Err, input)
		}

		output, err := fn(res.Output)
		if err != nil {
			return Failure[MO](NewError(input, err.Error()), input)
		}

		return Success(output, res.Remaining)
	}
}

// Map2 applies a function to the successful result of 2 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map2[PO1, PO2 any, MO any](parse1 Parser[PO1], parse2 Parser[PO2], fn func(PO1, PO2) (MO, error)) Parser[MO] {
	return func(input Input) Result[MO] {
		res1 := parse1(input)
		if res1.Err != nil {
			return Failure[MO](NewError(input, "Map2"), input)
		}

		res2 := parse2(res1.Remaining)
		if res2.Err != nil {
			return Failure[MO](NewError(input, "Map2"), input)
		}

		output, err := fn(res1.Output, res2.Output)
		if err != nil {
			return Failure[MO](NewError(input, err.Error()), input)
		}

		return Success(output, res2.Remaining)
	}
}

// Map3 applies a function to the successful result of 3 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map3[PO1, PO2, PO3 any, MO any](parse1 Parser[PO1], parse2 Parser[PO2], parse3 Parser[PO3],
	fn func(PO1, PO2, PO3) (MO, error),
) Parser[MO] {
	return func(input Input) Result[MO] {
		res1 := parse1(input)
		if res1.Err != nil {
			return Failure[MO](NewError(input, "Map3"), input)
		}

		res2 := parse2(res1.Remaining)
		if res2.Err != nil {
			return Failure[MO](NewError(input, "Map3"), input)
		}

		res3 := parse3(res2.Remaining)
		if res3.Err != nil {
			return Failure[MO](NewError(input, "Map3"), input)
		}

		output, err := fn(res1.Output, res2.Output, res3.Output)
		if err != nil {
			return Failure[MO](NewError(input, err.Error()), input)
		}

		return Success(output, res3.Remaining)
	}
}

// Map4 applies a function to the successful result of 4 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map4[PO1, PO2, PO3, PO4 any, MO any](parse1 Parser[PO1], parse2 Parser[PO2], parse3 Parser[PO3], parse4 Parser[PO4],
	fn func(PO1, PO2, PO3, PO4) (MO, error),
) Parser[MO] {
	return func(input Input) Result[MO] {
		res1 := parse1(input)
		if res1.Err != nil {
			return Failure[MO](NewError(input, "Map4"), input)
		}

		res2 := parse2(res1.Remaining)
		if res2.Err != nil {
			return Failure[MO](NewError(input, "Map4"), input)
		}

		res3 := parse3(res2.Remaining)
		if res3.Err != nil {
			return Failure[MO](NewError(input, "Map4"), input)
		}

		res4 := parse4(res3.Remaining)
		if res4.Err != nil {
			return Failure[MO](NewError(input, "Map4"), input)
		}

		output, err := fn(res1.Output, res2.Output, res3.Output, res4.Output)
		if err != nil {
			return Failure[MO](NewError(input, err.Error()), input)
		}

		return Success(output, res4.Remaining)
	}
}

// Map5 applies a function to the successful result of 5 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map5[PO1, PO2, PO3, PO4, PO5 any, MO any](
	parse1 Parser[PO1], parse2 Parser[PO2], parse3 Parser[PO3], parse4 Parser[PO4], parse5 Parser[PO5],
	fn func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) Parser[MO] {
	return func(input Input) Result[MO] {
		res1 := parse1(input)
		if res1.Err != nil {
			return Failure[MO](NewError(input, "Map5"), input)
		}

		res2 := parse2(res1.Remaining)
		if res2.Err != nil {
			return Failure[MO](NewError(input, "Map5"), input)
		}

		res3 := parse3(res2.Remaining)
		if res3.Err != nil {
			return Failure[MO](NewError(input, "Map5"), input)
		}

		res4 := parse4(res3.Remaining)
		if res4.Err != nil {
			return Failure[MO](NewError(input, "Map5"), input)
		}

		res5 := parse5(res4.Remaining)
		if res5.Err != nil {
			return Failure[MO](NewError(input, "Map5"), input)
		}

		output, err := fn(res1.Output, res2.Output, res3.Output, res4.Output, res5.Output)
		if err != nil {
			return Failure[MO](NewError(input, err.Error()), input)
		}

		return Success(output, res5.Remaining)
	}
}
