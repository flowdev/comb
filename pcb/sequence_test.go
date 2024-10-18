package pcb

import (
	"github.com/oleiade/gomme"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[[]string]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    []string
		wantRemaining string
	}{
		{
			name:  "matching parsers should succeed",
			input: "1a3",
			args: args{
				p: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       false,
			wantOutput:    []string{"1", "a", "3"},
			wantRemaining: "",
		},
		{
			name:  "matching parsers in longer input should succeed",
			input: "1a3bcd",
			args: args{
				p: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       false,
			wantOutput:    []string{"1", "a", "3"},
			wantRemaining: "bcd",
		},
		{
			name:  "partially matching parsers should fail",
			input: "1a3",
			args: args{
				p: Sequence(Digit1(), Digit1(), Digit1()),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "1a3",
		},
		{
			name:  "too short input should fail",
			input: "12",
			args: args{
				p: Sequence(Digit1(), Digit1(), Digit1()),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "12",
		},
		{
			name:  "empty input should succeed",
			input: "",
			args: args{
				p: Sequence(Digit1(), Digit1(), Digit1()),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := tc.args.p.It(gomme.NewFromString(1, nil, tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkSequence(b *testing.B) {
	parser := Sequence(Digit1(), Alpha0(), Digit1())
	input := gomme.NewFromString(1, nil, "123A45")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(input)
	}
}
