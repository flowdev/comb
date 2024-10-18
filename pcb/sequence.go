package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
)

// Sequence applies a sequence of parsers of the same type and
// returns either a slice of results or an error if any parser fails.
// Use one of the MapX parsers for differently typed parsers.
func Sequence[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[[]Output] {
	id := gomme.NewBranchParserID()

	containsNoWayBack := parsers[0].ContainsNoWayBack()
	for i := 1; i < len(parsers); i++ {
		containsNoWayBack = max(containsNoWayBack, parsers[i].ContainsNoWayBack())
	}

	// Construct myNoWayBackRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, len(parsers))
	for i, parser := range parsers {
		if parser.ContainsNoWayBack() > gomme.TernaryNo {
			subRecoverers[i] = parser.NoWayBackRecoverer
		}
	}
	myNoWayBackRecoverer := gomme.NewCombiningRecoverer(subRecoverers...)

	// finally the parse function
	parseSeq := func(state gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, len(parsers))
		return sequenceAny(id, parsers, state, 0, outputs)
	}

	return gomme.NewParser[[]Output](
		"Sequence",
		parseSeq,
		true,
		BasicRecovererFunc(parseSeq),
		containsNoWayBack,
		myNoWayBackRecoverer.Recover,
	)
}

func sequenceAny[Output any](
	id uint64,
	parsers []gomme.Parser[Output],
	state gomme.State,
	startIdx int,
	outputs []Output,
) (gomme.State, []Output) {
	if startIdx >= len(parsers) {
		return state, outputs
	}
	switch state.ParsingMode() {
	case gomme.ParsingModeHappy: // normal parsing
		return sequenceHappy(id, parsers, state, state, -1, -1, startIdx, outputs)
	case gomme.ParsingModeError: // find previous NoWayBack (backward)
		return sequenceError(id, parsers, state, startIdx, outputs)
	case gomme.ParsingModeHandle: // find error again (forward)
		return sequenceHandle(id, parsers, state, startIdx, outputs)
	case gomme.ParsingModeRewind: // go back to error / witness parser (1) (backward)
		return sequenceRewind(state, parsers, startIdx, outputs, id)
	case gomme.ParsingModeEscape: // escape the mess the hard way: use recoverer (forward)
	}
	return state.NewSemanticError(fmt.Sprintf(
		"programming error: Sequence didn't return in mode %v", state.ParsingMode())), []Output{}

}

func sequenceHappy[Output any]( // normal parsing (forward)
	id uint64,
	parsers []gomme.Parser[Output],
	state gomme.State,
	remaining gomme.State,
	startIdx int,
	noWayBackStart int,
	noWayBackIdx int,
	outputs []Output,
) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Error, Consumed, Output)
	result, ok := state.CachedParserResult(id)
	if ok {
		if result.Failed {
			return state.ErrorAgain(result.Error), nil
		}
		return state.MoveBy(result.Consumed), result.Output.([]Output)
	}

	// cache miss: parse
	for i := startIdx; i < len(parsers); i++ {
		parse := parsers[i]
		newState, output := parse.It(remaining)
		if newState.Failed() {
			state.CacheParserResult(id, i, noWayBackIdx, noWayBackStart, newState, outputs)
			remaining = gomme.IWitnessed(remaining, id, i, newState)
			if noWayBackStart < 0 { // we can't do anything here
				return state.Preserve(remaining), nil
			}
			return sequenceError(id, parsers, state, noWayBackIdx, outputs) // handle error locally
		}

		if remaining.NoWayBackMoved(newState) {
			noWayBackIdx = i
			noWayBackStart = state.ByteCount(remaining)
		}
		outputs = saveOutput(outputs, output, i)
		remaining = newState
	}

	state.CacheParserResult(id, len(parsers)-1, noWayBackIdx, noWayBackStart, remaining, outputs)
	return remaining, outputs
}

func sequenceError[Output any](
	id uint64,
	parsers []gomme.Parser[Output],
	state gomme.State,
	_ int, // we don't need `startIdx` because we rely on the cache
	outputs []Output,
) (gomme.State, []Output) {
	// use cache to know result immediately (HasNoWayBack, NoWayBackIdx, NoWayBackStart)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `Sequence(error)` parser",
		), nil
	}
	// found in cache
	if result.HasNoWayBack { // we should be able to switch to mode=handle
		newState, _ := parsers[result.NoWayBackIdx].It(state.MoveBy(result.NoWayBackStart))
		if newState.ParsingMode() != gomme.ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) "+
					"didn't switch to parsing mode `handle` in `Sequence(error)` parser",
				result.NoWayBackIdx, parsers[result.NoWayBackIdx].Expected())), nil
		}
		if result.Failed {
			return sequenceHandle(id, parsers, newState, result.Idx, outputs)
		}
	}
	return state, nil // we can't do anything
}

func sequenceHandle[Output any]( // find error again (forward)
	id uint64,
	parsers []gomme.Parser[Output],
	state gomme.State,
	_ int, // we don't need `startIdx` because we rely on the cache
	outputs []Output,
) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `Sequence` parser for parsing mode `handle`",
		), nil
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		newState, output := gomme.HandleWitness(state.MoveBy(result.ErrorStart), id, 0, parsers[result.Idx])
		outputs = saveOutput(outputs, output, result.Idx)
		return sequenceAny(id, parsers, newState, result.Idx+1, outputs)
	}
	return state, nil // we can't do anything
}

func sequenceRewind[Output any]( // go back to witness parser (1) (backward)
	state gomme.State,
	parsers []gomme.Parser[Output],
	startIdx int,
	outputs []Output,
	id uint64,
) (gomme.State, []Output) {
	return state, outputs // we can't do anything
}

func saveOutput[Output any](outputs []Output, output Output, i int) []Output {
	var zero Output
	for i >= len(outputs) {
		outputs = append(outputs, zero)
	}
	outputs[i] = output
	return outputs
}
