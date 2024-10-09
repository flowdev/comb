package pcb

import (
	"bytes"
	"fmt"
	"github.com/oleiade/gomme"
	"math"
	"reflect"
	"strings"
)

// Forbidden is the Recoverer for parsers that MUST NOT be used to recover at all.
// These are all parsers that are happy to consume the empty input and
// all look ahead parsers.
// The returned Recoverer panics if used.
func Forbidden(name string) gomme.Recoverer {
	return func(state gomme.State) int {
		panic("must not use parser `" + name + "` accepting empty input for recovering from an error")
	}
}

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
	if len(sstop)+len(bstop) == 0 { // one of the two MUST be filled
		panic("stop is empty")
	}

	recovery := func(state gomme.State) int {
		if len(sstop) > 0 {
			return strings.Index(state.CurrentString(), sstop)
		} else {
			return bytes.Index(state.CurrentBytes(), bstop)
		}
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
		if len(sstop[i])+len(bstop[i]) == 0 { // one of the two MUST be filled
			panic(fmt.Sprintf("stop with index %d is empty", i))
		}
	}

	recovery := func(state gomme.State) int {
		sinput := state.CurrentString()
		binput := state.CurrentBytes()
		pos := math.MaxInt
		for i := 0; i < n; i++ {
			j := 0
			if len(sstop[i]) > 0 {
				j = strings.Index(sinput, sstop[i])
			} else {
				j = bytes.Index(binput, bstop[i])
			}
			if j >= 0 {
				pos = min(pos, j)
			}
		}
		if pos == math.MaxInt {
			return -1
		}

		return pos
	}

	return recovery
}
