package gomme

import (
	"cmp"
	"slices"
	"strings"
)

// ============================================================================
// This File contains only the State and cache data structures and all of their
// methods.
// ============================================================================

type cachedWaste struct {
	pos   int // position in the input
	waste int // waste of the recoverer
}

type cachedWasteIdx struct {
	pos   int // position in the input
	waste int // waste of the recoverer
	idx   int // index of the best sub-recoverer
}

// State represents the current state of a parser.
type State struct {
	mode                   ParsingMode // one of: happy, error, handle, record, choose, play
	input                  Input
	pointOfNoReturn        int        // mark set by SignalNoWayBack/NoWayBack parser
	newError               *pcbError  // error that hasn't been handled yet
	maxDel                 int        // maximum number of tokens that should be deleted for error recovery
	deleter                Deleter    // used to get back on track in error recovery
	errHand                errHand    // everything for handling one error
	oldErrors              []pcbError // errors that are or have been handled
	recovererWasteCache    map[uint64][]cachedWaste
	recovererWasteIdxCache map[uint64][]cachedWasteIdx
}

// ============================================================================
// Handle Input
//

func (st State) AtEnd() bool {
	return st.input.pos >= len(st.input.bytes)
}

func (st State) BytesRemaining() int {
	return len(st.input.bytes) - st.input.pos
}

func (st State) CurrentString() string {
	return string(st.input.bytes[st.input.pos:])
}

func (st State) CurrentBytes() []byte {
	return st.input.bytes[st.input.pos:]
}

func (st State) StringTo(remaining State) string {
	return string(st.BytesTo(remaining))
}

func (st State) BytesTo(remaining State) []byte {
	if remaining.input.pos < st.input.pos {
		return []byte{}
	}
	if remaining.input.pos > len(st.input.bytes) {
		return st.input.bytes[st.input.pos:]
	}
	return st.input.bytes[st.input.pos:remaining.input.pos]
}

func (st State) ByteCount(remaining State) int {
	if remaining.input.pos < st.input.pos {
		return 0 // we never go back so we don't give negative count back
	}
	n := len(st.input.bytes)
	if remaining.input.pos > n {
		return n - st.input.pos
	}
	return remaining.input.pos - st.input.pos
}

func (st State) MoveBy(countBytes int) State {
	if countBytes < 0 {
		countBytes = 0
	}

	pos := st.input.pos
	n := min(len(st.input.bytes), pos+countBytes)
	st.input.pos = n

	moveText := string(st.input.bytes[pos:n])
	lastNlPos := strings.LastIndexByte(moveText, '\n') // this is Unicode safe!!!
	if lastNlPos >= 0 {
		st.input.prevNl += lastNlPos + 1 // this works even if '\n' wasn't found at all
		st.input.line += strings.Count(moveText, "\n")
	}

	return st
}

func (st State) Moved(other State) bool {
	return st.input.pos != other.input.pos
}

func (st State) Delete(countToken int) State {
	return st.deleter(st, countToken)
}

// ============================================================================
// Caching
//

// cacheRecovererWaste remembers the `waste` at the current input position
// for the CachingRecoverer with ID `id`.
func (st State) cacheRecovererWaste(id uint64, waste int) {
	cache, ok := st.recovererWasteCache[id]
	if !ok {
		cache = make([]cachedWaste, 0, st.maxDel+1)
		st.recovererWasteCache[id] = cache
	}

	if len(cache) < st.maxDel+1 {
		st.recovererWasteCache[id] = append(cache, cachedWaste{pos: st.input.pos, waste: waste})
		return
	}

	idx := MinFuncIdx(cache, func(a, b cachedWaste) int { // idx will never be -1
		return cmp.Compare(a.pos, b.pos)
	})
	cache[idx] = cachedWaste{pos: st.input.pos, waste: waste}
}

// cachedRecovererWaste returns the saved waste for the current
// input position and CachingRecoverer ID `id` or (-1, false) if not found.
func (st State) cachedRecovererWaste(id uint64) (waste int, ok bool) {
	var cache []cachedWaste
	if cache, ok = st.recovererWasteCache[id]; !ok {
		return -1, false
	}

	idx := slices.IndexFunc(cache, func(wasteData cachedWaste) bool {
		return wasteData.pos == st.input.pos
	})

	if idx < 0 {
		return -1, false
	}
	return cache[idx].waste, true
}

// cacheRecovererWasteIdx remembers the `waste` and index at the
// current input position for the CombiningRecoverer with ID `crID`.
func (st State) cacheRecovererWasteIdx(crID uint64, waste, idx int) {
	cache, ok := st.recovererWasteIdxCache[crID]
	if !ok {
		cache = make([]cachedWasteIdx, 0, st.maxDel+1)
		st.recovererWasteIdxCache[crID] = cache
	}

	if len(cache) < st.maxDel+1 {
		st.recovererWasteIdxCache[crID] = append(cache, cachedWasteIdx{pos: st.input.pos, waste: waste})
		return
	}

	i := MinFuncIdx(cache, func(a, b cachedWasteIdx) int { // idx will never be -1
		return cmp.Compare(a.pos, b.pos)
	})
	cache[i] = cachedWasteIdx{pos: st.input.pos, waste: waste, idx: idx}
}

// cachedRecovererWasteIdx returns the saved waste and index for the current
// input position and CombiningRecoverer ID or (-1, -1, false) if not found.
func (st State) cachedRecovererWasteIdx(crID uint64) (waste, idx int, ok bool) {
	var cache []cachedWasteIdx
	if cache, ok = st.recovererWasteIdxCache[crID]; !ok {
		return -1, -1, false
	}

	i := slices.IndexFunc(cache, func(wasteData cachedWasteIdx) bool {
		return wasteData.pos == st.input.pos
	})

	if i < 0 {
		return -1, -1, false
	}

	wasteData := cache[i]
	return wasteData.waste, wasteData.idx, true
}

// ============================================================================
// Handle success and failure
//

// Success return the State with NoWayBack saved from
// the subState.
func (st State) Success(subState State) State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, subState.pointOfNoReturn)
	return st
}

// Failure returns the State with the error, pointOfNoReturn and mode kept from
// the subState.
func (st State) Failure(subState State) State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, subState.pointOfNoReturn)
	st.mode = subState.mode

	if subState.newError != nil { // should be true
		st.newError = subState.newError
	}

	return st
}

// NewError sets a syntax error with the message in this state at the current position.
// For syntax errors `expected ` is prepended to the message and the usual
// position and source line including marker are appended.
func (st State) NewError(message string) State {
	line, col, srcLine := st.where(st.input.pos)
	newErr := pcbError{
		text: "expected " + message,
		pos:  st.input.pos, line: line, col: col,
		srcLine: srcLine,
	}

	switch st.mode {
	case ParsingModeHappy:
		st.newError = &newErr
		st.mode = ParsingModeError
	case ParsingModeError:
		// should NOT happen but keep error furthest in the input
		//if st.newError == nil || st.newError.pos < st.input.pos {
		//	st.newError = newErr
		//}

		// programming error
		newErr.text = "programming error: State.NewError called in mode `error` (while backtracking)"
		st.oldErrors = append(st.oldErrors, newErr)
	case ParsingModeHandle:
		if st.handlingNewError(&newErr) {
			st.mode = ParsingModeRecord
			st.newError = nil
		} else {
			newErr.text = "programming error: State.NewError called in mode `handle` with other error"
			st.oldErrors = append(st.oldErrors, newErr)
		}
	case ParsingModeRecord, ParsingModeCollect:
		// ignore error (we simulate the happy path) (should not happen)
	case ParsingModeChoose, ParsingModePlay:
		st.newError = &newErr
	}
	return st
}

// NewSemanticError sets a semantic error with the messages in this state at the
// current position.
// For semantic errors `expected ` is NOT prepended to the message but the usual
// position and source line including marker are appended.
func (st State) NewSemanticError(message string) State {
	line, col, srcLine := st.where(st.input.pos)
	err := pcbError{
		text: message,
		pos:  st.input.pos, line: line, col: col,
		srcLine: srcLine,
	}

	st.oldErrors = append(st.oldErrors, err)
	return st
}

// Failed returns whether this state is in a failed state or not.
func (st State) Failed() bool {
	return st.newError != nil
}

// ============================================================================
// Produce error messages and give them back
//

// CurrentSourceLine returns the source line corresponding to the current position
// including [line:column] at the start and a marker at the exact error position.
// This should be used for reporting errors that are detected later.
func (st State) CurrentSourceLine() string {
	return formatSrcLine(st.where(st.input.pos))
}

func (st State) where(pos int) (line, col int, srcLine string) {
	if len(st.input.bytes) == 0 {
		return 1, 0, ""
	}
	if pos > st.input.prevNl { // pos is ahead of prevNL => search forward
		return st.whereForward(pos, st.input.line, st.input.prevNl)
	} else if pos <= st.input.prevNl-pos { // pos is too far back => search from start
		return st.whereForward(pos, 1, -1)
	} else { // pos is just a little back => search backward
		return st.whereBackward(pos, st.input.line, st.input.prevNl)
	}
}
func (st State) whereForward(pos, lineNum, prevNl int) (line, col int, srcLine string) {
	text := string(st.input.bytes)
	var nextNl int // Position of next newline or end

	for {
		nextNl = strings.IndexByte(text[prevNl+1:], '\n')
		if nextNl < 0 {
			nextNl = len(text)
		} else {
			nextNl += prevNl + 1
		}

		stop := false
		line, col, srcLine, stop = st.tryWhere(prevNl, pos, nextNl, lineNum)
		if stop {
			return line, col, srcLine
		}
		prevNl = nextNl
		lineNum++
	}
}
func (st State) whereBackward(pos, lineNum, nextNl int) (line, col int, srcLine string) {
	text := string(st.input.bytes)
	var prevNl int // Line start (position of preceding newline)

	for {
		prevNl = strings.LastIndexByte(text[0:nextNl], '\n')
		lineNum--

		stop := false
		line, col, srcLine, stop = st.tryWhere(prevNl, pos, nextNl, lineNum)
		if stop {
			return line, col, srcLine
		}
		nextNl = prevNl
	}
}
func (st State) tryWhere(prevNl int, pos int, nextNl int, lineNum int) (line, col int, srcLine string, stop bool) {
	if prevNl < pos && pos <= nextNl {
		return lineNum, pos - prevNl - 1, string(st.input.bytes[prevNl+1 : nextNl]), true
	}
	return 1, 0, "", false
}

// Error returns a human readable error string.
func (st State) Error() string {
	slices.SortFunc(st.oldErrors, func(a, b pcbError) int { // always keep them sorted
		return cmp.Compare(a.pos, b.pos)
	})

	fullMsg := strings.Builder{}
	for _, pcbErr := range st.oldErrors {
		fullMsg.WriteString(singleErrorMsg(pcbErr))
		fullMsg.WriteRune('\n')
	}
	if st.newError != nil {
		fullMsg.WriteString(singleErrorMsg(*st.newError))
		fullMsg.WriteRune('\n')
	}

	return fullMsg.String()
}

// handlingNewError is currently only comparing the error position.
func (st State) handlingNewError(newErr *pcbError) bool {
	if st.errHand.err == nil || newErr == nil {
		return false
	}
	return st.errHand.err.pos == newErr.pos
}

// SignalNoWayBack sets a point of no return mark at the current position.
func (st State) SignalNoWayBack() State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, st.input.pos)
	return st
}

// NoWayBack is true iff we crossed a point of no return.
func (st State) NoWayBack() bool {
	return st.pointOfNoReturn >= st.input.pos
}
