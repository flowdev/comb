package pcb

import (
	"fmt"
	"github.com/oleiade/gomme"
	"slices"
)

// MapN is a helper for easily implementing Map like parsers.
// It is not meant for writing grammars, but only for implementing parsers.
// Only the `fn`n function has to be provided.
// All other `fn`X functions are expected to be `nil`.
// Only parsers up to `p`n have to be provided.
// All higher numbered parsers are expected to be nil.
func MapN[PO1, PO2, PO3, PO4, PO5 any, MO any](
	expected string,
	p1 gomme.Parser[PO1], p2 gomme.Parser[PO2], p3 gomme.Parser[PO3], p4 gomme.Parser[PO4], p5 gomme.Parser[PO5],
	n int,
	fn1 func(PO1) (MO, error), fn2 func(PO1, PO2) (MO, error), fn3 func(PO1, PO2, PO3) (MO, error),
	fn4 func(PO1, PO2, PO3, PO4) (MO, error), fn5 func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) gomme.Parser[MO] {
	var zero1 PO1
	var zero2 PO2
	var zero3 PO3
	var zero4 PO4
	var zero5 PO5

	if p1 == nil {
		panic("MapN: p1 is nil")
	}
	if n >= 2 {
		if p2 == nil {
			panic("MapN: p2 is nil (n >= 2)")
		}
		if n >= 3 {
			if p3 == nil {
				panic("MapN: p3 is nil (n >= 3)")
			}
			if n >= 4 {
				if p4 == nil {
					panic("MapN: p4 is nil (n >= 4)")
				}
				if n >= 5 {
					if p5 == nil {
						panic("MapN: p5 is nil (n >= 5)")
					}
				}
			}
		}
	}

	md := &mapData[PO1, PO2, PO3, PO4, PO5, MO]{
		expected: expected,
		p1:       p1, p2: p2, p3: p3, p4: p4, p5: p5,
		n:   n,
		fn1: fn1, fn2: fn2, fn3: fn3, fn4: fn4, fn5: fn5,
	}

	mapParse := func(state gomme.State) (gomme.State, MO, *gomme.ParserError) {
		return md.ParseAfterChild(gomme.ParseResult{ID: -1, State: state},
			state, state,
			0,
			-1, -1,
			zero1, zero2, zero3, zero4, zero5,
		)
	}

	return gomme.NewParser[MO](
		expected,
		mapParse,
		BasicRecovererFunc(mapParse),
	)
}

type mapData[PO1, PO2, PO3, PO4, PO5 any, MO any] struct {
	expected string
	p1       gomme.Parser[PO1]
	p2       gomme.Parser[PO2]
	p3       gomme.Parser[PO3]
	p4       gomme.Parser[PO4]
	p5       gomme.Parser[PO5]
	n        int
	fn1      func(PO1) (MO, error)
	fn2      func(PO1, PO2) (MO, error)
	fn3      func(PO1, PO2, PO3) (MO, error)
	fn4      func(PO1, PO2, PO3, PO4) (MO, error)
	fn5      func(PO1, PO2, PO3, PO4, PO5) (MO, error)
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) ParseAfterChild(
	childID int32,
	state gomme.State,
	err *gomme.ParserError,
	store gomme.Store,
) (gomme.State, MO, *gomme.ParserError) {
	var zero MO

	gomme.Debugf("MapN - pos=%d", state.CurrentPos())
	if startIdx >= md.n {
		if remaining.ParsingMode() == gomme.ParsingModeHappy {
			return md.mapn(remaining, out1, out2, out3, out4, out5)
		}
		return remaining, zero
	}

	switch remaining.ParsingMode() {
	case gomme.ParsingModeHappy: // normal parsing
		return md.happy(
			state, remaining, startIdx, saveSpotStart, saveSpotIdx,
			out1, out2, out3, out4, out5,
		)
	case gomme.ParsingModeError: // find previous SafeSpot (backward)
		return md.error(state.Preserve(remaining), startIdx, out1, out2, out3, out4, out5)
	case gomme.ParsingModeHandle: // find error again (forward)
		return md.handle(state.Preserve(remaining), startIdx, out1, out2, out3, out4, out5)
	case gomme.ParsingModeRewind: // go back to error / witness parser (1) (backward)
		return md.rewind(state.Preserve(remaining), startIdx, out1, out2, out3, out4, out5)
	case gomme.ParsingModeEscape: // escape the mess the hard way: use recoverer (forward)
		return md.escape(state, remaining, startIdx, out1, out2, out3, out4, out5)
	}
	return state.NewSemanticError(fmt.Sprintf(
		"programming error: MapN didn't handle parsing mode `%s`", state.ParsingMode())), zero

}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) happy(
	state, remaining gomme.State,
	startIdx int,
	saveSpotStart int, saveSpotIdx int,
	out1 PO1, out2 PO2, out3 PO3, out4 PO4, out5 PO5,
) (gomme.State, MO) {
	var zeroMO MO

	if startIdx <= 0 { // caching only works if parsing from the start
		// use cache to know result immediately (Failed, Error, Consumed, Output)
		result, ok := state.CachedParserResult(md.id)
		if ok {
			if result.Failed {
				return state.ErrorAgain(result.Error), zeroMO
			}
			return state.MoveBy(result.Consumed), result.Output.(MO)
		}
	}

	// cache miss: parse
	outputs := make([]interface{}, 0, 4)
	var newState1 gomme.State
	if startIdx <= 0 {
		newState1, out1 = md.p1.It(remaining)
		if newState1.Failed() {
			state.CacheParserResult(md.id, 0, saveSpotIdx, saveSpotStart, newState1, outputs)
			return gomme.IWitnessed(remaining, md.id, 0, newState1), zeroMO
		}
		if state.SaveSpotMoved(newState1) {
			saveSpotIdx = 0
			saveSpotStart = 0
		}
	}
	outputs = append(outputs, out1)

	if md.n > 1 {
		var newState2 gomme.State
		if startIdx <= 1 {
			if startIdx == 1 {
				newState1 = remaining
			}
			newState2, out2 = md.p2.It(newState1)
			if newState2.Failed() {
				state.CacheParserResult(md.id, 1, saveSpotIdx, saveSpotStart, newState2, outputs)
				state = gomme.IWitnessed(newState1, md.id, 0, newState2)
				if saveSpotStart < 0 { // we can't do anything here
					return state, zeroMO
				}
				return md.error(state, 1, out1, out2, out3, out4, out5) // handle error locally
			}
			if newState1.SaveSpotMoved(newState2) {
				saveSpotIdx = 1
				saveSpotStart = state.ByteCount(newState1)
			}
		}
		outputs = append(outputs, out2)

		if md.n > 2 {
			var newState3 gomme.State
			if startIdx <= 2 {
				if startIdx == 2 {
					newState2 = remaining
				}
				newState3, out3 = md.p3.It(newState2)
				if newState3.Failed() {
					state.CacheParserResult(md.id, 2, saveSpotIdx, saveSpotStart, newState3, outputs)
					state = gomme.IWitnessed(newState2, md.id, 0, newState3)
					if saveSpotStart < 0 { // we can't do anything here
						return state, zeroMO
					}
					return md.error(state, 2, out1, out2, out3, out4, out5) // handle error locally
				}
				if newState2.SaveSpotMoved(newState3) {
					saveSpotIdx = 2
					saveSpotStart = state.ByteCount(newState2)
				}
			}
			outputs = append(outputs, out3)

			if md.n > 3 {
				var newState4 gomme.State
				if startIdx <= 3 {
					if startIdx == 3 {
						newState3 = remaining
					}
					newState4, out4 = md.p4.It(newState3)
					if newState4.Failed() {
						state.CacheParserResult(md.id, 3, saveSpotIdx, saveSpotStart, newState4, outputs)
						state = gomme.IWitnessed(newState3, md.id, 0, newState4)
						if saveSpotStart < 0 { // we can't do anything here
							return state, zeroMO
						}
						return md.error(state, 3, out1, out2, out3, out4, out5) // handle error locally
					}
					if newState3.SaveSpotMoved(newState4) {
						saveSpotIdx = 3
						saveSpotStart = state.ByteCount(newState3)
					}
				}
				outputs = append(outputs, out4)

				if md.n > 4 {
					var newState5 gomme.State
					if startIdx == 4 {
						newState4 = remaining
					}
					newState5, out5 = md.p5.It(newState4)
					if newState5.Failed() {
						state.CacheParserResult(md.id, 4, saveSpotIdx, saveSpotStart, newState5, outputs)
						state = gomme.IWitnessed(newState4, md.id, 0, newState5)
						if saveSpotStart < 0 { // we can't do anything here
							return state, zeroMO
						}
						return md.error(state, 4, out1, out2, out3, out4, out5) // handle error locally
					}
					if newState4.SaveSpotMoved(newState5) {
						saveSpotIdx = 4
						saveSpotStart = state.ByteCount(newState4)
					}

					mapped, err := md.fn5(out1, out2, out3, out4, out5)
					if err != nil {
						state.CacheParserResult(md.id, 4, saveSpotIdx, saveSpotStart, newState5, zeroMO)
						return newState5.NewSemanticError(err.Error()), zeroMO
					}
					state.CacheParserResult(md.id, 4, saveSpotIdx, saveSpotStart, newState5, mapped)
					return newState5, mapped
				}
				mapped, err := md.fn4(out1, out2, out3, out4)
				if err != nil {
					state.CacheParserResult(md.id, 3, saveSpotIdx, saveSpotStart, newState4, zeroMO)
					return newState4.NewSemanticError(err.Error()), zeroMO
				}
				state.CacheParserResult(md.id, 3, saveSpotIdx, saveSpotStart, newState4, mapped)
				return newState4, mapped
			}
			mapped, err := md.fn3(out1, out2, out3)
			if err != nil {
				state.CacheParserResult(md.id, 2, saveSpotIdx, saveSpotStart, newState3, zeroMO)
				return newState3.NewSemanticError(err.Error()), zeroMO
			}
			state.CacheParserResult(md.id, 2, saveSpotIdx, saveSpotStart, newState3, mapped)
			return newState3, mapped
		}
		mapped, err := md.fn2(out1, out2)
		if err != nil {
			state.CacheParserResult(md.id, 1, saveSpotIdx, saveSpotStart, newState2, zeroMO)
			return newState2.NewSemanticError(err.Error()), zeroMO
		}
		state.CacheParserResult(md.id, 1, saveSpotIdx, saveSpotStart, newState2, mapped)
		return newState2, mapped
	}
	mapped, err := md.fn1(out1)
	if err != nil {
		state.CacheParserResult(md.id, 0, saveSpotIdx, saveSpotStart, newState1, zeroMO)
		return newState1.NewSemanticError(err.Error()), zeroMO
	}
	state.CacheParserResult(md.id, 0, saveSpotIdx, saveSpotStart, newState1, mapped)
	return newState1, mapped
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) error(
	state gomme.State,
	_ int, // we don't need startIdx because we rely on the cache
	out1 PO1, out2 PO2, out3 PO3, out4 PO4, out5 PO5,
) (gomme.State, MO) {
	var zeroMO MO

	// use cache to know result immediately (HasSaveSpot, SaveSpotIdx, SaveSpotStart)
	result, ok := state.CachedParserResult(md.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `MapN(error)` parser",
		), zeroMO
	}
	// found in cache
	if result.HasSaveSpot { // we should be able to switch to mode=handle
		var newState gomme.State
		expected := ""
		switch result.SaveSpotIdx {
		case 0:
			expected = md.p1.Expected()
			newState, _ = md.p1.It(state)
		case 1:
			expected = md.p2.Expected()
			newState, _ = md.p2.It(state.MoveBy(result.SaveSpotStart))
		case 2:
			expected = md.p3.Expected()
			newState, _ = md.p3.It(state.MoveBy(result.SaveSpotStart))
		case 3:
			expected = md.p4.Expected()
			newState, _ = md.p4.It(state.MoveBy(result.SaveSpotStart))
		default:
			expected = md.p5.Expected()
			newState, _ = md.p5.It(state.MoveBy(result.SaveSpotStart))
		}
		if newState.ParsingMode() != gomme.ParsingModeHandle {
			return state.NewSemanticError(fmt.Sprintf(
				"programming error: sub-parser (index: %d, expected: %q) didn't switch to "+
					"parsing mode `handle` in `MapN(error)` parser, but mode is: `%s`",
				result.SaveSpotIdx, expected, newState.ParsingMode())), zeroMO
		}
		if result.Failed {
			return md.handle(newState, result.Idx, out1, out2, out3, out4, out5)
		}
		return state.Preserve(newState), zeroMO
	}
	return state, zeroMO // we can't do anything
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) handle(
	state gomme.State,
	_ int,
	out1 PO1, out2 PO2, out3 PO3, out4 PO4, out5 PO5,
) (gomme.State, MO) {
	var zeroMO MO

	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(md.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `MapN(handle)` parser",
		), zeroMO
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		var newState gomme.State
		switch result.Idx {
		case 0:
			newState, out1 = gomme.HandleWitness(state, md.id, 0, md.p1)
		case 1:
			newState, out2 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p2,
			)
		case 2:
			newState, out3 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p3,
			)
		case 3:
			newState, out4 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p4,
			)
		default:
			newState, out5 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p5,
			)
		}
		return md.any(
			state, newState,
			result.Idx+1,
			result.SaveSpotStart, result.SaveSpotIdx,
			out1, out2, out3, out4, out5,
		)
	}
	return state, zeroMO // we can't do anything
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) rewind(
	state gomme.State,
	startIdx int,
	out1 PO1, out2 PO2, out3 PO3, out4 PO4, out5 PO5,
) (gomme.State, MO) {
	var zeroMO MO

	gomme.Debugf("MapN.rewind - startIdx=%d", startIdx)
	// use cache to know result immediately (Failed, Idx, ErrorStart)
	result, ok := state.CachedParserResult(md.id)
	if !ok {
		return state.NewSemanticError(
			"grammar error: cache was empty in `MapN(rewind)` parser",
		), zeroMO
	}
	// found in cache
	if result.Failed { // we should be able to switch to mode=happy (or escape)
		gomme.Debugf("MapN.rewind - result.Idx=%d", result.Idx)
		var newState gomme.State
		switch result.Idx {
		case 0:
			newState, out1 = gomme.HandleWitness(state, md.id, 0, md.p1)
		case 1:
			newState, out2 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p2,
			)
		case 2:
			newState, out3 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p3,
			)
		case 3:
			newState, out4 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p4,
			)
		default:
			newState, out5 = gomme.HandleWitness(
				state.MoveBy(result.ErrorStart), md.id, 0, md.p5,
			)
		}

		if newState.ParsingMode() == gomme.ParsingModeRewind && newState.StillHandlingError() {
			gomme.Debugf("MapN.rewind - return (zeroMO): mode=%s", newState.ParsingMode())
			return newState, zeroMO
		}
		gomme.Debugf("MapN.rewind - return (New Round): mode=%s", newState.ParsingMode())
		return md.any(
			state, newState,
			result.Idx+1,
			result.SaveSpotStart, result.SaveSpotIdx,
			out1, out2, out3, out4, out5,
		)
	}
	return state, zeroMO // we can't do anything
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) escape(
	state gomme.State, remaining gomme.State,
	startIdx int,
	out1 PO1, out2 PO2, out3 PO3, out4 PO4, out5 PO5,
) (gomme.State, MO) {
	var zeroMO MO

	idx, waste := 0, 0
	if startIdx <= 0 { // use md.saveSpotRecoverer
		ok := false
		waste, idx, ok = md.saveSpotRecoverer.CachedIndex(state)
		if !ok {
			waste = md.saveSpotRecoverer.Recover(state)
			idx = md.saveSpotRecoverer.LastIndex()
		}
	} else { // we have to use seq.subRecoverers
		recoverers := slices.Clone(md.subRecoverers) // make shallow copy, so we can set the first elements to nil
		for i := 0; i < startIdx; i++ {
			recoverers[i] = nil
		}
		crc := gomme.NewCombiningRecoverer(false, recoverers...)
		waste = crc.Recover(remaining) // find best Recoverer
		idx = crc.LastIndex()
	}

	if idx < 0 { // give up
		return remaining.NewSemanticError(
			"grammar error: found no way to recover from previous error",
		).MoveBy(remaining.BytesRemaining()), zeroMO
	}

	remaining = state.MoveBy(waste)
	var newState gomme.State
	switch idx {
	case 0:
		newState, out1 = md.p1.It(remaining)
	case 1:
		newState, out2 = md.p2.It(remaining)
	case 2:
		newState, out3 = md.p3.It(remaining)
	case 3:
		newState, out4 = md.p4.It(remaining)
	default:
		newState, out5 = md.p5.It(remaining)
	}
	if newState.ParsingMode() == gomme.ParsingModeHappy {
		result, ok := state.CachedParserResult(md.id)
		if !ok {
			result.SaveSpotIdx = -1
			result.SaveSpotStart = -1
		}
		return md.any(state, newState, idx+1, result.SaveSpotIdx, result.SaveSpotStart,
			out1, out2, out3, out4, out5)
	}
	if newState.ParsingMode() == gomme.ParsingModeEscape && !state.Moved(newState) {
		return newState.NewSemanticError(
			"grammar error: found no way to recover from previous error",
		).MoveBy(newState.BytesRemaining()), zeroMO
	}
	return newState, zeroMO // we can't do anything
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) mapn(
	state gomme.State,
	out1 PO1, out2 PO2, out3 PO3, out4 PO4, out5 PO5,
) (gomme.State, MO) {
	var zero, mo MO
	var err error

	switch md.n {
	case 1:
		mo, err = md.fn1(out1)
	case 2:
		mo, err = md.fn2(out1, out2)
	case 3:
		mo, err = md.fn3(out1, out2, out3)
	case 4:
		mo, err = md.fn4(out1, out2, out3, out4)
	case 5:
		mo, err = md.fn5(out1, out2, out3, out4, out5)
	}
	if err != nil {
		return state.NewSemanticError(err.Error()), zero
	}
	return state, mo
}
