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

	fsd := &firstSuccessfulData[Output]{
		id:                  gomme.NewBranchParserID(),
		parsers:             parsers,
		containingNoWayBack: containingNoWayBack,
		noWayBackRecoverer:  myNoWayBackRecoverer,
	}

	//
	// Finally the parsing function
	//
	newParse := func(state gomme.State) (gomme.State, Output) {
		switch state.ParsingMode() {
		case gomme.ParsingModeHappy: // normal parsing (forward)
			return fsd.happy(state)
		case gomme.ParsingModeError: // find previous NoWayBack (backward)
			return fsd.error(state)
		case gomme.ParsingModeHandle: // find error again (forward)
			return fsd.handle(state)
		case gomme.ParsingModeRewind: // go back to the witness parser (1)
			return fsd.rewind(state)
		case gomme.ParsingModeEscape: // find the NoWayBack recoverer with the least waste
			return fsd.escape(state)
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

type firstSuccessfulData[Output any] struct {
	id                  uint64
	parsers             []gomme.Parser[Output]
	containingNoWayBack gomme.Ternary
	noWayBackRecoverer  gomme.CombiningRecoverer
}

func (fsd *firstSuccessfulData[Output]) happy(state gomme.State) (gomme.State, Output) {
	var zero Output

	// use cache to know result immediately
	result, ok := state.CachedParserResult(fsd.id)
	if ok {
		if result.Failed {
			return state.ErrorAgain(result.Error), zero
		}
		return state.SucceedAgain(result), result.Output.(Output)
	}

	// cache miss: parse
	bestState := state
	idx := 0
	for i, parse := range fsd.parsers {
		newState, output := parse.It(state)
		if !newState.Failed() {
			if state.NoWayBackMoved(newState) {
				state.CacheParserResult(fsd.id, i, i, 0, newState, output)
			} else {
				state.CacheParserResult(fsd.id, i, -1, -1, newState, output)
			}
			return newState, output
		}

		if state.NoWayBackMoved(newState) { // don't look further than this
			state.CacheParserResult(fsd.id, i, i, 0, newState, output)
			return gomme.IWitnessed(state, fsd.id, i, newState), zero
		}

		// may the best error win:
		if i == 0 {
			bestState = newState
		} else {
			bestState = gomme.BetterOf(bestState, newState)
			idx = i
		}
	}
	state.CacheParserResult(fsd.id, idx, idx, 0, bestState, zero)
	return gomme.IWitnessed(state, fsd.id, idx, bestState), zero
}

func (fsd *firstSuccessfulData[Output]) error(state gomme.State) (gomme.State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, HasNoWayBack)
	result, ok := state.CachedParserResult(fsd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(error)` parser",
		), zero
	}
	if result.HasNoWayBack {
		parse := fsd.parsers[result.Idx]
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

func (fsd *firstSuccessfulData[Output]) handle(state gomme.State) (gomme.State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, Failed)
	result, ok := state.CachedParserResult(fsd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(handle)` parser",
		), zero
	}
	if result.Failed {
		newState, output := gomme.HandleWitness(state, fsd.id, result.Idx, fsd.parsers...)
		// the parser failed; so it MUST be the one with the error we are looking for
		if newState.ParsingMode() != gomme.ParsingModeHappy && newState.ParsingMode() != gomme.ParsingModeEscape {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
					"parsing mode `happy` or `escape` in `FirstSuccessful(handle)` parser, but mode is: `%s`",
				result.Idx, fsd.parsers[result.Idx].Expected(), newState.ParsingMode())), zero
		}
		return newState, output
	}
	return state, zero
}

func (fsd *firstSuccessfulData[Output]) rewind(state gomme.State) (gomme.State, Output) {
	var zero Output
	// use cache to know right parser immediately (Idx, Failed)
	result, ok := state.CachedParserResult(fsd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(rewind)` parser",
		), zero
	}
	if result.Failed {
		newState, output := gomme.HandleWitness(state, fsd.id, result.Idx, fsd.parsers...)
		// the parser failed; so it MUST be the one with the error we are looking for
		if newState.ParsingMode() != gomme.ParsingModeHappy && newState.ParsingMode() != gomme.ParsingModeEscape {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
					"parsing mode `happy` or `escape` in `FirstSuccessful(rewind)` parser, but mode is: `%s`",
				result.Idx, fsd.parsers[result.Idx].Expected(), newState.ParsingMode())), zero
		}
		return newState, output
	}
	return state, zero
}

func (fsd *firstSuccessfulData[Output]) escape(state gomme.State) (gomme.State, Output) {
	var zero Output
	if fsd.containingNoWayBack == gomme.TernaryNo { // we can't help
		return state, zero
	}

	idx, ok := fsd.noWayBackRecoverer.CachedIndex(state)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(escape)` parser",
		), zero
	}

	parse := fsd.parsers[idx]
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
