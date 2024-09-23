package gomme

// Alternative tests a list of parsers in order, one by one, until one
// succeeds.
//
// If none of the parsers succeed, this combinator produces an error Result.
func Alternative[Input Bytes, Output any](parsers ...Parser[Input, Output]) Parser[Input, Output] {
	return func(input InputBytes[Input]) Result[Output, Input] {
		err := Error[Input]{Input: input}
		err.Input.Pos = -1 // no error position at all so far
		for _, parse := range parsers {
			result := parse(input)
			if result.Err == nil {
				return result
			}

			// may the best error(s) win:
			if result.Err.Input.Pos > err.Input.Pos {
				err = *result.Err
			} else if result.Err.Input.Pos == err.Input.Pos {
				err.Expected = append(err.Expected, result.Err.Expected...)
			}
		}

		return Failure[Input, Output](&err, input)
	}
}
