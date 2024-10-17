package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
)

// Delimited parses and discards the result from the prefix parser, then
// parses the result of the main parser, and finally parses and discards
// the result of the suffix parser.
func Delimited[OP, O, OS any](prefix gomme.Parser[OP], parse gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return Map3(prefix, parse, suffix, func(output1 OP, output2 O, output3 OS) (O, error) {
		return output2, nil
	})
}

// Preceded parses and discards a result from the prefix parser. It
// then parses a result from the main parser and returns its result.
func Preceded[OP, O any](prefix gomme.Parser[OP], parse gomme.Parser[O]) gomme.Parser[O] {
	return Map2(prefix, parse, func(output1 OP, output2 O) (O, error) {
		return output2, nil
	})
}

// Terminated parses a result from the main parser, it then
// parses the result from the suffix parser and discards it; only
// returning the result of the main parser.
func Terminated[O, OS any](parse gomme.Parser[O], suffix gomme.Parser[OS]) gomme.Parser[O] {
	return Map2(parse, suffix, func(output1 O, output2 OS) (O, error) {
		return output1, nil
	})
}

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
		switch state.ParsingMode() {
		case gomme.ParsingModeHappy: // normal parsing
			return parseSequenceHappy(state, parsers, 0, outputs, id)
		case gomme.ParsingModeError: // find previous NoWayBack backward
			return parseSequenceError(state, parsers, 0, outputs, id)
		case gomme.ParsingModeHandle: // find error again (forward)
			return parseSequenceHandle(state, parsers, 0, outputs, id)
		case gomme.ParsingModeRewind:
			return parseSequenceRecord(state, parsers, 0, outputs, id)
		case gomme.ParsingModeEscape:
		}
		return state.NewSemanticError(fmt.Sprintf(
			"programming error: Sequence didn't return in mode %v", state.ParsingMode())), []Output{}
	}

	return gomme.NewParser[[]Output](
		"Sequence",
		parseSeq,
		BasicRecovererFunc(parseSeq),
		containsNoWayBack,
		myNoWayBackRecoverer.Recover,
	)
}

// parse function for mode `happy` (normal parsing)
func parseSequenceHappy[Output any](
	state gomme.State,
	parsers []gomme.Parser[Output],
	startIdx int,
	outputs []Output,
	id uint64,
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
	remaining := state
	noWayBackStart := -1
	idx := -1
	for i := startIdx; i < len(parsers); i++ {
		parse := parsers[i]
		newState, output := parse.It(remaining)
		if newState.Failed() {
			state.CacheParserResult(id, i, idx, noWayBackStart, newState, outputs)
			if noWayBackStart >= 0 { // handle error locally
				return parseSequenceError(state, parsers, i, outputs, id)
			}
			return state.Preserve(newState), nil
		}

		if remaining.NoWayBackMoved(newState) {
			idx = i
			noWayBackStart = state.ByteCount(remaining)
		}
		outputs = saveOutput(outputs, output, i)
		remaining = newState
	}

	state.CacheParserResult(id, len(parsers)-1, idx, noWayBackStart, remaining, outputs)
	return remaining, outputs
}

// parse function for mode `error` (find previous NoWayBack)
func parseSequenceError[Output any](
	state gomme.State,
	parsers []gomme.Parser[Output],
	startIdx int,
	outputs []Output,
	id uint64,
) (gomme.State, []Output) {
	var zero []Output

	// use cache to know result immediately (Idx, HasNoWayBack)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `Sequence` parser for parsing mode `error`",
		), zero
	}
	// found in cache
	if result.HasNoWayBack { // we should be able to switch to mode=handle
		newState, _ := parsers[result.NoWayBackIdx].It(state.MoveBy(result.NoWayBackStart))
		if newState.ParsingMode() != gomme.ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) "+
					"didn't switch to parsing mode `record`",
				result.NoWayBackIdx, parsers[result.NoWayBackIdx].Expected())), zero
		}
		return parseSequenceHandle(newState, parsers, result.NoWayBackIdx, outputs, id)
	}
	return state, zero // we can't do anything
}

// parse function for mode `handle` (find error again)
func parseSequenceHandle[Output any](
	state gomme.State,
	parsers []gomme.Parser[Output],
	startIdx int,
	outputs []Output,
	id uint64,
) (gomme.State, []Output) {
	var zero []Output

	// use cache to know result immediately (Idx, ErrorStart)
	result, ok := state.CachedParserResult(id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `Sequence` parser for parsing mode `handle`",
		), zero
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=record
		newState, output := parsers[result.Idx].It(state.MoveBy(result.ErrorStart))
		switch newState.ParsingMode() {
		case gomme.ParsingModeHappy:
			outputs = saveOutput(outputs, output, result.Idx)
			return parseSequenceHappy(newState, parsers, result.Idx+1, outputs, id)
		case gomme.ParsingModeError:
			return parseSequenceError(newState, parsers, result.Idx, outputs, id)
		case gomme.ParsingModeHandle:
			return parseSequenceHandle(newState, parsers, result.Idx+1, outputs, id)
		case gomme.ParsingModeRewind:
			return parseSequenceRecord(newState, parsers, result.Idx+1, outputs, id)
		case gomme.ParsingModeEscape:
		}
	}
	return state, zero // we can't do anything
}

// parse function for mode `handle` (find error again)
func parseSequenceRecord[Output any](
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
