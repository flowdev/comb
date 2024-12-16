package gomme

import (
	"fmt"
)

// NoWayBack applies a sub-parser and marks the new state as a
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
// NoWayBack is the cornerstone of good and performant parsing otherwise.
//
// Note:
//   - Parsers that accept the empty input or only perform look ahead are
//     NOT allowed as sub-parsers.
//     NoWayBack tests the optional recoverer of the parser during the
//     construction phase to provoke an early panic.
//     This way we won't have a panic at the runtime of the parser.
//   - Only leaf parsers MUST be given to NoWayBack as sub-parsers.
//     NoWayBack will treat the sub-parser as a leaf parser.
//     So it won't bother it with any error handling including witnessing errors.
func NoWayBack[Output any](parse Parser[Output]) Parser[Output] {
	id := NewBranchParserID()

	// call Recoverer to make a Forbidden recoverer panic during the construction phase
	recoverer := parse.MyRecoverer()
	if recoverer != nil {
		recoverer(NewFromBytes(-1, DefaultBinaryDeleter, []byte{}))
	}

	newParse := func(state State) (State, Output) {
		switch state.mode {
		case ParsingModeHappy:
			return noWayBackHappy(id, parse, state)
		case ParsingModeError: // we found the previous NoWayBack => switch to handle and find error again
			return noWayBackError(id, parse, state)
		case ParsingModeHandle: // the sub-parser must have failed, or we have a programming error
			return noWayBackHandle(id, parse, state)
		case ParsingModeRewind: // error didn't go away yet; go back to witness parser (1)
			return noWayBackRewind(id, parse, state)
		case ParsingModeEscape: // recover from the error the hard way; use the recoverer
			return noWayBackEscape(id, parse, state)
		}
		return state.NewSemanticError(fmt.Sprintf(
			"parsing mode %v hasn't been handled in `NoWayBack`", state.mode)), ZeroOf[Output]()
	}

	return NewParser[Output](
		"NoWayBack",
		newParse,
		true,
		parse.MyRecoverer(),
		CachingRecoverer(parse.MyRecoverer()),
	)
}
func noWayBackHappy[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	newState, output := parse.It(state)
	if !newState.Failed() {
		if newState.errHand.witnessID > 0 { // we just successfully handled an error :)
			newState.errHand = errHand{}
		}
		newState.noWayBackMark = newState.input.pos // move the mark!
		return newState.ClearAllCaches(), output
	}
	newState.errHand.witnessID = 0 // ensure we are the witness!
	return IWitnessed(state, id, 0, newState), output
}
func noWayBackError[Output any](_ uint64, _ Parser[Output], state State) (State, Output) {
	state.mode = ParsingModeHandle
	state.oldErrors = append(state.oldErrors, *state.errHand.err)
	state.errHand.err = nil
	return state, ZeroOf[Output]()
}
func noWayBackHandle[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	return HandleWitness(state, id, 0, parse)
}
func noWayBackRewind[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	return HandleWitness(state, id, 0, parse)
}
func noWayBackEscape[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	waste := parse.MyRecoverer()(state)
	if waste < 0 {
		return state.MoveBy(state.BytesRemaining()), ZeroOf[Output]() // give up
	}
	newState := state.MoveBy(waste)
	newState.errHand = errHand{}
	newState.mode = ParsingModeHappy

	return noWayBackHappy(id, parse, newState)
}
