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
		name           string
		givenInput     string
		givenBadParser bool
		expectedOutput interface{}
		expectedErrors int
	}{
		{
			name:           "goodInputBadParser",
			givenInput:     "abc",
			givenBadParser: true,
			expectedOutput: "abc",
			expectedErrors: 0,
		}, {
			name:           "goodInputGoodParser",
			givenInput:     "abc",
			givenBadParser: false,
			expectedOutput: "abc",
			expectedErrors: 0,
		}, {
			name:           "emptyBadParser",
			givenInput:     "",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "emptyGoodParser",
			givenInput:     "",
			givenBadParser: false,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "lastCharMissingBadParser",
			givenInput:     "ab",
			givenBadParser: true,
			expectedOutput: "ab�",
			expectedErrors: 1,
		}, {
			name:           "lastCharMissingGoodParser",
			givenInput:     "ab",
			givenBadParser: false,
			expectedOutput: "ab�",
			expectedErrors: 1,
		}, {
			name:           "middleCharMissingBadParser",
			givenInput:     "ac",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "middleCharMissingGoodParser",
			givenInput:     "ac",
			givenBadParser: false,
			expectedOutput: "a\ufffdc",
			expectedErrors: 1,
		}, {
			name:           "firstCharMissingBadParser",
			givenInput:     "bc",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "firstCharMissingGoodParser",
			givenInput:     "bc",
			givenBadParser: false,
			expectedOutput: "\ufffdbc",
			expectedErrors: 1,
		}, {
			name:           "firstCharOffBadParser",
			givenInput:     "1abc",
			givenBadParser: true,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "firstCharOffGoodParser",
			givenInput:     "1abc",
			givenBadParser: false,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "secondCharOffBadParser",
			givenInput:     "a1bc",
			givenBadParser: true,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "secondCharOffGoodParser",
			givenInput:     "a1bc",
			givenBadParser: false,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "thirdCharOffBadParser",
			givenInput:     "ab1c",
			givenBadParser: true,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "thirdCharOffGoodParser",
			givenInput:     "ab1c",
			givenBadParser: false,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "firstAndLastCharOffBadParser",
			givenInput:     "1ab2c",
			givenBadParser: true,
			expectedOutput: "abc",
			expectedErrors: 2,
		}, {
			name:           "firstAndLastCharOffGoodParser",
			givenInput:     "1ab2c",
			givenBadParser: false,
			expectedOutput: "abc",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingBadParser",
			givenInput:     "1ac",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingGoodParser",
			givenInput:     "1ac",
			givenBadParser: false,
			expectedOutput: "a�c",
			expectedErrors: 2,
		}, {
			name:           "allCharsOffBadParser",
			givenInput:     "1a2b3c",
			givenBadParser: true,
			expectedOutput: "abc",
			expectedErrors: 3,
		}, {
			name:           "allCharsOffGoodParser",
			givenInput:     "1a2b3c",
			givenBadParser: false,
			expectedOutput: "abc",
			expectedErrors: 3,
		}, {
			name:           "firstCharMissingLastCharOffBadParser",
			givenInput:     "b1c",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "firstCharMissingLastCharOffGoodParser",
			givenInput:     "b1c",
			givenBadParser: false,
			expectedOutput: "\ufffdbc",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingBadParser",
			givenInput:     "1ac",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingGoodParser",
			givenInput:     "1ac",
			givenBadParser: false,
			expectedOutput: "a�c",
			expectedErrors: 2,
		}, {
			name:           "onlyFirstCharBadParser",
			givenInput:     "a",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyFirstCharGoodParser",
			givenInput:     "a",
			givenBadParser: false,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyMiddleCharBadParser",
			givenInput:     "b",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyMiddleCharGoodParser",
			givenInput:     "b",
			givenBadParser: false,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "onlyLastCharBadParser",
			givenInput:     "c",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyLastCharGoodParser",
			givenInput:     "c",
			givenBadParser: false,
			expectedOutput: "c",
			expectedErrors: 1,
		}, {
			name:           "firstCharLastBadParser",
			givenInput:     "bca",
			givenBadParser: true,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "firstCharLastGoodParser",
			givenInput:     "bca",
			givenBadParser: false,
			expectedOutput: "\ufffdbc",
			expectedErrors: 1,
		},
	}
	SetDebug(true)
	for _, tc := range tests {
		tt := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tt.name, func(t *testing.T) {
			var givenParser Parser[string]
			if tt.givenBadParser {
				givenParser = Map2(
					Map2(Char('a'), Char('b'), runePlusRune),
					Char('c'),
					stringPlusRune,
				)
			} else {
				givenParser = Map2(
					Map2(SafeSpot(Char('a')), SafeSpot(Char('b')), runePlusRune),
					SafeSpot(Char('c')),
					stringPlusRune,
				)
			}
			prepp := NewPreparedParser[string](givenParser) // this calls ParserToAnyParser
			output, err := prepp.parseAll(NewFromString(tt.givenInput, 10))
			t.Logf("err=%v", err)
			if got, want := len(UnwrapErrors(err)), tt.expectedErrors; got != want {
				t.Errorf("err=%v, want=%d", err, want)
			}
			if got, want := output, tt.expectedOutput; got != want {
				t.Errorf("got output=%q, want=%q", got, want)
			}
		})
	}
}

func TestBranchParserToAnyParser(t *testing.T) {
	tests := []struct {
		name           string
		givenInput     string
		expectedID     int32
		expectedOutput interface{}
		expectedError  bool
	}{
		{
			name:           "allGoodBranchParser",
			givenInput:     "ab",
			expectedID:     0,
			expectedOutput: "ab",
			expectedError:  false,
		}, {
			name:           "firstSubparserMissesInput",
			givenInput:     "b",
			expectedID:     1,
			expectedOutput: "",
			expectedError:  true,
		}, {
			name:           "OneByteOff",
			givenInput:     "1ab",
			expectedID:     1,
			expectedOutput: "",
			expectedError:  true,
		}, {
			name:           "secondSubparserMissesInput",
			givenInput:     "a",
			expectedID:     2,
			expectedOutput: "a\ufffd",
			expectedError:  true,
		}, {
			name:           "secondSubparserOneByteOff",
			givenInput:     "a1b",
			expectedID:     2,
			expectedOutput: "a\ufffd",
			expectedError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			givenParser := Map2(Char('a'), Char('b'), func(out1 rune, out2 rune) (string, error) {
				return string([]rune{out1, out2}), nil
			})
			prepp := NewPreparedParser[string](givenParser) // this calls ParserToAnyParser
			aParse := prepp.parsers[0]
			_, out, err := aParse.ParseAny(-1, NewFromString(tt.givenInput, 10))
			if got, want := aParse.IsSaveSpot(), false; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			_, gotBranchParser := aParse.(BranchParser)
			if got, want := gotBranchParser, true; got != want {
				t.Errorf("branch parser=%t, want=%t", got, want)
			}
			if got, want := err != nil, tt.expectedError; got != want {
				t.Errorf("result.Error=%v, want=%t", got, want)
			}
			if err != nil {
				if got, want := err.parserID, tt.expectedID; got != want {
					t.Errorf("error parser ID=%d, want=%d", got, want)
				}
			}
			gotOutput, ok := out.(string)
			if got, want := gotOutput, tt.expectedOutput; got != want {
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
		name                  string
		givenInput            string
		givenSafeParser       bool
		expectedSaveSpot      bool
		expectedStepRecoverer bool
		expectedOutput        interface{}
		expectedError         bool
		expectedWaste         int
	}{
		{
			name:                  "allGoodSimple",
			givenInput:            "a",
			givenSafeParser:       false,
			expectedSaveSpot:      false,
			expectedStepRecoverer: false,
			expectedOutput:        'a',
			expectedError:         false,
			expectedWaste:         0,
		}, {
			name:                  "allGoodSaveSpot",
			givenInput:            "a",
			givenSafeParser:       true,
			expectedSaveSpot:      true,
			expectedStepRecoverer: false,
			expectedOutput:        'a',
			expectedError:         false,
			expectedWaste:         0,
		}, {
			name:                  "emptySimple",
			givenInput:            "",
			givenSafeParser:       false,
			expectedSaveSpot:      false,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         RecoverWasteTooMuch,
		}, {
			name:                  "emptySaveSpot",
			givenInput:            "",
			givenSafeParser:       true,
			expectedSaveSpot:      true,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         RecoverWasteTooMuch,
		}, {
			name:                  "twoBytesOffSimple",
			givenInput:            "bca",
			givenSafeParser:       false,
			expectedSaveSpot:      false,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         2,
		}, {
			name:                  "twoBytesOffSaveSpot",
			givenInput:            "bca",
			givenSafeParser:       true,
			expectedSaveSpot:      true,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var givenParser Parser[rune]
			if tt.givenSafeParser {
				givenParser = SafeSpot[rune](Char('a'))
			} else {
				givenParser = Char('a')
			}

			prepp := NewPreparedParser[rune](givenParser) // this calls ParserToAnyParser
			aParse := prepp.parsers[0]
			_, out, err := aParse.ParseAny(-1, NewFromString(tt.givenInput, 10))
			if got, want := givenParser.IsSaveSpot(), tt.expectedSaveSpot; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			_, gotBranchParser := givenParser.(BranchParser)
			if got, want := gotBranchParser, false; got != want {
				t.Errorf("branch parser=%t, want=%t", got, want)
			}
			if got, want := err != nil, tt.expectedError; got != want {
				t.Errorf("result.Error=%v, want=%t", got, want)
			}
			if err != nil {
				if got, want := err.parserID, aParse.ID(); got != want {
					t.Errorf("error parser ID=%d, want=%d", got, want)
				}
			}
			gotOutput, ok := out.(rune)
			if got, want := gotOutput, tt.expectedOutput; got != want {
				t.Errorf("output=%v (OK=%t), want=%v", out, ok, want)
			}

			if got, want := givenParser.IsStepRecoverer(), tt.expectedStepRecoverer; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			if !givenParser.IsStepRecoverer() {
				waste, _ := givenParser.Recover(NewFromString(tt.givenInput, 10), nil)
				if got, want := waste, tt.expectedWaste; got != want {
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
