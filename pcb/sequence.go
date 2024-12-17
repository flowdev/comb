package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
	"slices"
)

// Sequence applies a sequence of parsers of the same type and
// returns either a slice of results or an error if any parser fails.
// Use one of the MapX parsers for differently typed parsers.
func Sequence[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[[]Output] {
	// Construct mySaveSpotRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, len(parsers))
	for i, parser := range parsers {
		subRecoverers[i] = parser.SaveSpotRecoverer
	}
	mySaveSpotRecoverer := gomme.NewCombiningRecoverer(true, subRecoverers...)

	seq := &sequenceData[Output]{
		id:                gomme.NewBranchParserID(),
		parsers:           parsers,
		saveSpotRecoverer: mySaveSpotRecoverer,
		subRecoverers:     subRecoverers,
	}

	// finally the parse function
	parseSeq := func(state gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, len(parsers))
		return seq.any(
			state, state,
			0,
			-1, -1,
			outputs,
		)
	}

	myRecoverer := BasicRecovererFunc(parseSeq)
	if len(parsers) == 1 {
		myRecoverer = parsers[0].MyRecoverer()
	}

	return gomme.NewParser[[]Output](
		"Sequence",
		parseSeq,
		true,
		myRecoverer,
		mySaveSpotRecoverer.Recover,
	)
}

type sequenceData[Output any] struct {
	id                uint64
	parsers           []gomme.Parser[Output]
	saveSpotRecoverer gomme.CombiningRecoverer
	subRecoverers     []gomme.Recoverer
}

func (seq *sequenceData[Output]) any(
	state, remaining gomme.State,
	startIdx int,
	saveSpotIdx, saveSpotStart int,
	outputs []Output,
) (gomme.State, []Output) {
	gomme.Debugf("Sequence - mode=%s, pos=%d, startIdx=%d", remaining.ParsingMode(), remaining.CurrentPos(), startIdx)
	if startIdx >= len(seq.parsers) {
		return remaining, outputs
	}
	switch remaining.ParsingMode() {
	case gomme.ParsingModeHappy: // normal parsing
		return seq.happy(state, remaining, startIdx, saveSpotStart, saveSpotIdx, outputs)
	case gomme.ParsingModeError: // find previous SaveSpot (backward)
		return seq.error(state, startIdx, outputs)
	case gomme.ParsingModeHandle: // find error again (forward)
		return seq.handle(state, startIdx, outputs)
	case gomme.ParsingModeRewind: // go back to error / witness parser (1) (backward)
		return seq.rewind(state, startIdx, outputs)
	case gomme.ParsingModeEscape: // escape the mess the hard way: use recoverer (forward)
		return seq.escape(state, remaining, startIdx, outputs)
	}
	return state.NewSemanticError(fmt.Sprintf(
		"programming error: Sequence didn't handle parsing mode `%s`", state.ParsingMode())), nil

}

func (seq *sequenceData[Output]) happy( // normal parsing (forward)
	state, remaining gomme.State,
	startIdx int,
	saveSpotStart, saveSpotIdx int,
	outputs []Output,
) (gomme.State, []Output) {
	if startIdx <= 0 { // caching only works if parsing from the start
		// use cache to know result immediately (Failed, Error, Consumed, Output)
		result, ok := state.CachedParserResult(seq.id)
		if ok {
			if result.Failed {
				return state.ErrorAgain(result.Error), nil
			}
			return state.MoveBy(result.Consumed), result.Output.([]Output)
		}
	}

	// cache miss: parse
	for i := startIdx; i < len(seq.parsers); i++ {
		parse := seq.parsers[i]
		newState, output := parse.It(remaining)
		if newState.Failed() {
			state.CacheParserResult(seq.id, i, saveSpotIdx, saveSpotStart, newState, outputs)
			state = gomme.IWitnessed(remaining, seq.id, i, newState)
			if saveSpotStart < 0 { // we can't do anything here
				return state, nil
			}
			return seq.error(state, i, outputs) // handle error locally
		}

		if remaining.SaveSpotMoved(newState) {
			saveSpotIdx = i
			saveSpotStart = state.ByteCount(remaining)
		}
		outputs = saveOutput(outputs, output, i)
		remaining = newState
	}

	state.CacheParserResult(seq.id, len(seq.parsers)-1, saveSpotIdx, saveSpotStart, remaining, outputs)
	return remaining, outputs
}

func (seq *sequenceData[Output]) error(
	state gomme.State,
	_ int, // we don't need `startIdx` because we rely on the cache
	outputs []Output,
) (gomme.State, []Output) {
	// use cache to know result immediately (HasSaveSpot, SaveSpotIdx, SaveSpotStart)
	result, ok := state.CachedParserResult(seq.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `Sequence(error)` parser",
		), nil
	}
	// found in cache
	if result.HasSaveSpot { // we should be able to switch to mode=handle
		parse := seq.parsers[result.SaveSpotIdx]
		newState, _ := parse.It(state.MoveBy(result.SaveSpotStart))
		if newState.ParsingMode() != gomme.ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
					"parsing mode `handle` in `Sequence(error)` parser, but mode is: `%s`",
				result.SaveSpotIdx, parse.Expected(), newState.ParsingMode())), nil
		}
		if result.Failed {
			return seq.handle(newState, result.Idx, outputs)
		}
		return state.Preserve(newState), nil
	}
	return state, nil // we can't do anything
}

func (seq *sequenceData[Output]) handle( // find error again (forward)
	state gomme.State,
	_ int, // we don't need `startIdx` because we rely on the cache
	outputs []Output,
) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(seq.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `Sequence(handle)` parser",
		), nil
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		outputs = result.Output.([]Output)
		newState, output := gomme.HandleWitness(
			state.MoveBy(result.ErrorStart), seq.id, result.Idx, seq.parsers...,
		)
		outputs = saveOutput(outputs, output, result.Idx)
		return seq.any(
			state,
			newState,
			result.Idx+1,
			result.SaveSpotIdx,
			result.SaveSpotStart,
			outputs,
		)
	}
	return state, nil // we can't do anything
}

func (seq *sequenceData[Output]) rewind(
	state gomme.State,
	_ int, // we don't need `startIdx` because we rely on the cache
	outputs []Output,
) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(seq.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `Sequence(rewind)` parser",
		), nil
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		outputs = result.Output.([]Output)
		newState, output := gomme.HandleWitness(
			state.MoveBy(result.ErrorStart), seq.id, result.Idx, seq.parsers...,
		)
		outputs = saveOutput(outputs, output, result.Idx)
		return seq.any(
			state,
			newState,
			result.Idx+1,
			result.SaveSpotIdx,
			result.SaveSpotStart,
			outputs,
		)
	}
	return state, nil // we can't do anything
}

func (seq *sequenceData[Output]) escape(
	state, remaining gomme.State,
	startIdx int,
	outputs []Output,
) (gomme.State, []Output) {
	idx, waste := 0, 0
	if startIdx <= 0 { // use seq.saveSpotRecoverer
		ok := false
		waste, idx, ok = seq.saveSpotRecoverer.CachedIndex(state)
		if !ok {
			waste = seq.saveSpotRecoverer.Recover(state)
			idx = seq.saveSpotRecoverer.LastIndex()
		}
	} else { // we have to use seq.subRecoverers
		recoverers := slices.Clone(seq.subRecoverers) // make shallow copy, so we can set the first elements to nil
		for i := 0; i < startIdx; i++ {
			recoverers[i] = nil
		}
		crc := gomme.NewCombiningRecoverer(false, recoverers...)
		waste = crc.Recover(remaining) // find best Recoverer
		idx = crc.LastIndex()
	}

	if idx < 0 {
		return remaining.NewSemanticError(
			"grammar error: unable to recover; did you forget to use the SaveSpot parser?",
		).MoveBy(remaining.BytesRemaining()), nil // give up!
	}
	newState, output := seq.parsers[idx].It(remaining.MoveBy(waste))
	if newState.ParsingMode() == gomme.ParsingModeHappy {
		outputs = saveOutput(outputs, output, idx)
	}
	return newState, outputs
}

func saveOutput[Output any](outputs []Output, output Output, i int) []Output {
	var zero Output
	for i >= len(outputs) {
		outputs = append(outputs, zero)
	}
	outputs[i] = output
	return outputs
}
