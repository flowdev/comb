package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
)

// ManyMN applies a parser repeatedly until it fails, and returns a slice of all
// the results as the Result's Output.
//
// Note that ManyMN fails if the provided parser accepts empty inputs (such as
// `Digit0`, or `Alpha0`) in order to prevent infinite loops.
func ManyMN[Output any](parse gomme.Parser[Output], atLeast, atMost int) gomme.Parser[[]Output] {
	if atLeast < 0 {
		panic("ManyMN is unable to handle negative `atLeast` argument")
	}
	if atMost < 0 {
		panic("ManyMN is unable to handle negative `atMost` argument")
	}

	md := &manyData[Output]{
		id:      gomme.NewBranchParserID(),
		parse:   parse,
		atLeast: atLeast,
		atMost:  atMost,
	}

	parseMany := func(state gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, min(32, atMost))
		return md.any(state, state, -1, -1, outputs)
	}

	recoverer := Forbidden("Many(atLeast=0)")
	containsNoWayBack := gomme.TernaryNo
	if atLeast > 0 {
		recoverer = BasicRecovererFunc(parseMany)
		containsNoWayBack = parse.ContainsNoWayBack()
	}
	return gomme.NewParser[[]Output]("ManyMN", parseMany, true, recoverer,
		containsNoWayBack, parse.NoWayBackRecoverer)
}

type manyData[Output any] struct {
	id      uint64
	parse   gomme.Parser[Output]
	atLeast int
	atMost  int
}

func (md *manyData[Output]) any(
	state, remaining gomme.State,
	noWayBackIdx, noWayBackStart int,
	outputs []Output,
) (gomme.State, []Output) {
	count := len(outputs)

	if count >= md.atMost {
		return state, outputs
	}
	switch state.ParsingMode() {
	case gomme.ParsingModeHappy: // normal parsing
		return md.happy(state, remaining, count, noWayBackIdx, noWayBackStart, outputs)
	case gomme.ParsingModeError: // find previous NoWayBack (backward)
		return md.error(state, outputs)
	case gomme.ParsingModeHandle: // find error again (forward)
		return md.handle(state, outputs)
	case gomme.ParsingModeRewind: // go back to error / witness parser (1) (backward)
		return md.rewind(state, outputs)
	case gomme.ParsingModeEscape: // escape the mess the hard way: use recoverer (forward)
		return md.escape(state, remaining, outputs)
	}
	return state.NewSemanticError(fmt.Sprintf(
		"programming error: ManyMN didn't handle parsing mode `%s`", state.ParsingMode())), nil

}

func (md *manyData[Output]) happy(
	state, remaining gomme.State,
	count int,
	noWayBackIdx, noWayBackStart int,
	outputs []Output,
) (gomme.State, []Output) {
	for {
		if count >= md.atMost {
			return remaining, outputs
		}
		newState, output := md.parse.It(remaining)
		if newState.Failed() {
			if remaining.NoWayBackMoved(newState) { // fail because of NoWayBack
				state.CacheParserResult(md.id, 0, noWayBackIdx, noWayBackStart, newState, outputs)
				state = gomme.IWitnessed(state, md.id, 0, newState)
				return md.error(state, outputs)
			}
			if count >= md.atLeast { // success!
				state.CacheParserResult(md.id, 0, noWayBackIdx, noWayBackStart, remaining, outputs)
				return remaining, outputs
			}
			// fail:
			state.CacheParserResult(md.id, 0, noWayBackIdx, noWayBackStart, newState, outputs)
			state = gomme.IWitnessed(state, md.id, 0, newState)
			if noWayBackStart < 0 { // we can't do anything here
				return state, nil
			}
			return md.error(state, outputs) // handle error locally
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if !newState.Moved(remaining) {
			return state.NewError(fmt.Sprintf(
				"%s (found empty element, endless loop)", md.parse.Expected())), nil
		}

		if remaining.NoWayBackMoved(newState) {
			noWayBackIdx = 0
			noWayBackStart = state.ByteCount(remaining)
		}
		outputs = append(outputs, output)
		remaining = newState
		count++
	}
}

func (md *manyData[Output]) error(state gomme.State, outputs []Output) (gomme.State, []Output) {
	// use cache to know result immediately (HasNoWayBack, NoWayBackIdx, NoWayBackStart)
	result, ok := state.CachedParserResult(md.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `ManyMN(error)` parser",
		), nil
	}
	// found in cache
	if result.HasNoWayBack { // we should be able to switch to mode=handle
		newState, _ := md.parse.It(state.MoveBy(result.NoWayBackStart))
		if newState.ParsingMode() != gomme.ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (expected: %q) didn't switch to "+
					"parsing mode `handle` in `ManyMN(error)` parser, but mode is: `%s`",
				md.parse.Expected(), newState.ParsingMode())), nil
		}
		if result.Failed {
			return md.handle(newState, outputs)
		}
		return state.Preserve(newState), nil
	}
	return state, nil // we can't do anything
}

func (md *manyData[Output]) handle(state gomme.State, outputs []Output) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(md.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `ManyMN(handle)` parser",
		), nil
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		outputs = result.Output.([]Output)
		newState, output := gomme.HandleWitness(
			state.MoveBy(result.ErrorStart), md.id, result.Idx, md.parse,
		)
		outputs = append(outputs, output)
		return md.any(
			state,
			newState,
			result.NoWayBackStart,
			result.NoWayBackIdx,
			outputs,
		)
	}
	return state, nil // we can't do anything
}

func (md *manyData[Output]) rewind(state gomme.State, outputs []Output) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(md.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `ManyMN(rewind)` parser",
		), nil
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		outputs = result.Output.([]Output)
		newState, output := gomme.HandleWitness(
			state.MoveBy(result.ErrorStart), md.id, result.Idx, md.parse,
		)
		outputs = append(outputs, output)
		return md.any(
			state, newState,
			result.NoWayBackStart, result.NoWayBackIdx,
			outputs,
		)
	}
	return state, nil // we can't do anything
}

func (md *manyData[Output]) escape(state, remaining gomme.State, outputs []Output) (gomme.State, []Output) {
	if md.parse.ContainsNoWayBack() == gomme.TernaryNo {
		return state.Preserve(remaining.NewSemanticError(
			"programming error: no recoverer found in `ManyMN(escape)` parser ")), nil
	}
	newState, output := md.parse.It(remaining)
	if newState.ParsingMode() == gomme.ParsingModeHappy {
		outputs = append(outputs, output)
	}
	return newState, outputs
}
