package gomme

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

const errorMarker = 0x25B6 // easy to spot marker (▶) for exact error position

// pcbError is an error message from the parser.
// It consists of the text itself and the position in the input where it happened.
type pcbError struct {
	text      string
	pos       int // pos is the byte index in the input (state.input.pos)
	line, col int // col is the 0-based byte index within srcLine; convert to 1-based rune index for user
	srcLine   string
}

// errHand contains all data needed for handling one error.
type errHand struct {
	err             *pcbError // error that is currently handled
	witnessID       uint64    // ID of the immediate parent branch parser that witnessed the error
	witnessPos      int       // input position of the witness parser
	culpritIdx      int       // index of the sub-parser that created the error
	curDel          int       // current number of tokes to delete for error handling
	ignoreErrParser bool      // true if the failing parser should be ignored
	orgPos          int       // state.input.pos before starting to use deleter
	orgLine         int       // state.input.line before starting to use deleter
	orgPrevNl       int       // state.input.prevNl before starting to use deleter
}

// IWitnessed lets a branch parser report an error that it witnessed in
// the sub-parser with index `idx` (0 if it has only 1 sub-parser).
func IWitnessed(state State, witnessID uint64, idx int, errState State) State {
	state.saveSpot = max(state.saveSpot, errState.saveSpot)
	state.mode = errState.mode
	if errState.errHand.witnessID == 0 { // error hasn't been witnessed yet
		if idx < 0 {
			idx = 0
		}
		errState.errHand.witnessID = witnessID
		errState.errHand.witnessPos = state.input.pos
		errState.errHand.culpritIdx = idx
	} else if errState.errHand.ignoreErrParser || errState.errHand.curDel > 0 { // we try to recover
		errState.mode = ParsingModeRewind
	}
	state.errHand = errState.errHand
	return state
}

// HandleWitness returns the advanced state and output if the parser is
// the witness parser (1).
// If the branch parser isn't the witness, the sub-parser with index `idx` is used.
// If `state.maxDel` is 0, error handling is turned off and the state is returned
// with mode `escape` at EOF position.
func HandleWitness[Output any](state State, id uint64, idx int, parsers ...Parser[Output]) (State, Output) {
	var output, zero Output

	if state.maxDel <= 0 { // error handling is turned off
		state.mode = ParsingModeEscape
		return state.MoveBy(state.BytesRemaining()), zero
	}
	if state.mode == ParsingModeEscape && state.AtEnd() { // stop riding a dead horse
		return state, output
	}

	if state.errHand.witnessID != id || state.errHand.witnessPos != state.input.pos {
		parse := parsers[idx]
		if parse.PossibleWitness() {
			return parse.It(state) // this sub-parser or one of its sub-parsers might be the witness parser (1)
		}
		return state, zero
	}

	// we are witness
	if state.errHand.orgPos == 0 && state.errHand.orgLine == 0 && state.errHand.orgPrevNl == 0 {
		state.errHand.orgPos = state.input.pos
		state.errHand.orgLine = state.input.line
		state.errHand.orgPrevNl = state.input.prevNl
	}
	if state.AtEnd() { // we can't do anything -> give up
		state.input.pos = state.errHand.orgPos
		state.input.line = state.errHand.orgLine
		state.input.prevNl = state.errHand.orgPrevNl
		state.errHand.curDel = state.maxDel
		state.errHand.ignoreErrParser = true
		state.mode = ParsingModeEscape
		Debugf("HandleWitness - EOF -> escape: curDel=%d, ignoreErrParser=%t", state.errHand.curDel, state.errHand.ignoreErrParser)
		return state, zero
	}
	if state.errHand.culpritIdx >= len(parsers) {
		state = state.NewSemanticError(fmt.Sprintf(
			"programming error: length of sub-parsers is only %d but index of culprit sub-parser is %d",
			len(parsers), state.errHand.culpritIdx,
		))
		state.errHand.culpritIdx = len(parsers) - 1
	}
	parse := parsers[state.errHand.culpritIdx]
	for {
		switch state.mode {
		case ParsingModeHandle:
			state.errHand.curDel = 1
			state.errHand.ignoreErrParser = false
			Debugf("HandleWitness - handle: curDel=%d, ignoreErrParser=%t", state.errHand.curDel, state.errHand.ignoreErrParser)
		case ParsingModeRewind:
			state.errHand.curDel++
			if state.errHand.curDel > state.maxDel {
				if !state.errHand.ignoreErrParser {
					state.input.pos = state.errHand.orgPos
					state.input.line = state.errHand.orgLine
					state.input.prevNl = state.errHand.orgPrevNl
					state.errHand.curDel = 0
					state.errHand.ignoreErrParser = true
				} else {
					state.input.pos = state.errHand.orgPos
					state.input.line = state.errHand.orgLine
					state.input.prevNl = state.errHand.orgPrevNl
					state.mode = ParsingModeEscape // give up and go the hard way
					Debugf("HandleWitness - rewind -> escape: curDel=%d, ignoreErrParser=%t", state.errHand.curDel, state.errHand.ignoreErrParser)
					return state, zero
				}
			}
			Debugf("HandleWitness - rewind: curDel=%d, ignoreErrParser=%t", state.errHand.curDel, state.errHand.ignoreErrParser)
		default:
			Debugf("HandleWitness - %s: curDel=%d, ignoreErrParser=%t", state.mode, state.errHand.curDel, state.errHand.ignoreErrParser)
			return state, zero // we are witness parser but there is nothing to do
		}
		state.mode = ParsingModeHappy // try again
		state.errHand.err = nil
		oldRemaining := state.BytesRemaining()
		state = state.deleter(state, min(state.errHand.curDel, 1))
		if oldRemaining > state.BytesRemaining() || state.errHand.curDel == 0 {
			if state.errHand.ignoreErrParser {
				Debugf("HandleWitness - return -> %s: curDel=%d, ignoreErrParser=%t", state.mode, state.errHand.curDel, state.errHand.ignoreErrParser)
				return state, zero
			}
			state, output = parse.It(state)
			if !state.Failed() {
				Debugf("HandleWitness - SUCCESS - %s: curDel=%d, ignoreErrParser=%t", state.mode, state.errHand.curDel, state.errHand.ignoreErrParser)
				return state, output // first parser succeeded, now try the rest
			}
		} else { // speed up since we don't get further anyway
			state.errHand.curDel = state.maxDel
		}
		state.mode = ParsingModeRewind
		Debugf("HandleWitness - One More Round - %s: curDel=%d, ignoreErrParser=%t", state.mode, state.errHand.curDel, state.errHand.ignoreErrParser)
	}
}

// ============================================================================
// Recoverers
//

// DefaultRecoverer shouldn't be used outside of this package.
// Please use pcb.BasicRecovererFunc instead.
func DefaultRecoverer[Output any](parse Parser[Output]) Recoverer {
	return DefaultRecovererFunc(parse.It)
}

// DefaultRecovererFunc is the heart of the DefaultRecoverer and shouldn't be used
// outside of this package either.
// Please use pcb.BasicRecovererFunc instead.
func DefaultRecovererFunc[Output any](parse func(State) (State, Output)) Recoverer {
	return func(state State) int {
		curState := state
		for curState.BytesRemaining() > 0 {
			newState, _ := parse(curState)
			if !newState.Failed() {
				return state.ByteCount(curState) // return the bytes up to the successful position
			}
			curState = curState.Delete(1)
		}
		return -1 // absolut worst case! :(
	}
}

// CachingRecoverer should only be used in places where the Recoverer
// will be used multiple times with the exact same input position.
// The SaveSpot parser is such a case.
func CachingRecoverer(recoverer Recoverer) Recoverer {
	id := cachingRecovererIDs.Add(1)

	return func(state State) int {
		waste, ok := state.cachedRecovererWaste(id)
		if !ok {
			waste = recoverer(state)
			state.cacheRecovererWaste(id, waste)
		}
		return waste
	}
}

type CombiningRecoverer struct {
	recoverers []Recoverer
	lastIdx    int
	id         uint64
}

// NewCombiningRecoverer recovers by calling all sub-recoverers and returning
// the minimal waste.
// The index of the best Recoverer is stored in the cache.
// If `doCache` is false then no caching is performed.
func NewCombiningRecoverer(doCache bool, recoverers ...Recoverer) CombiningRecoverer {
	id := uint64(0)
	if doCache {
		id = combiningRecovererIDs.Add(1)
	}
	return CombiningRecoverer{
		recoverers: recoverers,
		lastIdx:    -1,
		id:         id,
	}
}

func (crc CombiningRecoverer) Recover(state State) int {
	if crc.id > 0 {
		waste, idx, ok := state.cachedRecovererWasteIdx(crc.id)
		if ok {
			crc.lastIdx = idx
			return waste
		}
	}

	waste := -1
	idx := -1
	for i, recoverer := range crc.recoverers {
		if recoverer == nil {
			continue
		}
		w := recoverer(state)
		switch {
		case w == -1: // ignore
		case w == 0: // it won't get better than this
			waste = 0
			idx = i
			break
		case waste < 0 || w < waste:
			waste = w
			idx = i
		}
	}
	crc.lastIdx = idx
	if crc.id > 0 {
		state.cacheRecovererWasteIdx(crc.id, waste, idx)
	}
	return waste
}

func (crc CombiningRecoverer) LastIndex() int {
	return crc.lastIdx
}

func (crc CombiningRecoverer) CachedIndex(state State) (waste, idx int, ok bool) {
	waste, idx, ok = state.cachedRecovererWasteIdx(crc.id)
	if !ok {
		return 0, -1, false
	}
	return waste, idx, true
}

// ============================================================================
// Deleters
//

// DefaultBinaryDeleter shouldn't be used outside of this package.
// Please use pcb.ByteDeleter instead.
func DefaultBinaryDeleter(state State, count int) State {
	if count <= 0 { // don't delete at all
		return state
	}
	return state.MoveBy(count)
}

// DefaultTextDeleter shouldn't be used outside of this package.
// Please use pcb.RuneDeleter instead.
func DefaultTextDeleter(state State, count int) State {
	if count <= 0 { // don't delete at all
		return state
	}
	byteCount, j := 0, 0
	for i := range state.CurrentString() {
		byteCount += i
		j++
		if j >= count {
			return state.MoveBy(byteCount)
		}
	}
	return state.MoveBy(byteCount)
}

// ============================================================================
// Error Reporting
//

func singleErrorMsg(pcbErr pcbError, binary bool) string {
	fullMsg := strings.Builder{}
	fullMsg.WriteString(pcbErr.text)
	if binary {
		fullMsg.WriteString(formatBinaryLine(pcbErr.line, pcbErr.col, pcbErr.srcLine))
	} else {
		fullMsg.WriteString(formatSrcLine(pcbErr.line, pcbErr.col, pcbErr.srcLine))
	}

	return fullMsg.String()
}

func formatBinaryLine(line, col int, srcLine string) string {
	start := line
	text := hex.Dump([]byte(srcLine))
	text = text[10:] // remove wrong offset and spaces

	m1 := col * 3
	if col >= 8 {
		m1++
	}
	// first hex + space + second hex + space + bar + col
	m2 := 8*3 + 1 + 8*3 + 1 + 1 + col
	return fmt.Sprintf(":\n %08x  %s%c%s%c%s",
		// offset, first hex, marker, last hex + ASCII, marker, last ASCII
		start, text[:m1], errorMarker, text[m1:m2], errorMarker, text[m2:len(text)-1])
}

func formatSrcLine(line, col int, srcLine string) string {
	result := strings.Builder{}
	lineStart := srcLine[:col]
	srcLine = srcLine[col:]
	result.WriteString(lastNRunes(lineStart, 10))
	result.WriteRune(errorMarker)
	result.WriteString(firstNRunes(srcLine, 20))
	return fmt.Sprintf(` [%d:%d] %s`,
		line, utf8.RuneCountInString(lineStart)+1, result.String()) // columns for the user start at 1
}
func firstNRunes(s string, n int) string {
	l := len(s)
	if n >= l {
		return s
	}
	i := 0
	j := 0
	for ; i < n && j < l; i++ { // i counts runes and j bytes
		_, size := utf8.DecodeRuneInString(s[j:])
		j += size
	}
	return s[:j]
}
func lastNRunes(s string, n int) string {
	l := len(s)
	if n >= l {
		return s
	}
	i := 0
	j := l
	for ; i < n && j > 0; i++ { // i counts runes and j bytes
		_, size := utf8.DecodeLastRuneInString(s[:j])
		j -= size
	}
	return s[j:]
}
