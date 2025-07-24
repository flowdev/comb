package parsify

import (
	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
	. "github.com/flowdev/comb/cute"
	"testing"
)

func TestDelimitedByChar(t *testing.T) {
	specs := []struct {
		name           string
		basicParser1   comb.Parser[rune]
		complexParser1 comb.Parser[string]
		basicParser2   Parser[rune]
		complexParser2 Parser[string]
	}{
		{
			name:           "normal parser without Parsify",
			basicParser1:   C('{'),
			complexParser1: cmb.Delimited(C('{'), cmb.UntilString("STOP"), C('}')),
		}, {
			name:           "normal parser with Parsify",
			basicParser2:   Char('{'),
			complexParser2: Delimited[rune, string, rune](Char('{'), UntilString("STOP"), Char('}')),
		}, {
			name:           "more generified normal parser with Parsify",
			basicParser2:   Char2('{'),
			complexParser2: Delimited[rune, string, rune](Char2('{'), UntilString("STOP"), Char2('}')),
		}, {
			name:           "literal rune parser with Parsify",
			basicParser2:   Parsify[rune]('{'),
			complexParser2: Delimited[rune, string, rune]('{', UntilString("STOP"), '}'),
		},
	}

	input := "{123abcSTOP}"
	wantOutput := "123abc"
	wantRemaining := ""

	for _, tc := range specs {
		t.Run(tc.name, func(t *testing.T) {
			var state comb.State
			var firstOutput rune
			var firstError *comb.ParserError

			if tc.basicParser2 != nil {
				state, firstOutput, firstError = tc.basicParser2(comb.NewFromString(input[:1], 10))
			} else {
				state, firstOutput, firstError = tc.basicParser1.Parse(-1, comb.NewFromString(input[:1], 10))
			}
			t.Log("Error1? :", firstError)

			if firstOutput != '{' {
				t.Errorf("got output %q, want output %q", firstOutput, '{')
			}

			gotRemaining := state.CurrentString()
			if gotRemaining != "" {
				t.Errorf("got remaining %q, want remaining %q", gotRemaining, "")
			}

			// ---------------------------------------------------------------

			var newState comb.State
			var gotOutput string
			var gotError *comb.ParserError

			if tc.basicParser2 != nil {
				newState, gotOutput, gotError = tc.complexParser2(comb.NewFromString(input, 10))
			} else {
				newState, gotOutput, gotError = tc.complexParser1.Parse(-1, comb.NewFromString(input, 10))
			}
			t.Log("Error2? :", gotError)

			if gotOutput != wantOutput {
				t.Errorf("got output %q, want output %q", gotOutput, wantOutput)
			}

			gotRemaining = newState.CurrentString()
			if gotRemaining != wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", gotRemaining, wantRemaining)
			}
		})
	}
}
