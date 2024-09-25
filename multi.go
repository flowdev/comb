package gomme

// Count runs the provided parser `count` times.
//
// If the provided parser cannot be successfully applied `count` times, the operation
// fails and the Result will contain an error.
func Count[Output any](parse Parser[Output], count uint) Parser[[]Output] {
	return func(input Input) Result[[]Output] {
		if input.AtEnd() || count == 0 {
			return Failure[[]Output](NewError(input, "Count"), input)
		}

		outputs := make([]Output, 0, int(count))
		remaining := input
		for i := uint(0); i < count; i++ {
			result := parse(remaining)
			if result.Err != nil {
				return Failure[[]Output](result.Err, input)
			}

			remaining = result.Remaining
			outputs = append(outputs, result.Output)
		}

		return Success(outputs, remaining)
	}
}

// Many0 applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output.
//
// Note that Many0 will succeed even if the parser fails to match at all. It will
// however fail if the provided parser accepts empty inputs (such as `Digit0`, or
// `Alpha0`) in order to prevent infinite loops.
func Many0[Output any](parse Parser[Output]) Parser[[]Output] {
	return func(input Input) Result[[]Output] {
		results := []Output{}

		remaining := input
		for {
			res := parse(remaining)
			if res.Err != nil {
				return Success(results, remaining)
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if res.Remaining.Pos == remaining.Pos {
				return Failure[[]Output](NewError(input, "Many0"), input)
			}

			results = append(results, res.Output)
			remaining = res.Remaining
		}
	}
}

// Many1 applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output. Many1 will fail if the parser fails to
// match at least once.
//
// Note that Many1 will fail if the provided parser accepts empty
// inputs (such as `Digit0`, or `Alpha0`) in order to prevent infinite loops.
func Many1[Output any](parse Parser[Output]) Parser[[]Output] {
	return func(input Input) Result[[]Output] {
		first := parse(input)
		if first.Err != nil {
			return Failure[[]Output](first.Err, input)
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if first.Remaining.Pos == input.Pos {
			return Failure[[]Output](NewError(input, "Many1"), input)
		}

		results := []Output{first.Output}
		remaining := first.Remaining

		for {
			res := parse(remaining)
			if res.Err != nil {
				return Success(results, remaining)
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if res.Remaining.Pos == remaining.Pos {
				return Failure[[]Output](NewError(remaining, "Many1"), input)
			}

			results = append(results, res.Output)
			remaining = res.Remaining
		}
	}
}

// SeparatedList0 applies an element parser and a separator parser repeatedly in order
// to produce a list of elements.
//
// Note that SeparatedList0 will succeed even if the element parser fails to match at all.
// It will however fail if the provided element parser accepts empty inputs (such as
// `Digit0`, or `Alpha0`) in order to prevent infinite loops.
//
// Because the `SeparatedList0` is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed even if the separator parser fails to
// match at all. It will however fail if the provided separator parser accepts empty
// inputs in order to prevent infinite loops.
func SeparatedList0[Output any, S Separator](
	parse Parser[Output],
	separator Parser[S],
	parseSeparatorAtEnd bool,
) Parser[[]Output] {
	return func(input Input) Result[[]Output] {
		results := []Output{}

		res := parse(input)
		if res.Err != nil {
			return Success(results, input)
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if res.Remaining.Pos == input.Pos {
			return Failure[[]Output](NewError(input, "SeparatedList0"), input)
		}

		results = append(results, res.Output)
		remaining := res.Remaining

		for {
			separatorResult := separator(remaining)
			if separatorResult.Err != nil {
				return Success(results, remaining)
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if separatorResult.Remaining.Pos == remaining.Pos {
				return Failure[[]Output](NewError(remaining, "SeparatedList0"), input)
			}

			parserResult := parse(separatorResult.Remaining)
			if parserResult.Err != nil {
				if parseSeparatorAtEnd {
					return Success(results, separatorResult.Remaining)
				} else {
					return Success(results, remaining)
				}
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if parserResult.Remaining.Pos == separatorResult.Remaining.Pos {
				return Failure[[]Output](NewError(separatorResult.Remaining, "SeparatedList0"), input)
			}

			results = append(results, parserResult.Output)
			remaining = parserResult.Remaining
		}
	}
}

// SeparatedList1 applies an element parser and a separator parser repeatedly in order
// to produce a list of elements.
//
// Note that SeparatedList1 will fail if the element parser fails to match at all.
//
// Because the `SeparatedList1` is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed even if the separator parser fails to
// match at all.
func SeparatedList1[Output any, S Separator](
	parse Parser[Output],
	separator Parser[S],
	parseSeparatorAtEnd bool,
) Parser[[]Output] {
	return func(input Input) Result[[]Output] {
		res := parse(input)
		if res.Err != nil {
			return Failure[[]Output](res.Err, input)
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if res.Remaining.Pos == input.Pos {
			return Failure[[]Output](NewError(input, "SeparatedList1"), input)
		}

		results := []Output{res.Output}
		remaining := res.Remaining

		for {
			separatorResult := separator(remaining)
			if separatorResult.Err != nil {
				return Success(results, remaining)
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if separatorResult.Remaining.Pos == remaining.Pos {
				return Failure[[]Output](NewError(remaining, "SeparatedList1"), input)
			}

			parserResult := parse(separatorResult.Remaining)
			if parserResult.Err != nil {
				if parseSeparatorAtEnd {
					return Success(results, separatorResult.Remaining)
				} else {
					return Success(results, remaining)
				}
			}

			// Checking for infinite loops, if nothing was consumed,
			// the provided parser would make us go around in circles.
			if parserResult.Remaining.Pos == separatorResult.Remaining.Pos {
				return Failure[[]Output](NewError(separatorResult.Remaining, "SeparatedList1"), input)
			}

			results = append(results, parserResult.Output)
			remaining = parserResult.Remaining
		}
	}
}
