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
	// Construct mySaveSpotRecoverer from the sub-parsers
	//
	subRecoverers := make([]gomme.Recoverer, len(parsers))
	for i, parser := range parsers {
		subRecoverers[i] = parser.SaveSpotRecoverer
	}
	mySaveSpotRecoverer := gomme.NewCombiningRecoverer(true, subRecoverers...)

	fsd := &firstSuccessfulData[Output]{
		id:                gomme.NewBranchParserID(),
		parsers:           parsers,
		saveSpotRecoverer: mySaveSpotRecoverer,
	}

	return gomme.NewParser[Output](
		"FirstSuccessful",
		fsd.any,
		true,
		gomme.DefaultRecovererFunc(fsd.any), // you really shouldn't use this parser as a Recoverer
		mySaveSpotRecoverer.Recover,
	)
}

type firstSuccessfulData[Output any] struct {
	id                uint64
	parsers           []gomme.Parser[Output]
	saveSpotRecoverer gomme.CombiningRecoverer
}

func (fsd *firstSuccessfulData[Output]) any(state gomme.State) (gomme.State, Output) {
	var zero Output

	gomme.Debugf("FirstSuccessful - mode=%s, pos=%d", state.ParsingMode(), state.CurrentPos())
	switch state.ParsingMode() {
	case gomme.ParsingModeHappy: // normal parsing (forward)
		return fsd.happy(state)
	case gomme.ParsingModeError: // find previous SafeSpot (backward)
		return fsd.error(state)
	case gomme.ParsingModeHandle: // find error again (forward)
		return fsd.handle(state)
	case gomme.ParsingModeRewind: // go back to the witness parser (1)
		return fsd.rewind(state)
	case gomme.ParsingModeEscape: // find the SafeSpot recoverer with the least waste
		return fsd.escape(state)
	}

	return state.NewSemanticError(fmt.Sprintf(
		"parsing mode `%s` hasn't been handled in `FirstSuccessful`", state.ParsingMode(),
	)), zero
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
			if state.SaveSpotMoved(newState) {
				state.CacheParserResult(fsd.id, i, i, 0, newState, output)
			} else {
				state.CacheParserResult(fsd.id, i, -1, -1, newState, output)
			}
			return newState, output
		}

		if state.SaveSpotMoved(newState) { // don't look further than this
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
	// use cache to know right parser immediately (Idx, HasSaveSpot)
	result, ok := state.CachedParserResult(fsd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `FirstSuccessful(error)` parser",
		), zero
	}
	if result.HasSaveSpot {
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

	waste, idx, ok := fsd.saveSpotRecoverer.CachedIndex(state)
	if !ok {
		waste = fsd.saveSpotRecoverer.Recover(state)
		idx = fsd.saveSpotRecoverer.LastIndex()
	}

	if idx < 0 {
		return state.MoveBy(state.BytesRemaining()), zero // give up
	}
	parse := fsd.parsers[idx]
	newState, output := parse.It(state.MoveBy(waste))
	// this parser has the best recoverer; so it MUST make us happy again
	if newState.ParsingMode() != gomme.ParsingModeHappy && newState.ParsingMode() != gomme.ParsingModeEscape {
		return state.NewSemanticError(fmt.Sprintf(
			"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
				"parsing mode `happy` or `escape` in `FirstSuccessful(escape)` parser, but mode is: `%s`",
			idx, parse.Expected(), newState.ParsingMode())), zero
	}
	return newState, output
}
