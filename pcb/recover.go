package pcb

import (
	"bytes"
	"fmt"
	"github.com/oleiade/gomme"
	"math"
	"reflect"
	"strings"
)

// SkipTo parses until it finds the stop token in the input.
// If found the recoverer moves up to the stop token but doesn't consume it.
// If the token could not be found, the recoverer returns an error result and -1.
// This function panics if `stop` is empty.
func SkipTo[S gomme.Separator](stop S) gomme.Recoverer {
	sstop := ""
	bstop := []byte{}

	// This IS type safe because of the `Separator` constraint!
	// Performance doesn't matter either because this is done only at set up.
	switch v := reflect.ValueOf(stop); v.Kind() {
	case reflect.Int32:
		sstop = string(rune(v.Int()))
	case reflect.String:
		sstop = v.String()
	case reflect.Uint8:
		bstop = []byte{byte(v.Uint())}
	case reflect.Slice:
		bstop = v.Bytes()
	default:
		// can never happen because of the `Separator` constraint!
	}
	n := len(sstop) + len(bstop) // one of the two should be filled
	if n == 0 {
		panic("stop is empty")
	}

	recovery := func(state gomme.State) (gomme.State, int) {
		i := 0
		if len(sstop) > 0 {
			input := state.CurrentString()
			i = strings.Index(input, sstop)
		} else {
			input := state.CurrentBytes()
			i = bytes.Index(input, bstop)
		}
		if i == -1 {
			return state.NewError(fmt.Sprintf("... %q", stop), state, 0), -1
		}

		newState := state.MoveBy(uint(i))
		return newState, i
	}

	return recovery
}

// SkipToOneOf parses until it finds a stop token in the input.
// If found the recoverer moves up to the stop token but doesn't consume it.
// If no stop token could be found, the recoverer returns an error result and -1.
// This function panics if any of the `stops` is empty.
func SkipToOneOf[S gomme.Separator](stops ...S) gomme.Recoverer {
	n := len(stops)
	sstop := make([]string, n)
	bstop := make([][]byte, n)
	lens := make([]int, n)

	for i, stop := range stops {
		// This IS type safe because of the `Separator` constraint!
		// Performance doesn't matter either because this is done only at set up.
		switch v := reflect.ValueOf(stop); v.Kind() {
		case reflect.Int32:
			sstop[i] = string(rune(v.Int()))
		case reflect.String:
			sstop[i] = v.String()
		case reflect.Uint8:
			bstop[i] = []byte{byte(v.Uint())}
		case reflect.Slice:
			bstop[i] = v.Bytes()
		default:
			// can never happen because of the `Separator` constraint!
		}
		lens[i] = len(sstop[i]) + len(bstop[i]) // one of the two MUST be filled
		if lens[i] == 0 {
			panic(fmt.Sprintf("stop with index %d is empty", i))
		}
	}

	recovery := func(state gomme.State) (gomme.State, int) {
		sinput := state.CurrentString()
		binput := state.CurrentBytes()
		pos := math.MaxInt
		for i := 0; i < n; i++ {
			p := 0
			if len(sstop) > 0 {
				p = strings.Index(sinput, sstop[i])
			} else {
				p = bytes.Index(binput, bstop[i])
			}
			if p >= 0 {
				pos = min(pos, p)
			}
		}
		if pos == math.MaxInt {
			return state.NewError(fmt.Sprintf("... one of %q", stops), state, 0), -1
		}

		newState := state.MoveBy(uint(pos))
		return newState, pos
	}

	return recovery
}
