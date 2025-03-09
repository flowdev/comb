package comb_test

import (
	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
	"strings"
	"testing"
)

func TestSaveSpot(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[string]
		input      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "head matching parser should succeed",
			input:      "123",
			parser:     cmb.FirstSuccessful(cmb.Digit1(), comb.SafeSpot(cmb.Alpha1())),
			wantErr:    false,
			wantOutput: "123",
		}, {
			name:       "tail matching parser should succeed",
			input:      "abc",
			parser:     cmb.FirstSuccessful(comb.SafeSpot(cmb.Digit1()), cmb.Alpha1()),
			wantErr:    false,
			wantOutput: "abc",
		}, {
			name:       "FirstSuccessful: tail matching parser after failing SafeSpot head parser should fail",
			input:      "abc",
			parser:     cmb.FirstSuccessful(cmb.Prefixed(comb.SafeSpot(cmb.String("a")), cmb.Digit1()), cmb.Alpha1()),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "Optional: tail matching parser after failing SafeSpot head parser should fail",
			input:      "abc",
			parser:     cmb.Optional(cmb.Prefixed(comb.SafeSpot(cmb.String("a")), cmb.Digit1())),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:  "Many0: tail matching parser after failing SafeSpot head parser should fail",
			input: "abc",
			parser: cmb.Map(cmb.Many0(cmb.Prefixed(comb.SafeSpot(cmb.String("a")), cmb.Digit1())), func(tokens []string) (string, error) {
				return strings.Join(tokens, ""), nil
			}),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:  "Seperated1: matching main parser after failing SafeSpot head parser should fail",
			input: "a,1",
			parser: cmb.Map(cmb.Separated0(cmb.Prefixed(comb.SafeSpot(cmb.String("a")), cmb.Digit1()), cmb.Char(','), false),
				func(tokens []string) (string, error) {
					return strings.Join(tokens, ""), nil
				},
			),
			wantErr:    true,
			wantOutput: "1",
		}, {
			name:       "no matching parser should fail",
			input:      "$%^*",
			parser:     cmb.FirstSuccessful(comb.SafeSpot(cmb.Digit1()), comb.SafeSpot(cmb.Alpha1())),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "empty input should fail",
			input:      "",
			parser:     cmb.FirstSuccessful(comb.SafeSpot(cmb.Digit1()), comb.SafeSpot(cmb.Alpha1())),
			wantErr:    true,
			wantOutput: "",
		},
	}
	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult, gotErr := comb.RunOnString(tc.input, tc.parser)
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
	p := comb.SafeSpot(cmb.Char('1'))
	input := comb.NewFromString("123", false, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Parse(input)
	}
}
