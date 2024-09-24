package gomme

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any](prefix Parser[OP], parser Parser[O], suffix Parser[OS]) Parser[O] {
	return func(input InputBytes) Result[O] {
		return Terminated(Preceded(prefix, parser), suffix)(input)
	}
}

// Pair applies two parsers and returns a Result containing a pair container holding
// the resulting values.
func Pair[LO, RO any, LP Parser[LO], RP Parser[RO]](
	leftParser LP, rightParser RP,
) Parser[PairContainer[LO, RO]] {
	return func(input InputBytes) Result[PairContainer[LO, RO]] {
		leftResult := leftParser(input)
		if leftResult.Err != nil {
			return Failure[PairContainer[LO, RO]](NewError(input, "Pair"), input)
		}

		rightResult := rightParser(leftResult.Remaining)
		if rightResult.Err != nil {
			return Failure[PairContainer[LO, RO]](NewError(leftResult.Remaining, "Pair"), input)
		}

		return Success(PairContainer[LO, RO]{leftResult.Output, rightResult.Output}, rightResult.Remaining)
	}
}

// Preceded parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
//
// Preceded is effectively equivalent to applying DiscardAll(prefix),
// and then applying the main parser.
func Preceded[OP, O any](prefix Parser[OP], parser Parser[O]) Parser[O] {
	return func(input InputBytes) Result[O] {
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

// SeparatedPair applies two separated parsers and returns a PairContainer as Result.
// The result of the separator parser is discarded.
func SeparatedPair[LO, RO any, S Separator, LP Parser[LO], SP Parser[S], RP Parser[RO]](
	leftParser LP, separator SP, rightParser RP,
) Parser[PairContainer[LO, RO]] {
	return func(input InputBytes) Result[PairContainer[LO, RO]] {
		leftResult := leftParser(input)
		if leftResult.Err != nil {
			return Failure[PairContainer[LO, RO]](NewError(input, "SeparatedPair"), input)
		}

		sepResult := separator(leftResult.Remaining)
		if sepResult.Err != nil {
			return Failure[PairContainer[LO, RO]](NewError(leftResult.Remaining, "SeparatedPair"), input)
		}

		rightResult := rightParser(sepResult.Remaining)
		if rightResult.Err != nil {
			return Failure[PairContainer[LO, RO]](NewError(sepResult.Remaining, "SeparatedPair"), input)
		}

		return Success(PairContainer[LO, RO]{leftResult.Output, rightResult.Output}, rightResult.Remaining)
	}
}

// Sequence applies a sequence of parsers and returns either a
// slice of results or an error if any parser fails.
// All parsers in the sequence have to produce the same result type.
func Sequence[O any](parsers ...Parser[O]) Parser[[]O] {
	return func(input InputBytes) Result[[]O] {
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
	return func(input InputBytes) Result[O] {
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
