package pcb

import (
	"github.com/oleiade/gomme"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlternative(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[string]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:  "head matching parser should succeed",
			input: "123",
			args: args{
				p: Alternative(Digit1(), Alpha0()),
			},
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:  "matching parser should succeed",
			input: "1",
			args: args{
				p: Alternative(Digit1(), Alpha0()),
			},
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:  "no matching parser should fail",
			input: "$%^*",
			args: args{
				p: Alternative(Digit1(), Alpha1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$%^*",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Alternative(Digit1(), Alpha1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := gomme.NewFromString(tc.input)
			newState, gotResult := tc.args.p(state)
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)

			if newState.CurrentString() != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", newState.CurrentString(), tc.wantRemaining)
			}
		})
	}
}

func BenchmarkAlternative(b *testing.B) {
	p := Alternative(Digit1(), Alpha1())
	input := gomme.NewFromString("123")

	for i := 0; i < b.N; i++ {
		_, _ = p(input)
	}
}
