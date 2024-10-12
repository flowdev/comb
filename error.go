package gomme

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

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
	err *pcbError // error that is currently handled
}

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

var cachingRecovererIDs = &atomic.Uint64{}

// CachingRecoverer should only be used in places where the Recoverer
// will be used multiple times with the exact same input position.
// The NoWayBack and Refuge parsers are such cases.
func CachingRecoverer(recoverer Recoverer) Recoverer {
	id := cachingRecovererIDs.Add(1)

	return func(state State) int {
		cachedWaste, ok := state.cachedRecovererWaste(id)
		if !ok {
			cachedWaste = recoverer(state)
			state.cacheRecovererWaste(id, cachedWaste)
		}
		return cachedWaste
	}
}

// DefaultBinaryDeleter shouldn't be used outside of this package.
// Please use pcb.ByteDeleter instead.
func DefaultBinaryDeleter(state State, count int) State {
	return state.MoveBy(count)
}

// DefaultTextDeleter shouldn't be used outside of this package.
// Please use pcb.RuneTypeChangeDeleter instead.
func DefaultTextDeleter(state State, count int) State {
	found := 0
	oldTyp := rune(0)

	byteCount := strings.IndexFunc(state.CurrentString(), func(r rune) bool {
		var typ, paren rune

		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_':
			typ = 'a'
		case unicode.IsSpace(r):
			typ = ' '
		case slices.Contains([]rune{'(', '[', '{', '}', ']', ')'}, r):
			typ = '('
		case slices.Contains([]rune{
			'+', '-', '*', '/', '%', '^', '=', ':', '<', '>', '~',
			'|', '\\', ';', '.', ',', '"', '`', '\'',
		}, r):
			typ = '+'
		default:
			typ = utf8.RuneError
		}

		if typ != oldTyp {
			if typ != ' ' && oldTyp != 0 {
				found++
			}
			if typ == '(' && oldTyp == '(' && r != paren {
				found++
			}
			oldTyp = typ
			paren = r // works just fine even if r isn't a parenthesis at all and saves an if
		}
		return found == count
	})

	if byteCount < 0 {
		return state.MoveBy(state.BytesRemaining())
	}
	return state.MoveBy(byteCount)
}

// ============================================================================
// Error Handling
//

func HandleAllErrors[Output any](state State, parse Parser[Output]) (State, Output) {
	var output Output
	var newState State

	return newState, output
}

func HandleCurrentError[Output any](state State, parse Parser[Output]) (State, Output) {
	if !state.handlingNewError(state.newError) {
		return state, ZeroOf[Output]()
	}

	return state, ZeroOf[Output]()
}

func singleErrorMsg(pcbErr pcbError) string {
	fullMsg := strings.Builder{}
	fullMsg.WriteString(pcbErr.text)
	fullMsg.WriteString(formatSrcLine(pcbErr.line, pcbErr.col, pcbErr.srcLine))

	return fullMsg.String()
}

func formatSrcLine(line, col int, srcLine string) string {
	result := strings.Builder{}
	lineStart := srcLine[:col]
	result.WriteString(lineStart)
	result.WriteRune(0x25B6) // easy to spot marker (▶) for exact error position
	result.WriteString(srcLine[col:])
	return fmt.Sprintf(" [%d:%d] %q",
		line, utf8.RuneCountInString(lineStart)+1, result.String()) // columns for the user start at 1
}

func pcbErrorsToGoErrors(pcbErrors []pcbError) error {
	if len(pcbErrors) == 0 {
		return nil
	}

	goErrors := make([]error, len(pcbErrors))
	for i, pe := range pcbErrors {
		goErrors[i] = errors.New(singleErrorMsg(pe))
	}

	return errors.Join(goErrors...)
}
