package parsify

import (
	"github.com/oleiade/gomme"
	"github.com/oleiade/gomme/pcb"
	"testing"
)

func TestDelimitedByChar(t *testing.T) {
	specs := []struct {
		name          string
		basicParser   gomme.Parser[rune]
		complexParser gomme.Parser[string]
	}{
		{
			name:          "normal parser without Parsify",
			basicParser:   pcb.Char('{'),
			complexParser: pcb.Delimited(pcb.Char('{'), pcb.UntilString("STOP"), pcb.Char('}')),
		}, {
			name:          "normal parser with Parsify",
			basicParser:   Char('{'),
			complexParser: Delimited[rune, string, rune](Char('{'), UntilString("STOP"), Char('}')),
		}, {
			name:          "more generified normal parser with Parsify",
			basicParser:   Char2('{'),
			complexParser: Delimited[rune, string, rune](Char2('{'), UntilString("STOP"), Char2('}')),
		}, {
			name:          "literal rune parser with Parsify",
			basicParser:   Parsify[rune]('{'),
			complexParser: Delimited[rune, string, rune]('{', UntilString("STOP"), '}'),
		},
	}

	input := "{123abcSTOP}"
	wantOutput := "123abc"
	wantRemaining := ""

	for _, tc := range specs {
		t.Run(tc.name, func(t *testing.T) {
			state, firstOutput := tc.basicParser(gomme.NewFromString(input[:1]))
			t.Log("Error1? :", state.Error())

			if firstOutput != '{' {
				t.Errorf("got output %q, want output %q", firstOutput, '{')
			}

			gotRemaining := state.CurrentString()
			if gotRemaining != "" {
				t.Errorf("got remaining %q, want remaining %q", gotRemaining, "")
			}

			// ---------------------------------------------------------------

			newState, gotOutput := tc.complexParser(gomme.NewFromString(input))
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
