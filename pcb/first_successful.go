package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
)

// FirstSuccessful tests a list of parsers in order, one by one,
// until one succeeds.
// All parsers have to be of the same type.
//
// If no parser succeeds, this combinator produces an error Result.
func FirstSuccessful[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[Output] {
	if len(parsers) == 0 {
		panic("FirstSuccessful(missing parsers)")
	}

	id := gomme.NewBranchParserID()

	//
	// Are we a real NoWayBack parser? Yes? No? Maybe?
	//
	noWayBacks := 0
	maxNoWayBacks := len(parsers)
	for _, parser := range parsers {
		switch parser.ContainsNoWayBack() {
		case gomme.TernaryYes:
			noWayBacks++
		case gomme.TernaryMaybe:
			noWayBacks++
			maxNoWayBacks++
			break // it will be a Maybe anyway
		default:
			// intentionally left blank
		}
	}
	containingNoWayBack := gomme.TernaryNo
	if noWayBacks >= maxNoWayBacks {
		containingNoWayBack = gomme.TernaryYes
	} else if noWayBacks > 0 {
		containingNoWayBack = gomme.TernaryMaybe
	}

	//
	// Construct myNoWayBackRecoverer from the sub-parsers
	//
	subRecoverers := make([]gomme.Recoverer, len(parsers))
	for i, parser := range parsers {
		if parser.ContainsNoWayBack() != gomme.TernaryNo {
			subRecoverers[i] = parser.NoWayBackRecoverer
		}
	}
	myNoWayBackRecoverer := gomme.NewCombiningRecoverer(true, subRecoverers...)

	//
	// Finally the parsing function
	//
	newParse := func(state gomme.State) (gomme.State, Output) {
		switch state.ParsingMode() {
		case gomme.ParsingModeHappy: // normal parsing (forward)
			return firstSuccessfulHappy(id, parsers, state)
		case gomme.ParsingModeError: // find previous NoWayBack (backward)
			return firstSuccessfulError(id, parsers, state)
		case gomme.ParsingModeHandle: // find error again (forward)
			return firstSuccessfulHandle(id, parsers, state)
		case gomme.ParsingModeRewind: // go back to the witness parser (1)
			return firstSuccessfulRewind(id, parsers, state)
		case gomme.ParsingModeEscape: // find the NoWayBack recoverer with the least waste
			return firstSuccessfulEscape(parsers, state, containingNoWayBack, myNoWayBackRecoverer)
		}

		return state.NewSemanticError(fmt.Sprintf(
			"parsing mode `%s` hasn't been handled in `FirstSuccessful`", state.ParsingMode(),
		)), gomme.ZeroOf[Output]()
	}

	return gomme.NewParser[Output](
		"FirstSuccessful",
		newParse,
		true,
		gomme.DefaultRecovererFunc(newParse), // you really shouldn't use this parser as a Recoverer
		containingNoWayBack,
		myNoWayBackRecoverer.Recover,
	)
}

func firstSuccessfulHappy[Output any](id uint64, parsers []gomme.Parser[Output], state gomme.State,
) (gomme.State, Output) {
	var zero Output

	// use cache to know result immediately
	result, ok := state.CachedParserResult(id)
	if ok {
		if result.Failed {
			return state.ErrorAgain(result.Error), zero
		}
		return state.SucceedAgain(result), result.Output.(Output)
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
			return state.Preserve(newState), zero
		}

		// may the best error win:
		if i == 0 {
			bestState = newState
		} else {
			bestState = gomme.BetterOf(bestState, newState)
			idx = i
		}
	}
	state.CacheParserResult(id, idx, idx, 0, bestState, zero)
	return state.Preserve(bestState), zero
}

func firstSuccessfulError[Output any](id uint64, parsers []gomme.Parser[Output], state gomme.State,
) (gomme.State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, HasNoWayBack)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(error)` parser",
		), zero
	}
	if result.HasNoWayBack {
		parse := parsers[result.Idx]
		newState, _ := parse.It(state)
		if newState.ParsingMode() != gomme.ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
					"parsing mode `handle` in `FirstSuccessful(error)` parser, but mode is: `%s`",
				result.Idx, parse.Expected(), newState.ParsingMode())), zero
		}
		return newState, zero
	}
	return state, zero
}

func firstSuccessfulHandle[Output any](id uint64, parsers []gomme.Parser[Output], state gomme.State,
) (gomme.State, Output) {
	var zero Output
	// TODO: FirstSuccessful might be witness parser (1)
	// use cache to know right parser immediately (Idx, Failed)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(handle)` parser",
		), zero
	}
	if result.Failed {
		parse := parsers[result.Idx]
		newState, output := parse.It(state)
		// the parser failed; so it MUST be the one with the error we are looking for
		if newState.ParsingMode() != gomme.ParsingModeHappy && newState.ParsingMode() != gomme.ParsingModeEscape {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
					"parsing mode `happy` or `escape` in `FirstSuccessful(handle)` parser, but mode is: `%s`",
				result.Idx, parse.Expected(), newState.ParsingMode())), zero
		}
		return newState, output
	}
	return state, zero
}

func firstSuccessfulRewind[Output any](id uint64, parsers []gomme.Parser[Output], state gomme.State,
) (gomme.State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, Failed)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(rewind)` parser",
		), zero
	}
	if result.Failed {
		parse := parsers[result.Idx]
		newState, output := parse.It(state)
		// the parser failed; so it MUST be the one with the error we are looking for
		if newState.ParsingMode() != gomme.ParsingModeHappy && newState.ParsingMode() != gomme.ParsingModeEscape {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
					"parsing mode `happy` or `escape` in `FirstSuccessful(rewind)` parser, but mode is: `%s`",
				result.Idx, parse.Expected(), newState.ParsingMode())), zero
		}
		return newState, output
	}
	return state, zero
}

func firstSuccessfulEscape[Output any](
	parsers []gomme.Parser[Output],
	state gomme.State,
	containingNoWayBack gomme.Ternary,
	noWayBackRecoverer gomme.CombiningRecoverer,
) (gomme.State, Output) {
	var zero Output
	if containingNoWayBack == gomme.TernaryNo { // we can't help
		return state, zero
	}

	idx, ok := noWayBackRecoverer.CachedIndex(state)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(escape)` parser",
		), zero
	}

	parse := parsers[idx]
	newState, output := parse.It(state)
	// this parser has the best recoverer; so it MUST make us happy again
	if newState.ParsingMode() != gomme.ParsingModeHappy && newState.ParsingMode() != gomme.ParsingModeEscape {
		return state.NewSemanticError(fmt.Sprintf(
			"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
				"parsing mode `happy` or `escape` in `FirstSuccessful(escape)` parser, but mode is: `%s`",
			idx, parse.Expected(), newState.ParsingMode())), zero
	}
	return newState, output
}
