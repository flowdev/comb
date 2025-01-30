package pcb

import (
	"bytes"
	"github.com/oleiade/gomme"
	"reflect"
	"strings"
)

// Forbidden is the Recoverer for parsers that MUST NOT be used to recover at all.
// These are all parsers that are happy to consume the empty input and
// all look ahead parsers.
func Forbidden() gomme.Recoverer {
	return func(_ gomme.State) int {
		return gomme.RecoverWasteNever // don't recover at all with this parser
	}
}

// IndexOf searches until it finds the stop token in the input.
// If found the Recoverer returns the number of bytes up to the stop.
// If the token could not be found, the recoverer returns gomme.RecoverWasteTooMuch.
// This function panics during the construction phase if `stop` is empty.
func IndexOf[S gomme.Separator](stop S) gomme.Recoverer {
	// This IS type safe because of the `Separator` constraint!
	// Performance doesn't matter either because this is done during the
	// construction phase.
	switch v := reflect.ValueOf(stop); v.Kind() {
	case reflect.Uint8:
		xstop := interface{}(stop).(byte)
		return func(state gomme.State) int {
			waste := bytes.IndexByte(state.CurrentBytes(), xstop)
			if waste < 0 {
				return gomme.RecoverWasteTooMuch
			}
			return waste
		}
	case reflect.Int32:
		rstop := interface{}(stop).(rune)
		return func(state gomme.State) int {
			waste := strings.IndexRune(state.CurrentString(), rstop)
			if waste < 0 {
				return gomme.RecoverWasteTooMuch
			}
			return waste
		}
	case reflect.String:
		sstop := interface{}(stop).(string)
		if len(sstop) == 0 {
			panic("stop is empty")
		}
		return func(state gomme.State) int {
			waste := strings.Index(state.CurrentString(), sstop)
			if waste < 0 {
				return gomme.RecoverWasteTooMuch
			}
			return waste
		}
	case reflect.Slice:
		bstop := interface{}(stop).([]byte)
		if len(bstop) == 0 {
			panic("stop is empty")
		}
		return func(state gomme.State) int {
			waste := bytes.Index(state.CurrentBytes(), bstop)
			if waste < 0 {
				return gomme.RecoverWasteTooMuch
			}
			return waste
		}
	default:
		return nil // can never happen because of the `Separator` constraint!
	}
}

// IndexOfAny searches until it finds a stop token in the input.
// If found the recoverer returns the number of bytes up to the stop.
// If no stop token could be found, the recoverer returns gomme.RecoverWasteTooMuch.
//
// NOTE:
//   - If any of the `stops` is empty it returns 0.
//   - If no stops are provided then this function panics during
//     the construction phase.
func IndexOfAny[S gomme.Separator](stops ...S) gomme.Recoverer {
	const (
		modeByte = iota
		modeRune
		modeBytes
		modeString
	)
	var mode int
	n := len(stops)

	if n == 0 {
		panic("no stops provided")
	}

	// This IS type safe because of the `Separator` constraint!
	// Performance doesn't matter either because this is done during the
	// construction phase.
	switch v := reflect.ValueOf(stops[0]); v.Kind() {
	case reflect.Uint8:
		mode = modeByte
	case reflect.Int32:
		mode = modeRune
	case reflect.String:
		mode = modeString
	case reflect.Slice:
		mode = modeBytes
	default:
		// can never happen because of the `Separator` constraint!
	}

	indexOfOneOfByte := func(state gomme.State) int {
		input := state.CurrentBytes()
		xstops := interface{}(stops).([]byte)
		pos := gomme.RecoverWasteTooMuch
		for i := 0; i < n; i++ {
			switch j := bytes.IndexByte(input, xstops[i]); j {
			case -1: // ignore
			case 0: // it won't get better than this
				return 0
			default:
				if pos < 0 || j < pos {
					pos = j
				}
			}
		}
		return pos
	}
	indexOfOneOfRune := func(state gomme.State) int {
		return strings.IndexAny(state.CurrentString(), string(interface{}(stops).([]rune)))
	}
	indexOfOneOfBytes := func(state gomme.State) int {
		input := state.CurrentBytes()
		bstops := interface{}(stops).([][]byte)
		pos := gomme.RecoverWasteTooMuch
		for i := 0; i < n; i++ {
			switch j := bytes.Index(input, bstops[i]); j {
			case -1: // ignore
			case 0: // it won't get better than this
				return 0
			default:
				if pos < 0 || j < pos {
					pos = j
				}
			}
		}
		return pos
	}
	indexOfOneOfString := func(state gomme.State) int {
		input := state.CurrentString()
		sstops := interface{}(stops).([]string)
		pos := gomme.RecoverWasteTooMuch
		for i := 0; i < n; i++ {
			switch j := strings.Index(input, sstops[i]); j {
			case -1: // ignore
			case 0: // it won't get better than this
				return 0
			default:
				if pos < 0 || j < pos {
					pos = j
				}
			}
		}
		return pos
	}
	switch mode {
	case modeByte:
		return indexOfOneOfByte
	case modeRune:
		return indexOfOneOfRune
	case modeBytes:
		return indexOfOneOfBytes
	default:
		return indexOfOneOfString
	}
}
