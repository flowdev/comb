package gomme_test

import (
	"github.com/flowdev/comb"
	"github.com/flowdev/comb/pcb"
	"strings"
	"testing"
)

func TestSaveSpot(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "head matching parser should succeed",
			input:         "123",
			parser:        pcb.FirstSuccessful(pcb.Digit1(), gomme.SafeSpot(pcb.Alpha1())),
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		}, {
			name:          "tail matching parser should succeed",
			input:         "abc",
			parser:        pcb.FirstSuccessful(gomme.SafeSpot(pcb.Digit1()), pcb.Alpha1()),
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		}, {
			name:          "FirstSuccessful: tail matching parser after failing SafeSpot head parser should fail",
			input:         "abc",
			parser:        pcb.FirstSuccessful(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1()), pcb.Alpha1()),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		}, {
			name:          "Optional: tail matching parser after failing SafeSpot head parser should fail",
			input:         "abc",
			parser:        pcb.Optional(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1())),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "bc",
		}, {
			name:  "Many0: tail matching parser after failing SafeSpot head parser should fail",
			input: "abc",
			parser: pcb.Map(pcb.Many0(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1())), func(tokens []string) (string, error) {
				return strings.Join(tokens, ""), nil
			}),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		}, {
			name:  "Seperated1: matching main parser after failing SafeSpot head parser should fail",
			input: "a,1",
			parser: pcb.Map(pcb.Separated0(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1()), pcb.Char(','), false),
				func(tokens []string) (string, error) {
					return strings.Join(tokens, ""), nil
				},
			),
			wantErr:       true,
			wantOutput:    "1",
			wantRemaining: "a,1",
		}, {
			name:          "no matching parser should fail",
			input:         "$%^*",
			parser:        pcb.FirstSuccessful(gomme.SafeSpot(pcb.Digit1()), gomme.SafeSpot(pcb.Alpha1())),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$%^*",
		}, {
			name:          "empty input should fail",
			input:         "",
			parser:        pcb.FirstSuccessful(gomme.SafeSpot(pcb.Digit1()), gomme.SafeSpot(pcb.Alpha1())),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult, gotErr := gomme.RunOnString(tc.input, tc.parser)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkSaveSpot(b *testing.B) {
	p := gomme.SafeSpot(pcb.Char('1'))
	input := gomme.NewFromString("123", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Parse(input)
	}
}
