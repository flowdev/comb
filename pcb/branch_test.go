package pcb

import (
	"github.com/oleiade/gomme"
	"strings"
	"testing"
)

func TestNoWayBack(t *testing.T) {
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
				p: Alternative(Digit1(), NoWayBack(Alpha1())),
			},
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:  "tail matching parser should succeed",
			input: "abc",
			args: args{
				p: Alternative(NoWayBack(Digit1()), Alpha1()),
			},
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:  "Alternative: tail matching parser after failing NoWayBack head parser should fail",
			input: "abc",
			args: args{
				p: Alternative(Preceded(NoWayBack(String("a")), Digit1()), Alpha1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
		{
			name:  "Optional: tail matching parser after failing NoWayBack head parser should fail",
			input: "abc",
			args: args{
				p: Optional(Preceded(NoWayBack(String("a")), Digit1())),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
		{
			name:  "Many0: tail matching parser after failing NoWayBack head parser should fail",
			input: "abc",
			args: args{
				p: Map(Many0(Preceded(NoWayBack(String("a")), Digit1())), func(tokens []string) (string, error) {
					return strings.Join(tokens, ""), nil
				}),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
		{
			name:  "Seperated1: matching main parser after failing NoWayBack head parser should fail",
			input: "a,1",
			args: args{
				p: Map(Separated0(Preceded(NoWayBack(String("a")), Digit1()), Char(','), false),
					func(tokens []string) (string, error) {
						return strings.Join(tokens, ""), nil
					},
				),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "a,1",
		},
		{
			name:  "no matching parser should fail",
			input: "$%^*",
			args: args{
				p: Alternative(NoWayBack(Digit1()), NoWayBack(Alpha1())),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$%^*",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Alternative(NoWayBack(Digit1()), NoWayBack(Alpha1())),
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want %q", gotResult, tc.wantOutput)
			}

			if newState.CurrentString() != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", newState.CurrentString(), tc.wantRemaining)
			}
		})
	}
}

func BenchmarkNoWayBack(b *testing.B) {
	p := NoWayBack(Char('1'))
	input := gomme.NewFromString("123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p(input)
	}
}

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
			name:  "tail matching parser should succeed",
			input: "abc",
			args: args{
				p: Alternative(Digit1(), Alpha0()),
			},
			wantErr:       false,
			wantOutput:    "abc",
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want %q", gotResult, tc.wantOutput)
			}

			if newState.CurrentString() != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", newState.CurrentString(), tc.wantRemaining)
			}
		})
	}
}

func BenchmarkAlternative(b *testing.B) {
	p := Alternative(Char('b'), Char('a'))
	input := gomme.NewFromString("abc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p(input)
	}
}
