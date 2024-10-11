package gomme

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DefaultRecoverer shouldn't be used outside of this package.
// Please use pcb.BasicRecoverer instead.
func DefaultRecoverer[Output any](parse Parser[Output]) Recoverer {
	return func(state State) int {
		curState := state
		for curState.BytesRemaining() > 0 {
			newState, _ := parse.It(curState)
			if !newState.Failed() {
				return state.ByteCount(curState) // return the bytes up to the successful position
			}
			curState = curState.Delete(1)
		}
		return -1 // absolut worst case! :(
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
