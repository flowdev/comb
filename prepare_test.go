package gomme

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
	badParse := Map2(
		Map2(Char('a'), Char('b'), runePlusRune),
		Char('c'),
		stringPlusRune,
	)
	goodParse := Map2(
		Map2(SafeSpot(Char('a')), SafeSpot(Char('b')), runePlusRune),
		SafeSpot(Char('c')),
		stringPlusRune,
	)

	tests := []struct {
		name           string
		givenInput     string
		givenParser    Parser[string]
		expectedOutput interface{}
		expectedErrors int
	}{
		{
			name:           "goodInputBadParser",
			givenInput:     "abc",
			givenParser:    badParse,
			expectedOutput: "abc",
			expectedErrors: 0,
		}, {
			name:           "goodInputGoodParser",
			givenInput:     "abc",
			givenParser:    goodParse,
			expectedOutput: "abc",
			expectedErrors: 0,
		}, {
			name:           "emptyBadParser",
			givenInput:     "",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "emptyGoodParser",
			givenInput:     "",
			givenParser:    goodParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "lastCharMissingBadParser",
			givenInput:     "ab",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "lastCharMissingGoodParser",
			givenInput:     "ab",
			givenParser:    goodParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "middleCharMissingBadParser",
			givenInput:     "ac",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "middleCharMissingGoodParser",
			givenInput:     "ac",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 1,
		}, {
			name:           "firstCharMissingBadParser",
			givenInput:     "bc",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "firstCharMissingGoodParser",
			givenInput:     "bc",
			givenParser:    goodParse,
			expectedOutput: "\x00bc",
			expectedErrors: 1,
		}, {
			name:           "firstCharOffBadParser",
			givenInput:     "1abc",
			givenParser:    badParse,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "firstCharOffGoodParser",
			givenInput:     "1abc",
			givenParser:    goodParse,
			expectedOutput: "abc",
			expectedErrors: 1,
		}, {
			name:           "secondCharOffBadParser",
			givenInput:     "a1bc",
			givenParser:    badParse,
			expectedOutput: "\x00bc",
			expectedErrors: 1,
		}, {
			name:           "secondCharOffGoodParser",
			givenInput:     "a1bc",
			givenParser:    goodParse,
			expectedOutput: "\x00bc",
			expectedErrors: 1,
		}, {
			name:           "thirdCharOffBadParser",
			givenInput:     "ab1c",
			givenParser:    badParse,
			expectedOutput: "c",
			expectedErrors: 1,
		}, {
			name:           "thirdCharOffGoodParser",
			givenInput:     "ab1c",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 1,
		}, {
			name:           "firstAndLastCharOffBadParser",
			givenInput:     "1ab2c",
			givenParser:    badParse,
			expectedOutput: "c",
			expectedErrors: 2,
		}, {
			name:           "firstAndLastCharOffGoodParser",
			givenInput:     "1ab2c",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingBadParser",
			givenInput:     "1ac",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingGoodParser",
			givenInput:     "1ac",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 2,
		}, {
			name:           "allCharsOffBadParser",
			givenInput:     "1a2b3c",
			givenParser:    badParse,
			expectedOutput: "c",
			expectedErrors: 3,
		}, {
			name:           "allCharsOffGoodParser",
			givenInput:     "1a2b3c",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 3,
		}, {
			name:           "firstCharMissingLastCharOffBadParser",
			givenInput:     "b1c",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "firstCharMissingLastCharOffGoodParser",
			givenInput:     "b1c",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingBadParser",
			givenInput:     "1ac",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "firstCharOffMiddleCharMissingGoodParser",
			givenInput:     "1ac",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 2,
		}, {
			name:           "onlyFirstCharBadParser",
			givenInput:     "a",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyFirstCharGoodParser",
			givenInput:     "a",
			givenParser:    goodParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyMiddleCharBadParser",
			givenInput:     "b",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyMiddleCharGoodParser",
			givenInput:     "b",
			givenParser:    goodParse,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "onlyLastCharBadParser",
			givenInput:     "c",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 1,
		}, {
			name:           "onlyLastCharGoodParser",
			givenInput:     "c",
			givenParser:    goodParse,
			expectedOutput: "c",
			expectedErrors: 1,
		}, {
			name:           "firstCharLastBadParser",
			givenInput:     "bca",
			givenParser:    badParse,
			expectedOutput: "",
			expectedErrors: 2,
		}, {
			name:           "firstCharLastGoodParser",
			givenInput:     "bca",
			givenParser:    goodParse,
			expectedOutput: "\x00bc",
			expectedErrors: 1,
		},
	}
	SetDebug(true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepp := NewPreparedParser[string](tt.givenParser) // this calls ParserToAnyParser
			output, err := prepp.parseAll(NewFromString(tt.givenInput, true))
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
	bParse := Map2(Char('a'), Char('b'), func(out1 rune, out2 rune) (string, error) {
		return string([]rune{out1, out2}), nil
	})

	tests := []struct {
		name           string
		givenInput     string
		givenParser    Parser[string]
		expectedID     int32
		expectedOutput interface{}
		expectedError  bool
	}{
		{
			name:           "allGoodBranchParser",
			givenInput:     "ab",
			givenParser:    bParse,
			expectedID:     0,
			expectedOutput: "ab",
			expectedError:  false,
		}, {
			name:           "firstSubparserMissesInput",
			givenInput:     "b",
			givenParser:    bParse,
			expectedID:     1,
			expectedOutput: "",
			expectedError:  true,
		}, {
			name:           "OneByteOff",
			givenInput:     "1ab",
			givenParser:    bParse,
			expectedID:     1,
			expectedOutput: "",
			expectedError:  true,
		}, {
			name:           "secondSubparserMissesInput",
			givenInput:     "a",
			givenParser:    bParse,
			expectedID:     2,
			expectedOutput: "",
			expectedError:  true,
		}, {
			name:           "secondSubparserOneByteOff",
			givenInput:     "a1b",
			givenParser:    bParse,
			expectedID:     2,
			expectedOutput: "",
			expectedError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepp := NewPreparedParser[string](tt.givenParser) // this calls ParserToAnyParser
			aParse := prepp.parsers[0].parser
			result := aParse.parse(NewFromString(tt.givenInput, true))
			if got, want := aParse.IsSaveSpot(), false; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			_, gotBranchParser := aParse.(BranchParser)
			if got, want := gotBranchParser, true; got != want {
				t.Errorf("branch parser=%t, want=%t", got, want)
			}
			if got, want := result.Error, tt.expectedError; (got != nil) != want {
				t.Errorf("result.Error=%v, want=%t", got, want)
			}
			if result.Error != nil {
				if got, want := result.Error.parserID, tt.expectedID; got != want {
					t.Errorf("error parser ID=%d, want=%d", got, want)
				}
			}
			gotOutput, ok := result.Output.(string)
			if got, want := gotOutput, tt.expectedOutput; got != want {
				t.Errorf("output=%v (OK=%t), want=%v", got, ok, want)
			}

			if got, want := aParse.IsStepRecoverer(), false; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
		})
	}
}

func TestLeafParserToAnyParser(t *testing.T) {
	parse := Char('a')
	sParse := SafeSpot[rune](parse)

	tests := []struct {
		name                  string
		givenInput            string
		givenParser           Parser[rune]
		expectedSaveSpot      bool
		expectedStepRecoverer bool
		expectedOutput        interface{}
		expectedError         bool
		expectedWaste         int
	}{
		{
			name:                  "allGoodSimple",
			givenInput:            "a",
			givenParser:           parse,
			expectedSaveSpot:      false,
			expectedStepRecoverer: false,
			expectedOutput:        'a',
			expectedError:         false,
			expectedWaste:         0,
		}, {
			name:                  "allGoodSaveSpot",
			givenInput:            "a",
			givenParser:           sParse,
			expectedSaveSpot:      true,
			expectedStepRecoverer: false,
			expectedOutput:        'a',
			expectedError:         false,
			expectedWaste:         0,
		}, {
			name:                  "emptySimple",
			givenInput:            "",
			givenParser:           parse,
			expectedSaveSpot:      false,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         RecoverWasteTooMuch,
		}, {
			name:                  "emptySaveSpot",
			givenInput:            "",
			givenParser:           sParse,
			expectedSaveSpot:      true,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         RecoverWasteTooMuch,
		}, {
			name:                  "twoBytesOffSimple",
			givenInput:            "bca",
			givenParser:           parse,
			expectedSaveSpot:      false,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         2,
		}, {
			name:                  "twoBytesOffSaveSpot",
			givenInput:            "bca",
			givenParser:           sParse,
			expectedSaveSpot:      true,
			expectedStepRecoverer: false,
			expectedOutput:        utf8.RuneError,
			expectedError:         true,
			expectedWaste:         2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepp := NewPreparedParser[rune](tt.givenParser) // this calls ParserToAnyParser
			aParse := prepp.parsers[0].parser
			result := aParse.parse(NewFromString(tt.givenInput, true))
			if got, want := tt.givenParser.IsSaveSpot(), tt.expectedSaveSpot; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			_, gotBranchParser := tt.givenParser.(BranchParser)
			if got, want := gotBranchParser, false; got != want {
				t.Errorf("branch parser=%t, want=%t", got, want)
			}
			if got, want := result.Error, tt.expectedError; (got != nil) != want {
				t.Errorf("result.Error=%v, want=%t", got, want)
			}
			if result.Error != nil {
				if got, want := result.Error.parserID, aParse.ID(); got != want {
					t.Errorf("error parser ID=%d, want=%d", got, want)
				}
			}
			gotOutput, ok := result.Output.(rune)
			if got, want := gotOutput, tt.expectedOutput; got != want {
				t.Errorf("output=%v (OK=%t), want=%v", got, ok, want)
			}

			if got, want := tt.givenParser.IsStepRecoverer(), tt.expectedStepRecoverer; got != want {
				t.Errorf("save spot parser=%t, want=%t", got, want)
			}
			if !tt.givenParser.IsStepRecoverer() {
				waste := tt.givenParser.Recover(NewFromString(tt.givenInput, true))
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
}

func (md *map2data[PO1, PO2, MO]) children() []AnyParser {
	return []AnyParser{md.p1, md.p2}
}
func (md *map2data[PO1, PO2, MO]) parseAfterChild(childID int32, childResult ParseResult) ParseResult {
	var zero MO
	var zero1 PO1
	var zero2 PO2

	if childResult.Error != nil {
		return childResult // we can't avoid any errors by going another path
	}

	state := childResult.EndState
	id := childID
	Debugf("Map2 - pos=%d; parse after ID %d", state.CurrentPos(), id)
	if id >= 0 && id != md.p1.ID() && id != md.p2.ID() {
		return ParseResult{
			StartState: state,
			EndState:   state,
			Output:     zero,
			Error:      state.NewSemanticError("unable to parse after child with unknown ID %d", id),
		}
	}

	state1, out1, err1 := state, zero1, (*ParserError)(nil)
	if id < 0 {
		state1, out1, err1 = md.p1.Parse(state)
		if err1 != nil {
			return ParseResult{StartState: state, EndState: state1, Output: out1, Error: err1}
		}
	}
	if id == md.p1.ID() {
		state1 = childResult.EndState
		out1, _ = childResult.Output.(PO1)
		err1 = childResult.Error
	}

	state2, out2, err2 := state, zero2, (*ParserError)(nil)
	if id == md.p2.ID() {
		state2 = childResult.EndState
		out2, _ = childResult.Output.(PO2)
		err2 = childResult.Error
	} else {
		state2, out2, err2 = md.p2.Parse(state1)
		if err2 != nil {
			return ParseResult{StartState: state1, EndState: state2, Output: out2, Error: err2}
		}
	}

	out, err := md.fn(out1, out2)
	var pErr *ParserError
	if err != nil {
		pErr = state2.NewSemanticError(err.Error())
	}

	return ParseResult{StartState: state, EndState: state2, Output: out, Error: pErr}
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
	return NewBranchParser[MO]("Map2", m2d.children, m2d.parseAfterChild)
}

// ============================================================================
// Char Parser
//

func Char(char rune) Parser[rune] {
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

	return NewParser[rune](expected, parse, IndexOf(char))
}

func IndexOf[S Separator](stop S) Recoverer {
	// This IS type safe because of the `Separator` constraint!
	// Performance doesn't matter either because this is done during the
	// construction phase.
	switch v := reflect.ValueOf(stop); v.Kind() {
	case reflect.Uint8:
		xstop := interface{}(stop).(byte)
		return func(state State) int {
			waste := bytes.IndexByte(state.CurrentBytes(), xstop)
			if waste < 0 {
				return RecoverWasteTooMuch
			}
			return waste
		}
	case reflect.Int32:
		rstop := interface{}(stop).(rune)
		return func(state State) int {
			waste := strings.IndexRune(state.CurrentString(), rstop)
			if waste < 0 {
				return RecoverWasteTooMuch
			}
			return waste
		}
	case reflect.String:
		sstop := interface{}(stop).(string)
		if len(sstop) == 0 {
			panic("stop is empty")
		}
		return func(state State) int {
			waste := strings.Index(state.CurrentString(), sstop)
			if waste < 0 {
				return RecoverWasteTooMuch
			}
			return waste
		}
	case reflect.Slice:
		bstop := interface{}(stop).([]byte)
		if len(bstop) == 0 {
			panic("stop is empty")
		}
		return func(state State) int {
			waste := bytes.Index(state.CurrentBytes(), bstop)
			if waste < 0 {
				return RecoverWasteTooMuch
			}
			return waste
		}
	default:
		return nil // can never happen because of the `Separator` constraint!
	}
}
