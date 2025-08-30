package comb_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/flowdev/comb"
	"github.com/stretchr/testify/assert"
)

func TestErrorReporting(t *testing.T) {
	input := "content\nline2\nline3\nand4\n"
	input2 := "line1\nline2"
	txtState := comb.NewFromString(input, 10)
	binState := comb.NewFromBytes([]byte(input), 10)

	specs := []struct {
		name          string
		givenState    comb.State
		givenPosition int
		expectedError string
	}{
		{
			name:          "at start of line in the middle",
			givenState:    txtState,
			givenPosition: 14,
			expectedError: "error [3:1] ▶line3",
		}, {
			name:          "at end of input with last NL",
			givenState:    txtState,
			givenPosition: len(input),
			expectedError: "error [5:1] ▶",
		}, {
			name:          "at NL in the middle",
			givenState:    txtState,
			givenPosition: 13,
			expectedError: "error [2:6] line2▶",
		}, {
			name:          "at start of input",
			givenState:    txtState,
			givenPosition: 0,
			expectedError: "error [1:1] ▶content",
		}, {
			name:          "empty input",
			givenState:    comb.NewFromString("", 10),
			givenPosition: 0,
			expectedError: "error [1:1] ▶",
		}, {
			name:          "at end of input without last NL",
			givenState:    comb.NewFromString(input2, 10),
			givenPosition: len(input2),
			expectedError: "error [2:6] line2▶",
		}, {
			name:          "binary: at start of input",
			givenState:    binState,
			givenPosition: 0,
			expectedError: "error:\n 00000000  ▶63 6f 6e 74 65 6e 74 0a  6c 69 6e 65 32 0a 6c 69  |▶content.line2.li|",
		}, {
			name:          "binary: in middle of input",
			givenState:    binState,
			givenPosition: 10,
			expectedError: "error:\n 00000002  6e 74 65 6e 74 0a 6c 69  ▶6e 65 32 0a 6c 69 6e 65  |ntent.li▶ne2.line|",
		}, {
			name:          "binary: at end of input",
			givenState:    binState,
			givenPosition: len(input),
			expectedError: "error:\n 00000009  69 6e 65 32 0a 6c 69 6e  65 33 0a 61 6e 64 34 0a ▶ |ine2.line3.and4.▶|",
		}, {
			name:          "binary: at start of short input",
			givenState:    comb.NewFromBytes([]byte(input2), 10),
			givenPosition: 0,
			expectedError: "error:\n 00000000  ▶6c 69 6e 65 31 0a 6c 69  6e 65 32                 |▶line1.line2|",
		}, {
			name:          "binary: in middle of short input",
			givenState:    comb.NewFromBytes([]byte(input2), 10),
			givenPosition: 8,
			expectedError: "error:\n 00000000  6c 69 6e 65 31 0a 6c 69  ▶6e 65 32                 |line1.li▶ne2|",
		}, {
			name:          "binary: at end of short input",
			givenState:    comb.NewFromBytes([]byte(input2), 10),
			givenPosition: len(input2) - 1,
			expectedError: "error:\n 00000000  6c 69 6e 65 31 0a 6c 69  6e 65 ▶32                 |line1.line▶2|",
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			nState := spec.givenState.MoveBy(spec.givenPosition)
			gotError := nState.SaveError(nState.NewSemanticError("error")).Errors().Error()

			if gotError != spec.expectedError {
				t.Errorf("Expected error %q, got: %q", spec.expectedError, gotError)
			}

		})
	}
}

func TestErrorReportingWithMoveBackTo(t *testing.T) {
	input := "content\nline2\nline3\nand4\n"
	txtState := comb.NewFromString(input, 10).MoveBy(len(input))
	binState := comb.NewFromBytes([]byte(input), 10).MoveBy(len(input))

	specs := []struct {
		name          string
		givenState    comb.State
		givenPosition int
		expectedError string
	}{
		{
			name:          "at start of line in the middle",
			givenState:    txtState,
			givenPosition: 14,
			expectedError: "error [3:1] ▶line3",
		}, {
			name:          "at end of input with last NL",
			givenState:    txtState,
			givenPosition: len(input) + 1,
			expectedError: "error [5:1] ▶",
		}, {
			name:          "at NL in the middle",
			givenState:    txtState,
			givenPosition: 13,
			expectedError: "error [2:6] line2▶",
		}, {
			name:          "at start of input",
			givenState:    txtState,
			givenPosition: 0,
			expectedError: "error [1:1] ▶content",
		}, {
			name:          "empty input",
			givenState:    comb.NewFromString("", 10),
			givenPosition: 0,
			expectedError: "error [1:1] ▶",
		}, {
			name:          "binary: at start of input",
			givenState:    binState,
			givenPosition: 0,
			expectedError: "error:\n 00000000  ▶63 6f 6e 74 65 6e 74 0a  6c 69 6e 65 32 0a 6c 69  |▶content.line2.li|",
		}, {
			name:          "binary: in middle of input",
			givenState:    binState,
			givenPosition: 10,
			expectedError: "error:\n 00000002  6e 74 65 6e 74 0a 6c 69  ▶6e 65 32 0a 6c 69 6e 65  |ntent.li▶ne2.line|",
		}, {
			name:          "binary: at end of input",
			givenState:    binState,
			givenPosition: len(input),
			expectedError: "error:\n 00000009  69 6e 65 32 0a 6c 69 6e  65 33 0a 61 6e 64 34 0a ▶ |ine2.line3.and4.▶|",
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			nState := spec.givenState.MoveBackTo(spec.givenPosition)
			gotError := nState.SaveError(nState.NewSemanticError("error")).Errors().Error()

			if gotError != spec.expectedError {
				t.Errorf("Expected error %q, got: %q", spec.expectedError, gotError)
			}

		})
	}
}

func TestRunOnBytes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[[]byte]
		input      []byte
		wantErr    bool
		wantOutput []byte
	}{
		{
			name:       "simple match",
			input:      []byte{1, 2, 3, 4},
			parser:     Bytes([]byte{1, 2}),
			wantErr:    false,
			wantOutput: []byte{1, 2},
		}, {
			name:       "no match",
			input:      []byte{1, 2, 3, 4},
			parser:     Bytes([]byte{2, 3}),
			wantErr:    true,
			wantOutput: []byte{2, 3},
		}, {
			name:       "empty input should fail",
			input:      nil,
			parser:     Bytes([]byte{1, 2}),
			wantErr:    true,
			wantOutput: []byte{},
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult, gotErr := comb.RunOnBytes(tc.input, tc.parser)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			// testify makes it easier to compare slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)
		})
	}
}

func TestMiscStuff(t *testing.T) {
	t.Parallel()

	t.Run("test BetterOf", func(t *testing.T) {
		t.Parallel()

		state1 := comb.NewFromString("12345", 0)
		state2 := state1.MoveBy(1)

		gotState, gotOther := comb.BetterOf(state1, state2)
		if got, want := gotState.CurrentPos(), 1; got != want {
			t.Errorf("got state position: %d, want: %d", got, want)
		}
		if got, want := gotOther, true; got != want {
			t.Errorf("got other: %t, want: %t", got, want)
		}

		gotState, gotOther = comb.BetterOf(state2, state1)
		if got, want := gotState.CurrentPos(), 1; got != want {
			t.Errorf("got state position: %d, want: %d", got, want)
		}
		if got, want := gotOther, false; got != want {
			t.Errorf("got other: %t, want: %t", got, want)
		}
	})

	t.Run("test UnwrapErrors", func(t *testing.T) {
		t.Parallel()

		err1 := errors.New("error 1")
		err2 := errors.New("error 2")

		assert.Equal(t, []error(nil), comb.UnwrapErrors(nil))
		assert.Equal(t, []error{err1}, comb.UnwrapErrors(err1))
		assert.Equal(t, []error{err2, err1}, comb.UnwrapErrors(errors.Join(err2, err1)))
	})

	t.Run("test ZeroOf", func(t *testing.T) {
		t.Parallel()

		if got, want := comb.ZeroOf[string](), ""; got != want {
			t.Errorf("got %#v, want: %#v", got, want)
		}
		if got, want := comb.ZeroOf[int](), 0; got != want {
			t.Errorf("got %#v, want: %#v", got, want)
		}
		assert.Equal(t, []byte(nil), comb.ZeroOf[[]byte]())
	})

	t.Run("test SetDebug", func(t *testing.T) {
		t.Parallel()

		comb.SetDebug(true)
		if got, want := slog.SetLogLoggerLevel(slog.LevelDebug), slog.LevelDebug; got != want {
			t.Errorf("got %#v, want: %#v", got, want)
		}
		comb.SetDebug(false)
		if got, want := slog.SetLogLoggerLevel(slog.LevelDebug), slog.LevelInfo; got != want {
			t.Errorf("got %#v, want: %#v", got, want)
		}
		comb.SetDebug(false) // turn debug logging of for other tests
	})
}

// Bytes parses a token from the input and returns the part of the input that
// matched the token.
// If the token could not be found at the current position,
// the parser returns an error result.
func Bytes(token []byte) comb.Parser[[]byte] {
	var p comb.Parser[[]byte]

	expected := fmt.Sprintf("0x%x", token)

	parse := func(state comb.State) (comb.State, []byte, *comb.ParserError) {
		if !bytes.HasPrefix(state.CurrentBytes(), token) {
			return state, []byte{}, state.NewSyntaxError(expected)
		}

		newState := state.MoveBy(len(token))
		return newState, token, nil
	}

	p = comb.NewParser[[]byte](expected, parse, IndexOf(token))
	return p
}
func IndexOf(stop []byte) comb.Recoverer {
	if len(stop) == 0 {
		panic("stop is empty")
	}
	return func(state comb.State, _ interface{}) (int, interface{}) {
		waste := bytes.Index(state.CurrentBytes(), stop)
		if waste < 0 {
			return comb.RecoverWasteTooMuch, nil
		}
		return waste, nil
	}
}
