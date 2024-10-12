package gomme

import (
	"sync/atomic"
)

// NoWayBack applies a sub-parser and marks the new state as a
// point of no return if successful.
// Use this to signal that the right alternative has been found by the
// FirstSuccessful parser even in case of a later error.
// NoWayBack can also be used to minimize the amount of backtracking
// in other places.
//
// If you only want to mark your parser as a place to recover to,
// please use Refuge instead.
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

		return newState.SignalNoWayBack(), output
	}

	return NewParser[Output](
		"NoWayBack",
		newParse,
		parse.MyRecoverer(),
		TernaryYes, // Refuge and NoWayBack are the only ones to be sure
		CachingRecoverer(parse.MyRecoverer()),
	)
}

// Refuge applies its sub-parser and marks it as a possible place to
// recover to with the Recoverer of its sub-parser.
// Please use this a lot. The user experience will be much better.
// I am even thinking of using this automatically for every parser that
// consumes at least one token (according to the Deleter).
//
// Parsers that accept the empty input or only perform look ahead are
// not allowed as sub-parsers.
// It tests the optional recoverer of the parser during the construction phase
// to provoke an early panic.
// This way we won't have a panic at the runtime of the parser.
func Refuge[Output any](parse Parser[Output]) Parser[Output] {
	// call Recoverer to make a Forbidden recoverer panic during the construction phase
	recoverer := parse.MyRecoverer()
	if recoverer != nil {
		recoverer(NewState(0, DefaultBinaryDeleter, []byte{}))
	}

	newParse := func(state State) (State, Output) {
		return parse.It(state)
	}

	return NewParser[Output](
		"NoWayBack",
		newParse,
		parse.MyRecoverer(),
		TernaryYes, // Refuge and NoWayBack are the only ones to be sure
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
		switch parser.ContainsRefuge() {
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
	// Construct myRefugeRecoverer from the sub-parsers
	//
	subRecoverers := make([]Recoverer, len(parsers))
	for i, parser := range parsers {
		if parser.ContainsRefuge() > TernaryNo {
			subRecoverers[i] = parser.RefugeRecoverer
		}
	}
	myRefugeRecoverer := NewCombiningRecoverer(subRecoverers...)

	//
	// Finally the parsing function
	//
	newParse := func(state State) (State, Output) {
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
	}

	return NewParser[Output](
		"FirstSuccessful",
		newParse,
		DefaultRecovererFunc(newParse), // you really shouldn't use this parser as a Recoverer
		containingNoWayBack,
		myRefugeRecoverer.Recover,
	)
}

var firstSuccessfulIDs = &atomic.Uint64{}
