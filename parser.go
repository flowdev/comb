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
		recoverer(NewState(0, DefaultBinaryDeleter, []byte{}))
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
		parse.MyRecoverer(),
		TernaryYes, // NoWayBack is the only one to be sure
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
		return newState, output
	}
	newState.errHand.witnessID = 0 // ensure we are the witness!
	return IWitnessed(state, id, 0, newState), output
}
func noWayBackError[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	state.mode = ParsingModeHandle
	return state, ZeroOf[Output]()
}
func noWayBackHandle[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	return HandleWitness(state, id, parse)
}
func noWayBackRewind[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	return HandleWitness(state, id, parse)
}
func noWayBackEscape[Output any](id uint64, parse Parser[Output], state State) (State, Output) {
	if state.input.pos <= state.errHand.err.pos {
		return state, ZeroOf[Output]() // we are too far in front in the input
	}
	newState := state.MoveBy(parse.MyRecoverer()(state))
	newState.errHand = errHand{}
	newState.mode = ParsingModeHappy

	return noWayBackHappy(id, parse, newState)
}

// FirstSuccessful tests a list of parsers in order, one by one,
// until one succeeds.
// All parsers have to be of the same type.
//
// If no parser succeeds, this combinator produces an error Result.
func FirstSuccessful[Output any](parsers ...Parser[Output]) Parser[Output] {
	if len(parsers) == 0 {
		panic("FirstSuccessful(missing parsers)")
	}

	id := NewBranchParserID()

	//
	// Are we a real NoWayBack parser? Yes? No? Maybe?
	//
	noWayBacks := 0
	maxNoWayBacks := len(parsers)
	for _, parser := range parsers {
		switch parser.ContainsNoWayBack() {
		case TernaryYes:
			noWayBacks++
		case TernaryMaybe:
			noWayBacks++
			maxNoWayBacks++
			break // it will be a Maybe anyway
		default:
			// intentionally left blank
		}
	}
	containingNoWayBack := TernaryNo
	if noWayBacks >= maxNoWayBacks {
		containingNoWayBack = TernaryYes
	} else if noWayBacks > 0 {
		containingNoWayBack = TernaryMaybe
	}

	//
	// Construct myNoWayBackRecoverer from the sub-parsers
	//
	subRecoverers := make([]Recoverer, len(parsers))
	for i, parser := range parsers {
		if parser.ContainsNoWayBack() != TernaryNo {
			subRecoverers[i] = parser.NoWayBackRecoverer
		}
	}
	myNoWayBackRecoverer := NewCombiningRecoverer(subRecoverers...)

	//
	// Finally the parsing function
	//
	newParse := func(state State) (State, Output) {
		switch state.mode {
		case ParsingModeHappy: // normal parsing (forward)
			return firstSuccessfulHappy(id, parsers, state)
		case ParsingModeError: // find previous NoWayBack (backward)
			return firstSuccessfulError(id, parsers, state)
		case ParsingModeHandle: // find error again (forward)
			return firstSuccessfulHandle(id, parsers, state)
		case ParsingModeRewind: // go back to the witness parser (1)
			return firstSuccessfulRewind(id, parsers, state)
		case ParsingModeEscape: // find the NoWayBack recoverer with the least waste
			return firstSuccessfulEscape(parsers, state, containingNoWayBack, myNoWayBackRecoverer)
		}

		return state.NewSemanticError(fmt.Sprintf(
			"parsing mode %v hasn't been handled in `FirstSuccessful`", state.mode)), ZeroOf[Output]()
	}

	return NewParser[Output](
		"FirstSuccessful",
		newParse,
		DefaultRecovererFunc(newParse), // you really shouldn't use this parser as a Recoverer
		containingNoWayBack,
		myNoWayBackRecoverer.Recover,
	)
}
func firstSuccessfulHappy[Output any](id uint64, parsers []Parser[Output], state State) (State, Output) {
	var zero Output

	// use cache to know result immediately
	result, ok := state.CachedParserResult(id)
	if ok {
		if result.Failed {
			return state.ErrorAgain(result.Error), zero
		}
		return state.MoveBy(result.Consumed), result.Output.(Output)
	}
	// cache miss: parse
	bestState := state
	idx := 0
	for i, parse := range parsers {
		newState, output := parse.It(state)
		if !newState.Failed() {
			state.CacheParserResult(id, i, i, 0, newState, output)
			return newState, output
		}

		if state.NoWayBackMoved(newState) { // don't look further than this
			state.CacheParserResult(id, i, i, 0, newState, output)
			return state.Failure(newState), zero
		}

		// may the best error win:
		if i == 0 {
			bestState = newState
		} else {
			bestState = BetterOf(bestState, newState)
			idx = i
		}
	}
	state.CacheParserResult(id, idx, idx, 0, bestState, zero)
	return state.Failure(bestState), zero
}
func firstSuccessfulError[Output any](id uint64, parsers []Parser[Output], state State) (State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, HasNoWayBack)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(error)` parser",
		), zero
	}
	if result.HasNoWayBack {
		newState, _ := parsers[result.Idx].It(state)
		if newState.mode != ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) "+
					"didn't switch to parsing mode `handle` in `FirstSuccessful(error)` parser",
				result.Idx, parsers[result.Idx].Expected())), zero
		}
		return newState, zero
	}
	return state, zero
}
func firstSuccessfulHandle[Output any](id uint64, parsers []Parser[Output], state State) (State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, Failed)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(handle)` parser",
		), zero
	}
	if result.Failed {
		parser := parsers[result.Idx]
		newState, output := parser.It(state)
		// the parser failed; so it MUST be the one with the error we are looking for
		if newState.mode != ParsingModeHappy {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) "+
					"didn't switch to parsing mode `happy` in `FirstSuccessful(handle)` parser",
				result.Idx, parser.Expected())), zero
		}
		return newState, output
	}
	return state, zero
}
func firstSuccessfulRewind[Output any](id uint64, parsers []Parser[Output], state State) (State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, Failed)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(rewind)` parser",
		), zero
	}
	if result.Failed {
		parser := parsers[result.Idx]
		newState, output := parser.It(state)
		// the parser failed; so it MUST be the one with the error we are looking for
		if newState.mode != ParsingModeHappy && newState.mode != ParsingModeEscape {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) "+
					"didn't switch to parsing mode `happy` or `escape` in `FirstSuccessful(rewind)` parser",
				result.Idx, parser.Expected())), zero
		}
		return newState, output
	}
	return state, zero
}
func firstSuccessfulEscape[Output any](
	parsers []Parser[Output],
	state State,
	containingNoWayBack Ternary,
	noWayBackRecoverer CombiningRecoverer,
) (State, Output) {
	var zero Output
	if containingNoWayBack == TernaryNo { // we can't help
		return state, zero
	}

	idx, ok := noWayBackRecoverer.CachedIndex(state)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(escape)` parser",
		), zero
	}

	parser := parsers[idx]
	newState, output := parser.It(state)
	// this parser has the best recoverer; so it MUST make us happy again
	if newState.mode != ParsingModeHappy {
		return state.NewSemanticError(fmt.Sprintf(
			"programming error: sub-parser (index: %d, expected: %q) "+
				"didn't switch to parsing mode `happy` in `FirstSuccessful(handle)` parser",
			idx, parser.Expected())), zero
	}
	return newState, output
}
