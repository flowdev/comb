package gomme

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any](prefix Parser[OP], parser Parser[O], suffix Parser[OS]) Parser[O] {
	return func(input Input) Result[O] {
		return Terminated(Preceded(prefix, parser), suffix)(input)
	}
}

// Preceded parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
//
// Preceded is effectively equivalent to applying DiscardAll(prefix),
// and then applying the main parser.
func Preceded[OP, O any](prefix Parser[OP], parser Parser[O]) Parser[O] {
	return func(input Input) Result[O] {
		prefixResult := prefix(input)
		if prefixResult.Err != nil {
			return Failure[O](prefixResult.Err, input)
		}

		result := parser(prefixResult.Remaining)
		if result.Err != nil {
			return Failure[O](result.Err, input)
		}

		return result
	}
}

// Sequence applies a sequence of parsers and returns either a
// slice of results or an error if any parser fails.
// All parsers in the sequence have to produce the same result type.
func Sequence[O any](parsers ...Parser[O]) Parser[[]O] {
	return func(input Input) Result[[]O] {
		remaining := input
		outputs := make([]O, 0, len(parsers))

		for _, parser := range parsers {
			res := parser(remaining)
			if res.Err != nil {
				return Failure[[]O](res.Err, input)
			}

			outputs = append(outputs, res.Output)
			remaining = res.Remaining
		}

		return Success(outputs, remaining)
	}
}

// Terminated parses a result from the main parser, it then
// parses the result from the suffix parser and discards it; only
// returning the result of the main parser.
func Terminated[O, OS any](parser Parser[O], suffix Parser[OS]) Parser[O] {
	return func(input Input) Result[O] {
		result := parser(input)
		if result.Err != nil {
			return Failure[O](result.Err, input)
		}

		suffixResult := suffix(result.Remaining)
		if suffixResult.Err != nil {
			return Failure[O](suffixResult.Err, input)
		}

		return Success(result.Output, suffixResult.Remaining)
	}
}
