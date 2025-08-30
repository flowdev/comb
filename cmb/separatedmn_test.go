package cmb

import (
	"testing"

	"github.com/flowdev/comb"
)

func TestSeparatedMN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		wantElements  int
		wantErr       bool
		wantRemaining string
	}{
		{
			name:          "1 element",
			input:         "abc,",
			wantErr:       false,
			wantElements:  1,
			wantRemaining: "",
		}, {
			name:          "3 elements",
			input:         "abc,abc,abc,",
			wantErr:       false,
			wantElements:  3,
			wantRemaining: "",
		}, {
			name:          "4 elements",
			input:         "abc,abc,abc,abc,",
			wantErr:       false,
			wantElements:  3,
			wantRemaining: "abc,",
		}, {
			name:          "no separator at end",
			input:         "abc,abc;",
			wantErr:       false,
			wantElements:  2,
			wantRemaining: ";",
		}, {
			name:          "error input",
			input:         "ab,",
			wantErr:       true,
			wantElements:  1,
			wantRemaining: "ab,",
		}, {
			name:          "empty input",
			input:         "",
			wantErr:       true,
			wantElements:  1,
			wantRemaining: "",
		},
	}
	for _, tt := range tests {
		tt := tt // needed for truly different test cases!
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := SeparatedMN[string, rune](String("abc"), Char(','), 1, 3, true)
			nState, gotOut, err := p.Parse(comb.NewFromString(tt.input, 9))
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("got error %v, want: %t", err, tt.wantErr)
			}
			gotRemaining := nState.CurrentString()
			if gotRemaining != tt.wantRemaining {
				t.Errorf("got remaining input %q, want: %q", gotRemaining, tt.wantRemaining)
			}
			t.Logf("got output %q", gotOut)
			if got, want := len(gotOut), tt.wantElements; got != want {
				t.Errorf("got %d elements, want: %d", got, want)
			}
		})
	}
}
