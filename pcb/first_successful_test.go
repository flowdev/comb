package pcb

import (
	"github.com/oleiade/gomme"
	"testing"
)

func TestFirstSuccessful(t *testing.T) {
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
				p: FirstSuccessful(Digit1(), Alpha0()),
			},
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:  "tail matching parser should succeed",
			input: "abc",
			args: args{
				p: FirstSuccessful(Digit1(), Alpha0()),
			},
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:  "no matching parser should fail",
			input: "$%^*",
			args: args{
				p: FirstSuccessful(Digit1(), Alpha1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$%^*",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: FirstSuccessful(Digit1(), Alpha1()),
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

			state := gomme.NewFromString(0, nil, tc.input)
			newState, gotResult := tc.args.p.It(state)
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want %q", gotResult, tc.wantOutput)
			}

			if newState.CurrentString() != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", newState.CurrentString(), tc.wantRemaining)
			}
		})
	}
}

func BenchmarkFirstSuccessful(b *testing.B) {
	p := FirstSuccessful(Char('b'), Char('a'))
	input := gomme.NewFromString(1, nil, "abc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
	}
}
