package gomme

import (
	"testing"
)

func TestWhere(t *testing.T) {
	st := NewFromString("content\nline2\nline3\nand4\n").MoveBy(15)

	specs := []struct {
		name            string
		givenState      State
		givenPosition   int
		expectedLine    int
		expectedCol     int
		expectedSrcLine string
	}{
		{
			name:            "at start of line in the middle",
			givenState:      st,
			givenPosition:   14,
			expectedLine:    3,
			expectedCol:     0,
			expectedSrcLine: "line3",
		}, {
			name:            "at end of input with last NL",
			givenState:      st,
			givenPosition:   len(st.input.bytes) - 1,
			expectedLine:    4,
			expectedCol:     4,
			expectedSrcLine: "and4",
		}, {
			name:            "at NL in the middle",
			givenState:      st,
			givenPosition:   13,
			expectedLine:    2,
			expectedCol:     5,
			expectedSrcLine: "line2",
		}, {
			givenState:      st,
			givenPosition:   0,
			expectedLine:    1,
			expectedCol:     0,
			expectedSrcLine: "content",
		}, {
			name: "empty input",
			givenState: State{
				input: Input{
					bytes:  []byte{},
					pos:    0,
					prevNl: -1,
					line:   1,
				},
				pointOfNoReturn: -1,
			},
			givenPosition:   0,
			expectedLine:    1,
			expectedCol:     0,
			expectedSrcLine: "",
		}, {
			name: "at end of input without last NL",
			givenState: State{
				input: Input{
					bytes:  []byte("line1\nline2"),
					pos:    7,
					prevNl: 5,
					line:   2,
				},
				pointOfNoReturn: -1,
			},
			givenPosition:   10,
			expectedLine:    2,
			expectedCol:     4,
			expectedSrcLine: "line2",
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			gotLine, gotCol, gotSrcLine := spec.givenState.where(spec.givenPosition)

			if gotLine != spec.expectedLine {
				t.Errorf("Expected line %d, got: %d", spec.expectedLine, gotLine)
			}
			if gotCol != spec.expectedCol {
				t.Errorf("Expected col %d, got: %d", spec.expectedCol, gotCol)
			}
			if gotSrcLine != spec.expectedSrcLine {
				t.Errorf("Expected source line %q, got: %q", spec.expectedSrcLine, gotSrcLine)
			}

		})
	}
}
