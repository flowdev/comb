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
//  2. Prevent unnecessary backtracking in case of an error.
//  3. Mark a parser as a potential safe place to recover to
//     when recovering from an error.
//
// So you don't need this parser at all if your input is always correct.
// NoWayBack is the cornerstone of good and performant parsing otherwise.
//
// Parsers that accept the empty input or only perform look ahead are
// not allowed as sub-parsers.
// It tests the optional recoverer of the parser during the construction phase
// to provoke an early panic.
// This way we won't have a panic at the runtime of the parser.
func NoWayBack[Output any](parse Parser[Output]) Parser[Output] {
	// call Recoverer to make a Forbidden recoverer panic during the construction phase
	recoverer := parse.MyRecoverer()
	if recoverer != nil {
		recoverer(NewState(0, DefaultBinaryDeleter, []byte{}))
	}

	newParse := func(state State) (State, Output) {
		switch state.mode {
		case ParsingModeHappy:
			newState, output := parse.It(state)
			if !newState.Failed() {
				newState.noWayBackMark = newState.input.pos
			}
			return newState, output
		case ParsingModeError: // we found the previous NoWayBack => switch to handle and find error again
			state.mode = ParsingModeHandle
			return state, ZeroOf[Output]()
		case ParsingModeHandle: // the sub-parser must have failed, or we have a programming error
			newState, output := parse.It(state)
			if newState.mode != ParsingModeRecord {
				return newState.NewSemanticError(
					"programming error: NoWayBack called in mode `handle`; " +
						"we must have missed the error to be handled",
				), output
			}
			// TODO: record sub-parser???
			return newState, output
		case ParsingModeRecord:
			// we found the next NoWayBack => stop recording; play recorded parsers and switch to happy
			// TODO: implement!
		case ParsingModeCollect, ParsingModeChoose, ParsingModePlay:
			// TODO: Think about it???
		}
		return state.NewSemanticError(fmt.Sprintf("parsing mode %v hasn't been handled in NoWayBack", state.mode)), ZeroOf[Output]()
	}

	return NewParser[Output](
		"NoWayBack",
		newParse,
		parse.MyRecoverer(),
		TernaryYes, // NoWayBack is the only one to be sure
		CachingRecoverer(parse.MyRecoverer()),
	)
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
			return parseFirstSuccessfulHappy(state, parsers, id)
		case ParsingModeError: // find previous NoWayBack (backward)
		case ParsingModeHandle: // find error again (forward)
			// use cache to know right parser immediately (Idx, Failed)
			result, ok := state.CachedParserResult(id)
			if !ok {
				return state.NewSemanticError(
					"grammar error: cache was empty in `FirstSuccessful` parser for parsing mode `handle`",
				), ZeroOf[Output]()
			}
			if result.Failed {
				parser := parsers[result.Idx]
				newState, output := parser.It(state)
				// the parser failed; so it MUST be the one with the error we are looking for
				if newState.mode != ParsingModeRecord && newState.mode != ParsingModeHappy {
					return state.NewSemanticError(fmt.Sprintf(
						"programming errror: sub-parser (index: %d, expected: %q) "+
							"didn't switch to parsing mode `record` or `happy`",
						result.Idx, parser.Expected())), ZeroOf[Output]()
				}
				return newState, output
			}
			return state, ZeroOf[Output]()
		case ParsingModeRecord: // find next NoWayBack, recording on the way (forward)
		case ParsingModeCollect:
		case ParsingModeChoose:
		case ParsingModePlay:
		}

		return state, ZeroOf[Output]()
	}

	return NewParser[Output](
		"FirstSuccessful",
		newParse,
		DefaultRecovererFunc(newParse), // you really shouldn't use this parser as a Recoverer
		containingNoWayBack,
		myNoWayBackRecoverer.Recover,
	)
}
func parseFirstSuccessfulHappy[Output any](state State, parsers []Parser[Output], id uint64) (State, Output) {
	var zero Output

	// use cache to know result immediately (Idx, Output)
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
func parseFirstSuccessfulError[Output any](state State, parsers []Parser[Output], id uint64) (State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, HasNoWayBack)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful` parser for parsing mode `error`",
		), zero
	}
	if result.HasNoWayBack {
		newState, _ := parsers[result.Idx].It(state)
		if newState.mode != ParsingModeRecord {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) "+
					"didn't switch to parsing mode `record`",
				result.Idx, parsers[result.Idx].Expected())), zero
		}
	}
	return state, zero
}
func parseFirstSuccessfulHandle[Output any](state State, parsers []Parser[Output], id uint64) (State, Output) {
	var zero Output
	return state, zero
}
