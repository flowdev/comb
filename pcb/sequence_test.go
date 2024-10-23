package pcb

import (
	"github.com/oleiade/gomme"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[[]string]
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
				parser: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       false,
			wantOutput:    []string{"1", "a", "3"},
			wantRemaining: "",
		},
		{
			name:  "matching parsers in longer input should succeed",
			input: "1a3bcd",
			args: args{
				parser: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       false,
			wantOutput:    []string{"1", "a", "3"},
			wantRemaining: "bcd",
		},
		{
			name:  "partially matching parsers should fail",
			input: "1a?",
			args: args{
				parser: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       true,
			wantOutput:    []string{"1", "a", ""},
			wantRemaining: "?",
		},
		{
			name:  "too short input should fail",
			input: "1a",
			args: args{
				parser: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       true,
			wantOutput:    []string{"1", "a", ""},
			wantRemaining: "",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				parser: Sequence(Digit1(), Alpha0(), Digit1()),
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

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(0, nil, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
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
