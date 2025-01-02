package gomme

// SaveSpot applies a sub-parser and marks the new state as a
// point of no return if successful.
// It really serves 3 slightly different purposes:
//
//  1. Prevent a `FirstSuccessful` parser from trying later sub-parsers even
//     in case of an error.
//  2. Prevent other unnecessary backtracking in case of an error.
//  3. Mark a parser as a potential safe place to recover to
//     when recovering from an error.
//
// So you don't need this parser at all if your input is always correct.
// SaveSpot is the cornerstone of good and performant parsing otherwise.
//
// Note:
//   - Parsers that accept the empty input or only perform look ahead are
//     NOT allowed as sub-parsers.
//     SaveSpot tests the optional recoverer of the parser during the
//     construction phase to provoke an early panic.
//     This way we won't have a panic at the runtime of the parser.
//   - Only leaf parsers MUST be given to SaveSpot as sub-parsers.
//     SaveSpot will treat the sub-parser as a leaf parser.
//     So the sub-parser will never get a chance to parse with given internal data.
func SaveSpot[Output any](parse Parser[Output]) Parser[Output] {
	// call Recoverer to make a Forbidden recoverer panic during the construction phase
	recoverer := parse.Recover
	if recoverer != nil {
		recoverer(NewFromBytes(-1, DefaultBinaryDeleter, -1, 1, []byte{}))
	}

	//newParse := func(state State) (State, Output, *ParserError) {
	//	switch state.mode {
	//	case ParsingModeHappy:
	//		return saveSpotHappy(id, parse, state)
	//	case ParsingModeError: // we found the previous SaveSpot => switch to handle and find error again
	//		return saveSpotError(id, parse, state)
	//	case ParsingModeHandle: // the sub-parser must have failed, or we have a programming error
	//		return saveSpotHandle(id, parse, state)
	//	case ParsingModeRewind: // error didn't go away yet; go back to witness parser (1)
	//		return saveSpotRewind(id, parse, state)
	//	case ParsingModeEscape: // recover from the error the hard way; use the recoverer
	//		return saveSpotEscape(id, parse, state)
	//	}
	//	return state.NewSemanticError(fmt.Sprintf(
	//		"parsing mode %v hasn't been handled in `SaveSpot`", state.mode)), ZeroOf[Output]()
	//}

	sp := NewParser[Output]("SaveSpot", parse.It, recoverer)
	sp.setSaveSpot()
	return sp
}
func saveSpotHappy[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	newState, output, err := parse.It(state)
	if err == nil {
		if newState.errHand.witnessID > 0 { // we just successfully handled an error :)
			newState.errHand = errHand{}
		}
		newState.saveSpot = newState.input.pos // move the mark!
		return newState.ClearAllCaches(), output
	}
	newState.errHand.witnessID = 0 // ensure we are the witness!
	return IWitnessed(state, id, 0, newState), output
}
func saveSpotError[Output any](_ uint64, _ Parser[Output], state State) (State, Output) {
	state.mode = ParsingModeHandle
	state.oldErrors = append(state.oldErrors, *state.errHand.err)
	state.errHand.err = nil
	return state, ZeroOf[Output]()
}
func saveSpotHandle[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	return HandleWitness(state, id, 0, parse)
}
func saveSpotRewind[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	return HandleWitness(state, id, 0, parse)
}
func saveSpotEscape[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	waste := parse.Recover(state)
	if waste < 0 {
		return state.MoveBy(state.BytesRemaining()), ZeroOf[Output]() // give up
	}
	newState := state.MoveBy(waste)
	newState.errHand = errHand{}
	newState.mode = ParsingModeHappy

	return saveSpotHappy(id, parse, newState)
}
