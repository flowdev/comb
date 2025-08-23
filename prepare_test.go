package comb

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPreparedParserParseAll(t *testing.T) {
	runePlusRune := func(out1 rune, out2 rune) (string, error) {
		return string([]rune{out1, out2}), nil
	}
	stringPlusRune := func(out1 string, out2 rune) (string, error) {
		return out1 + string([]rune{out2}), nil
	}

	tests := []struct {
		name       string
		input      string
		badParser  bool
		wantOutput interface{}
		wantErrors int
	}{
		{
			name:       "goodInputBadParser",
			input:      "abc",
			badParser:  true,
			wantOutput: "abc",
			wantErrors: 0,
		}, {
			name:       "goodInputGoodParser",
			input:      "abc",
			badParser:  false,
			wantOutput: "abc",
			wantErrors: 0,
		}, {
			name:       "emptyBadParser",
			input:      "",
			badParser:  true,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "emptyGoodParser",
			input:      "",
			badParser:  false,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "lastCharMissingBadParser",
			input:      "ab",
			badParser:  true,
			wantOutput: "ab�",
			wantErrors: 1,
		}, {
			name:       "lastCharMissingGoodParser",
			input:      "ab",
			badParser:  false,
			wantOutput: "ab�",
			wantErrors: 1,
		}, {
			name:       "middleCharMissingBadParser",
			input:      "ac",
			badParser:  true,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "middleCharMissingGoodParser",
			input:      "ac",
			badParser:  false,
			wantOutput: "a\ufffdc",
			wantErrors: 1,
		}, {
			name:       "firstCharMissingBadParser",
			input:      "bc",
			badParser:  true,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "firstCharMissingGoodParser",
			input:      "bc",
			badParser:  false,
			wantOutput: "\ufffdbc",
			wantErrors: 1,
		}, {
			name:       "firstCharOffBadParser",
			input:      "1abc",
			badParser:  true,
			wantOutput: "abc",
			wantErrors: 1,
		}, {
			name:       "firstCharOffGoodParser",
			input:      "1abc",
			badParser:  false,
			wantOutput: "abc",
			wantErrors: 1,
		}, {
			name:       "secondCharOffBadParser",
			input:      "a1bc",
			badParser:  true,
			wantOutput: "abc",
			wantErrors: 1,
		}, {
			name:       "secondCharOffGoodParser",
			input:      "a1bc",
			badParser:  false,
			wantOutput: "abc",
			wantErrors: 1,
		}, {
			name:       "thirdCharOffBadParser",
			input:      "ab1c",
			badParser:  true,
			wantOutput: "abc",
			wantErrors: 1,
		}, {
			name:       "thirdCharOffGoodParser",
			input:      "ab1c",
			badParser:  false,
			wantOutput: "abc",
			wantErrors: 1,
		}, {
			name:       "firstAndLastCharOffBadParser",
			input:      "1ab2c",
			badParser:  true,
			wantOutput: "abc",
			wantErrors: 2,
		}, {
			name:       "firstAndLastCharOffGoodParser",
			input:      "1ab2c",
			badParser:  false,
			wantOutput: "abc",
			wantErrors: 2,
		}, {
			name:       "firstCharOffMiddleCharMissingBadParser",
			input:      "1ac",
			badParser:  true,
			wantOutput: "",
			wantErrors: 2,
		}, {
			name:       "firstCharOffMiddleCharMissingGoodParser",
			input:      "1ac",
			badParser:  false,
			wantOutput: "a�c",
			wantErrors: 2,
		}, {
			name:       "allCharsOffBadParser",
			input:      "1a2b3c",
			badParser:  true,
			wantOutput: "abc",
			wantErrors: 3,
		}, {
			name:       "allCharsOffGoodParser",
			input:      "1a2b3c",
			badParser:  false,
			wantOutput: "abc",
			wantErrors: 3,
		}, {
			name:       "firstCharMissingLastCharOffBadParser",
			input:      "b1c",
			badParser:  true,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "firstCharMissingLastCharOffGoodParser",
			input:      "b1c",
			badParser:  false,
			wantOutput: "\ufffdbc",
			wantErrors: 2,
		}, {
			name:       "firstCharOffMiddleCharMissingBadParser",
			input:      "1ac",
			badParser:  true,
			wantOutput: "",
			wantErrors: 2,
		}, {
			name:       "firstCharOffMiddleCharMissingGoodParser",
			input:      "1ac",
			badParser:  false,
			wantOutput: "a�c",
			wantErrors: 2,
		}, {
			name:       "onlyFirstCharBadParser",
			input:      "a",
			badParser:  true,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "onlyFirstCharGoodParser",
			input:      "a",
			badParser:  false,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "onlyMiddleCharBadParser",
			input:      "b",
			badParser:  true,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "onlyMiddleCharGoodParser",
			input:      "b",
			badParser:  false,
			wantOutput: "",
			wantErrors: 2,
		}, {
			name:       "onlyLastCharBadParser",
			input:      "c",
			badParser:  true,
			wantOutput: "",
			wantErrors: 1,
		}, {
			name:       "onlyLastCharGoodParser",
			input:      "c",
			badParser:  false,
			wantOutput: "c",
			wantErrors: 1,
		}, {
			name:       "firstCharLastBadParser",
			input:      "bca",
			badParser:  true,
			wantOutput: "",
			wantErrors: 2,
		}, {
			name:       "firstCharLastGoodParser",
			input:      "bca",
			badParser:  false,
			wantOutput: "\ufffdbc",
			wantErrors: 1,
		},
	}
	SetDebug(true)
	for _, tc := range tests {
		tt := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tt.name, func(t *testing.T) {
			var parser Parser[string]
			if tt.badParser {
				parser = Map2(
					Map2(Char('a'), Char('b'), runePlusRune),
					Char('c'),
					stringPlusRune,
				)
			} else {
				parser = Map2(
					Map2(SafeSpot(Char('a')), SafeSpot(Char('b')), runePlusRune),
					SafeSpot(Char('c')),
					stringPlusRune,
				)
			}
			prepp := NewPreparedParser[string](parser) // this calls ParserToAnyParser
			gotOutput, err := prepp.parseAll(NewFromString(tt.input, 10))
			t.Logf("err=%v", err)
			if got, want := len(UnwrapErrors(err)), tt.wantErrors; got != want {
				t.Errorf("err=%v, want=%d", err, want)
			}
			if gotOutput != tt.wantOutput {
				t.Errorf("got output=%q, want=%q", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestBranchParserToAnyParser(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantID     int32
		wantOutput interface{}
		wantError  bool
	}{
		{
			name:       "allGoodBranchParser",
			input:      "ab",
			wantID:     0,
			wantOutput: "ab",
			wantError:  false,
		}, {
			name:       "firstSubparserMissesInput",
			input:      "b",
			wantID:     1,
			wantOutput: "",
			wantError:  true,
		}, {
			name:       "OneByteOff",
			input:      "1ab",
			wantID:     1,
			wantOutput: "",
			wantError:  true,
		}, {
			name:       "secondSubparserMissesInput",
			input:      "a",
			wantID:     2,
			wantOutput: "a\ufffd",
			wantError:  true,
		}, {
			name:       "secondSubparserOneByteOff",
			input:      "a1b",
			wantID:     2,
			wantOutput: "a\ufffd",
			wantError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := Map2(Char('a'), Char('b'), func(out1 rune, out2 rune) (string, error) {
				return string([]rune{out1, out2}), nil
			})
			prepp := NewPreparedParser[string](parser) // this calls ParserToAnyParser
			aParse := prepp.parsers[0]
			_, out, err := aParse.ParseAny(-1, NewFromString(tt.input, 10))
			if got, want := aParse.IsSaveSpot(), false; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			_, gotBranchParser := aParse.(BranchParser)
			if got, want := gotBranchParser, true; got != want {
				t.Errorf("branch parser=%t, want=%t", got, want)
			}
			if got, want := err != nil, tt.wantError; got != want {
				t.Errorf("result.Error=%v, want=%t", got, want)
			}
			if err != nil {
				if got, want := err.parserID, tt.wantID; got != want {
					t.Errorf("error parser ID=%d, want=%d", got, want)
				}
			}
			gotOutput, ok := out.(string)
			if got, want := gotOutput, tt.wantOutput; got != want {
				t.Errorf("output=%v (OK=%t), want=%v", gotOutput, ok, want)
			}

			if got, want := aParse.IsStepRecoverer(), false; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
		})
	}
}

func TestLeafParserToAnyParser(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		safeParser        bool
		wantSaveSpot      bool
		wantStepRecoverer bool
		wantOutput        interface{}
		wantError         bool
		wantWaste         int
	}{
		{
			name:              "allGoodSimple",
			input:             "a",
			safeParser:        false,
			wantSaveSpot:      false,
			wantStepRecoverer: false,
			wantOutput:        'a',
			wantError:         false,
			wantWaste:         0,
		}, {
			name:              "allGoodSaveSpot",
			input:             "a",
			safeParser:        true,
			wantSaveSpot:      true,
			wantStepRecoverer: false,
			wantOutput:        'a',
			wantError:         false,
			wantWaste:         0,
		}, {
			name:              "emptySimple",
			input:             "",
			safeParser:        false,
			wantSaveSpot:      false,
			wantStepRecoverer: false,
			wantOutput:        utf8.RuneError,
			wantError:         true,
			wantWaste:         RecoverWasteTooMuch,
		}, {
			name:              "emptySaveSpot",
			input:             "",
			safeParser:        true,
			wantSaveSpot:      true,
			wantStepRecoverer: false,
			wantOutput:        utf8.RuneError,
			wantError:         true,
			wantWaste:         RecoverWasteTooMuch,
		}, {
			name:              "twoBytesOffSimple",
			input:             "bca",
			safeParser:        false,
			wantSaveSpot:      false,
			wantStepRecoverer: false,
			wantOutput:        utf8.RuneError,
			wantError:         true,
			wantWaste:         2,
		}, {
			name:              "twoBytesOffSaveSpot",
			input:             "bca",
			safeParser:        true,
			wantSaveSpot:      true,
			wantStepRecoverer: false,
			wantOutput:        utf8.RuneError,
			wantError:         true,
			wantWaste:         2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parser Parser[rune]
			if tt.safeParser {
				parser = SafeSpot[rune](Char('a'))
			} else {
				parser = Char('a')
			}

			prepp := NewPreparedParser[rune](parser) // this calls ParserToAnyParser
			aParse := prepp.parsers[0]
			_, out, err := aParse.ParseAny(-1, NewFromString(tt.input, 10))
			if got, want := parser.IsSaveSpot(), tt.wantSaveSpot; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			_, gotBranchParser := parser.(BranchParser)
			if got, want := gotBranchParser, false; got != want {
				t.Errorf("branch parser=%t, want=%t", got, want)
			}
			if got, want := err != nil, tt.wantError; got != want {
				t.Errorf("result.Error=%v, want=%t", got, want)
			}
			if err != nil {
				if got, want := err.parserID, aParse.ID(); got != want {
					t.Errorf("error parser ID=%d, want=%d", got, want)
				}
			}
			gotOutput, ok := out.(rune)
			if gotOutput != tt.wantOutput {
				t.Errorf("output=%v (OK=%t), want=%v", out, ok, tt.wantOutput)
			}

			if got, want := parser.IsStepRecoverer(), tt.wantStepRecoverer; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			if !parser.IsStepRecoverer() {
				waste, _ := parser.Recover(NewFromString(tt.input, 10), nil)
				if got, want := waste, tt.wantWaste; got != want {
					t.Errorf("save spot parser=%d, want=%d", got, want)
				}
			}
		})
	}
}

// ============================================================================
// Map2 Parser
//

type map2data[PO1, PO2 any, MO any] struct {
	p1 Parser[PO1]
	p2 Parser[PO2]
	fn func(PO1, PO2) (MO, error)
	id func() int32
}

func (md *map2data[PO1, PO2, MO]) children() []AnyParser {
	return []AnyParser{md.p1, md.p2}
}
func (md *map2data[PO1, PO2, MO]) parseAfterError(
	childID int32,
	childStartState, childState State,
	childOut interface{},
	childErr *ParserError,
	data interface{},
) (State, MO, *ParserError, interface{}) {
	var zero MO
	var out1 PO1

	Debugf("Map2 - pos=%d; parse after ID %d", childState.CurrentPos(), childID)

	if childID == md.p1.ID() { // on the way up: Fetch
		out1, _ = childOut.(PO1)
	}

	if childErr != nil {
		return childState, zero, childErr, out1
	}

	if childID >= 0 && childID != md.p1.ID() && childID != md.p2.ID() {
		childErr = childState.NewSemanticError("unable to parse after child with unknown ID %d", childID)
		return childState, zero, childErr, out1
	}

	if childID < 0 {
		childStartState = childState
		childState, childOut, childErr = md.p1.ParseAny(md.id(), childStartState)
		if childErr != nil {
			return childState, zero, childErr, childOut
		}
		out1, _ = childOut.(PO1)
	}

	var out2 PO2
	if childID != md.p2.ID() {
		childStartState = childState
		childState, childOut, childErr = md.p2.ParseAny(md.id(), childStartState)
		out2, _ = childOut.(PO2)
		if childErr != nil {
			out, _ := md.fn(out1, out2)
			return childState, out, childErr, out1
		}
	} else {
		out1, _ = data.(PO1)
		out2, _ = childOut.(PO2)
	}

	out, err := md.fn(out1, out2)
	if err != nil {
		childErr = childState.NewSemanticError(err.Error())
	}

	return childState, out, childErr, nil
}
func Map2[PO1, PO2 any, MO any](p1 Parser[PO1], p2 Parser[PO2], fn func(PO1, PO2) (MO, error)) Parser[MO] {
	if p1 == nil {
		panic("Map2: p1 is nil")
	}
	if p2 == nil {
		panic("Map2: p2 is nil")
	}
	if fn == nil {
		panic("Map2: fn is nil")
	}

	m2d := &map2data[PO1, PO2, MO]{
		p1: p1,
		p2: p2,
		fn: fn,
	}
	p := NewBranchParser[MO]("Map2", m2d.children, m2d.parseAfterError)
	m2d.id = p.ID
	return p
}

// ============================================================================
// Char Parser
//

func Char(char rune) Parser[rune] {
	var p Parser[rune]

	expected := strconv.QuoteRune(char)

	parse := func(state State) (State, rune, *ParserError) {
		r, size := utf8.DecodeRuneInString(state.CurrentString())
		if r == utf8.RuneError {
			if size == 0 {
				return state, utf8.RuneError, state.NewSyntaxError("%s (at EOF)", expected)
			}
			return state, utf8.RuneError, state.NewSyntaxError("%s (got UTF-8 error)", expected)
		}
		if r != char {
			return state, utf8.RuneError, state.NewSyntaxError("%s (got %q)", expected, r)
		}

		return state.MoveBy(size), r, nil
	}

	p = NewParser[rune](expected, parse, IndexOf(char))
	return p
}

func IndexOf[S Separator](stop S) Recoverer {
	// This IS type safe because of the `Separator` constraint!
	// Performance doesn't matter either because this is done during the
	// construction phase.
	switch v := reflect.ValueOf(stop); v.Kind() {
	case reflect.Uint8:
		xstop := interface{}(stop).(byte)
		return func(state State, _ interface{}) (int, interface{}) {
			waste := bytes.IndexByte(state.CurrentBytes(), xstop)
			if waste < 0 {
				return RecoverWasteTooMuch, nil
			}
			return waste, nil
		}
	case reflect.Int32:
		rstop := interface{}(stop).(rune)
		return func(state State, _ interface{}) (int, interface{}) {
			waste := strings.IndexRune(state.CurrentString(), rstop)
			if waste < 0 {
				return RecoverWasteTooMuch, nil
			}
			return waste, nil
		}
	case reflect.String:
		sstop := interface{}(stop).(string)
		if len(sstop) == 0 {
			panic("stop is empty")
		}
		return func(state State, _ interface{}) (int, interface{}) {
			waste := strings.Index(state.CurrentString(), sstop)
			if waste < 0 {
				return RecoverWasteTooMuch, nil
			}
			return waste, nil
		}
	case reflect.Slice:
		bstop := interface{}(stop).([]byte)
		if len(bstop) == 0 {
			panic("stop is empty")
		}
		return func(state State, _ interface{}) (int, interface{}) {
			waste := bytes.Index(state.CurrentBytes(), bstop)
			if waste < 0 {
				return RecoverWasteTooMuch, nil
			}
			return waste, nil
		}
	default:
		return nil // can never happen because of the `Separator` constraint!
	}
}
