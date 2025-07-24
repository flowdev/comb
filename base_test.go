package comb_test

import (
	"github.com/flowdev/comb"
	"testing"
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
			gotError := nState.SaveError(nState.NewSemanticError(42, "error")).Errors().Error()

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
			gotError := nState.SaveError(nState.NewSemanticError(42, "error")).Errors().Error()

			if gotError != spec.expectedError {
				t.Errorf("Expected error %q, got: %q", spec.expectedError, gotError)
			}

		})
	}
}
