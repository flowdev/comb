package gomme

// Alternative tests a list of parsers in order, one by one, until one
// succeeds.
//
// If none of the parsers succeed, this combinator produces an error Result.
func Alternative[Output any](parsers ...Parser[Output]) Parser[Output] {
	return func(input InputBytes) Result[Output] {
		if len(parsers) == 0 {
			return Failure[Output](NewError(input, "Alternative(no parser given)"), input)
		}

		err := Error{Input: input}
		for i, parse := range parsers {
			result := parse(input)
			if result.Err == nil {
				return result
			}

			// may the best error(s) win:
			if result.Err.Input.Pos > err.Input.Pos || i == 0 {
				err = *result.Err
			} else if result.Err.Input.Pos == err.Input.Pos {
				err.Expected = append(err.Expected, result.Err.Expected...)
			}
		}

		return Failure[Output](&err, input)
	}
}
