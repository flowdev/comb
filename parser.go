package gomme

import (
	"sync/atomic"
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
		newState, output := parse.It(state)
		if newState.Failed() {
			return newState, output
		}

		newState.noWayBackMark = newState.input.pos
		return newState, output
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
		case ParsingModeHappy:
			bestState := state
			for i, parse := range parsers {
				newState, output := parse.It(state)
				if !newState.Failed() {
					return newState, output
				}

				// may the best error win:
				if i == 0 {
					bestState = newState
				} else {
					bestState = BetterOf(bestState, newState)
				}
			}
			return state.Failure(bestState), ZeroOf[Output]()
		case ParsingModeError:
			// TODO: Help finding last NoWayBack
		case ParsingModeHandle:
		case ParsingModeRecord:
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

var firstSuccessfulIDs = &atomic.Uint64{}
