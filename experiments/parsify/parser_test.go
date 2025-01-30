package parsify

import (
	"github.com/oleiade/gomme"
	. "github.com/oleiade/gomme/cute"
	"github.com/oleiade/gomme/pcb"
	"testing"
)

func TestDelimitedByChar(t *testing.T) {
	specs := []struct {
		name           string
		basicParser1   gomme.Parser[rune]
		complexParser1 gomme.Parser[string]
		basicParser2   Parser[rune]
		complexParser2 Parser[string]
	}{
		{
			name:           "normal parser without Parsify",
			basicParser1:   C('{'),
			complexParser1: pcb.Delimited(C('{'), pcb.UntilString("STOP"), C('}')),
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
			var state gomme.State
			var firstOutput rune
			var firstError *gomme.ParserError

			if tc.basicParser2 != nil {
				state, firstOutput, firstError = tc.basicParser2(gomme.NewFromString(input[:1], true))
			} else {
				state, firstOutput, firstError = tc.basicParser1.Parse(gomme.NewFromString(input[:1], true))
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

			var newState gomme.State
			var gotOutput string
			var gotError *gomme.ParserError

			if tc.basicParser2 != nil {
				state, gotOutput, gotError = tc.complexParser2(gomme.NewFromString(input, true))
			} else {
				state, gotOutput, gotError = tc.complexParser1.Parse(gomme.NewFromString(input, true))
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
