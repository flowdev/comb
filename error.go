package gomme

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

const errorMarker = 0x25B6 // easy to spot marker (▶) for exact error position

// ParserError is an error message from the parser.
// It consists of the text itself and the position in the input where it happened.
type ParserError struct {
	text      string // the error message from the parser
	pos       int    // pos is the byte index in the input (state.input.pos)
	line, col int    // col is the 0-based byte index within srcLine; convert to 1-based rune index for user
	srcLine   string // line of the source code containing the error or bytes around the error in binary case
	binary    bool   // are we in binary or text mode?
	parserID  int32  // ID of the parser reporting the error (only set for syntax errors)
}

func (e *ParserError) Error() string {
	return singleErrorMsg(*e)
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
func DefaultRecovererFunc[Output any](parse func(State) (State, Output, *ParserError)) func(State) int {
	return func(state State) int {
		curState := state
		for curState.BytesRemaining() > 0 {
			_, _, err := parse(curState)
			if err == nil {
				return state.ByteCount(curState) // return the bytes up to the successful position
			}
			curState = curState.Delete1()
		}
		return -1 // absolut worst case! :(
	}
}

// ============================================================================
// Error Reporting
//

func singleErrorMsg(pcbErr ParserError) string {
	fullMsg := strings.Builder{}
	fullMsg.WriteString(pcbErr.text)
	if pcbErr.binary {
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
