package pcb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/oleiade/gomme"
)

// noSeparator is a parser used to signal that no separator should be parsed at all.
// It basically turns SeparatedMN into ManyMN.
var noSeparator = func() gomme.Parser[rune] {
	// created a unique marker as expected string
	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	expected := "NO-SEPARATOR:"
	if err != nil {
		expected += err.Error()
	} else {
		expected += hex.EncodeToString(bs)
	}
	return gomme.NewParser[rune](expected, nil, false, nil, nil)
}()

// SeparatedMN applies an element parser and a separator parser repeatedly in order
// to produce a slice of elements.
//
// Because the `SeparatedListMN` is really looking to produce a list of elements resulting
// from the provided main parser, it will succeed even if the separator parser fails to
// match at the end.
//
// The parser will fail if both parsers together accepted an empty input
// in order to prevent infinite loops.
func SeparatedMN[Output any, S gomme.Separator](
	parse gomme.Parser[Output], separator gomme.Parser[S],
	atLeast, atMost int,
	parseSeparatorAtEnd bool,
) gomme.Parser[[]Output] {
	if atLeast < 0 {
		panic("SeparatedMN is unable to handle negative `atLeast`")
	}
	if atMost < 0 {
		panic("SeparatedMN is unable to handle negative `atMost`")
	}

	md := &separatedData[Output, S]{
		id:                  gomme.NewBranchParserID(),
		parse:               parse,
		separator:           separator,
		atLeast:             atLeast,
		atMost:              atMost,
		parseSeparatorAtEnd: parseSeparatorAtEnd,
	}
	parseSep := func(state gomme.State) (gomme.State, []Output) {
		outputs := make([]Output, 0, min(32, md.atMost))
		return md.any(state, state, -1, -1, outputs)
	}

	recoverer := Forbidden("SeparatedMN(atLeast=0)")
	if atLeast > 0 {
		recoverer = BasicRecovererFunc(parseSep)
	}
	return gomme.NewParser[[]Output]("SeparatedMN", parseSep, true,
		recoverer, parse.NoWayBackRecoverer)
}

type separatedData[Output any, S gomme.Separator] struct {
	id                  uint64
	parse               gomme.Parser[Output]
	separator           gomme.Parser[S]
	atLeast             int
	atMost              int
	parseSeparatorAtEnd bool
}

func (sd *separatedData[Output, S]) any(
	state, remaining gomme.State,
	noWayBackIdx, noWayBackStart int,
	outputs []Output,
) (gomme.State, []Output) {
	count := len(outputs)

	gomme.Debugf("SeparatedMN - mode=%s, pos=%d, count=%d", remaining.ParsingMode(), remaining.CurrentPos(), count)
	if count >= sd.atMost {
		return remaining, outputs
	}
	switch remaining.ParsingMode() {
	case gomme.ParsingModeHappy: // normal parsing
		return sd.happy(state, remaining, count, noWayBackIdx, noWayBackStart, outputs)
	case gomme.ParsingModeError: // find previous NoWayBack (backward)
		return sd.error(state, outputs)
	case gomme.ParsingModeHandle: // find error again (forward)
		return sd.handle(state, outputs)
	case gomme.ParsingModeRewind: // go back to error / witness parser (1) (backward)
		return sd.rewind(state, outputs)
	case gomme.ParsingModeEscape: // escape the mess the hard way: use recoverer (forward)
		return sd.escape(state, remaining, outputs)
	}
	return state.NewSemanticError(fmt.Sprintf(
		"programming error: SeparatedMN didn't handle parsing mode `%s`", state.ParsingMode())), nil

}

func (sd *separatedData[Output, S]) happy(
	state, remaining gomme.State,
	count int,
	noWayBackIdx, noWayBackStart int,
	outputs []Output,
) (gomme.State, []Output) {
	retState := remaining

	for {
		if count >= sd.atMost {
			return retState, outputs
		}

		newState, output := sd.parse.It(remaining)
		if newState.Failed() {
			if remaining.NoWayBackMoved(newState) { // fail because of NoWayBack
				state.CacheParserResult(sd.id, 0, noWayBackIdx, noWayBackStart, newState, outputs)
				state = gomme.IWitnessed(state, sd.id, 0, newState)
				return sd.error(state, outputs)
			}
			if count >= sd.atLeast { // success!
				state.CacheParserResult(sd.id, 0, noWayBackIdx, noWayBackStart, retState, outputs)
				return retState, outputs
			}
			// fail:
			state.CacheParserResult(sd.id, 0, noWayBackIdx, noWayBackStart, newState, outputs)
			state = gomme.IWitnessed(state, sd.id, 0, newState)
			if noWayBackStart < 0 { // we can't do anything here
				return state, nil
			}
			return sd.error(state, outputs) // handle error locally
		}
		if remaining.NoWayBackMoved(newState) {
			noWayBackIdx = 0
			noWayBackStart = state.ByteCount(remaining)
		}
		outputs = append(outputs, output)
		count++

		retState = newState
		sepState := newState
		if sd.separator.Expected() != noSeparator.Expected() {
			sepState, _ = sd.separator.It(newState)
			if sepState.Failed() {
				if newState.NoWayBackMoved(sepState) { // fail because of NoWayBack
					state.CacheParserResult(sd.id, 1, noWayBackIdx, noWayBackStart, sepState, outputs)
					state = gomme.IWitnessed(state, sd.id, 1, sepState)
					return sd.error(state, outputs)
				}
				if count >= sd.atLeast { // success!
					state.CacheParserResult(sd.id, 1, noWayBackIdx, noWayBackStart, newState, outputs)
					return retState, outputs
				}
				// fail:
				state.CacheParserResult(sd.id, 1, noWayBackIdx, noWayBackStart, sepState, outputs)
				state = gomme.IWitnessed(state, sd.id, 1, sepState)
				if noWayBackStart < 0 { // we can't do anything here
					return state, nil
				}
				return sd.error(state, outputs) // handle error locally
			}
			if newState.NoWayBackMoved(sepState) {
				noWayBackIdx = 1
				noWayBackStart = state.ByteCount(newState)
			}
			if sd.parseSeparatorAtEnd {
				retState = sepState
			}
		}

		// Checking for infinite loops, if nothing was consumed,
		// the provided parser would make us go around in circles.
		if !sepState.Moved(remaining) {
			return state.NewError(fmt.Sprintf(
				"many %s (empty element incl. separator => endless loop)", sd.parse.Expected())), nil
		}
		remaining = sepState
	}
}

func (sd *separatedData[Output, S]) error(state gomme.State, outputs []Output) (gomme.State, []Output) {
	// use cache to know result immediately (HasNoWayBack, NoWayBackIdx, NoWayBackStart)
	result, ok := state.CachedParserResult(sd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `SeparatedMN(error)` parser",
		), nil
	}
	// found in cache
	if result.HasNoWayBack { // we should be able to switch to mode=handle
		newState := state
		if result.Idx == 0 {
			newState, _ = sd.parse.It(state.MoveBy(result.NoWayBackStart))
		} else {
			newState, _ = sd.separator.It(state.MoveBy(result.NoWayBackStart))
		}
		if newState.ParsingMode() != gomme.ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (expected: %q) didn't switch to "+
					"parsing mode `handle` in `SeparatedMN(error)` parser, but mode is: `%s`",
				sd.parse.Expected(), newState.ParsingMode())), nil
		}
		if result.Failed {
			return sd.handle(newState, outputs)
		}
		return state.Preserve(newState), nil
	}
	return state, nil // we can't do anything
}

func (sd *separatedData[Output, S]) handle(state gomme.State, outputs []Output) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(sd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `SeparatedMN(handle)` parser",
		), nil
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		outputs = result.Output.([]Output)
		newState, output := gomme.HandleWitness(
			state.MoveBy(result.ErrorStart), sd.id, result.Idx, sd.parse, gomme.ParserToZeroOutput[Output, S](sd.separator),
		)
		outputs = append(outputs, output)
		return sd.any(
			state,
			newState,
			result.NoWayBackStart,
			result.NoWayBackIdx,
			outputs,
		)
	}
	return state, nil // we can't do anything
}

func (sd *separatedData[Output, S]) rewind(state gomme.State, outputs []Output) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(sd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `SeparatedMN(rewind)` parser",
		), nil
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		outputs = result.Output.([]Output)
		newState, output := gomme.HandleWitness(
			state.MoveBy(result.ErrorStart), sd.id, result.Idx, sd.parse, gomme.ParserToZeroOutput[Output, S](sd.separator),
		)
		outputs = append(outputs, output)
		return sd.any(
			state, newState,
			result.NoWayBackStart, result.NoWayBackIdx,
			outputs,
		)
	}
	return state, nil // we can't do anything
}

func (sd *separatedData[Output, S]) escape(state, remaining gomme.State, outputs []Output) (gomme.State, []Output) {
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(sd.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `SeparatedMN(escape)` parser",
		), nil
	}

	waste := 0
	if result.NoWayBackIdx == 0 {
		waste = sd.parse.NoWayBackRecoverer(remaining)
	} else {
		waste = sd.separator.NoWayBackRecoverer(remaining)
	}

	if waste < 0 { // give up
		return remaining.NewSemanticError(
			"found no way to recover from previous error",
		).MoveBy(remaining.BytesRemaining()), nil
	}

	outputs = result.Output.([]Output)
	remaining = remaining.MoveBy(waste)
	var newState gomme.State
	var output Output
	if result.NoWayBackIdx == 0 {
		newState, output = sd.parse.It(remaining)
		if newState.ParsingMode() == gomme.ParsingModeHappy {
			outputs = append(outputs, output)
		}
	} else {
		newState, _ = sd.separator.It(remaining)
	}
	return newState, outputs
}
