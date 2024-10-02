package parsify

import (
	"github.com/oleiade/gomme"
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
			basicParser1:   pcb.Char('{'),
			complexParser1: pcb.Delimited(pcb.Char('{'), pcb.UntilString("STOP"), pcb.Char('}')),
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

			if tc.basicParser2 != nil {
				state, firstOutput = tc.basicParser2(gomme.NewFromString(input[:1]))
			} else {
				state, firstOutput = tc.basicParser1.It(gomme.NewFromString(input[:1]))
			}
			t.Log("Error1? :", state.Error())

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

			if tc.basicParser2 != nil {
				newState, gotOutput = tc.complexParser2(gomme.NewFromString(input))
			} else {
				newState, gotOutput = tc.complexParser1.It(gomme.NewFromString(input))
			}
			t.Log("Error2? :", newState.Error())

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
