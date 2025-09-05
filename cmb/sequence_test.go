package cmb

import (
	"slices"
	"testing"

	"github.com/flowdev/comb"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		input      string
		parser     comb.Parser[[]string]
		wantErr    bool
		wantOutput []string
	}{
		{
			name:       "matching parsers should succeed",
			input:      "123abc",
			parser:     Sequence(Digit1(), Alpha1()),
			wantErr:    false,
			wantOutput: []string{"123", "abc"},
		}, {
			name:       "head matching parser should fail",
			input:      "123;",
			parser:     Sequence(Digit1(), Alpha1()),
			wantErr:    true,
			wantOutput: []string{"123", ""},
		}, {
			name:       "tail matching parser should fail",
			input:      "abc",
			parser:     Sequence(comb.SafeSpot(Digit1()), comb.SafeSpot(Alpha1())),
			wantErr:    true,
			wantOutput: []string{"", "abc"},
		}, {
			name:       "no matching parser should fail",
			input:      "$%^*",
			parser:     Sequence(Digit1(), Alpha1()),
			wantErr:    true,
			wantOutput: []string{"", ""},
		}, {
			name:       "empty input should fail",
			input:      "",
			parser:     Sequence(Digit1(), Alpha1()),
			wantErr:    true,
			wantOutput: []string{"", ""},
		},
	}
	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotOutput, gotErr := comb.RunOnString(tc.input, tc.parser)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			if !slices.Equal(gotOutput, tc.wantOutput) {
				t.Errorf("got output %q, want %q", gotOutput, tc.wantOutput)
			}
		})
	}
}

func BenchmarkSequence(b *testing.B) {
	p := Sequence(Char('a'), Char('b'))
	input := comb.NewFromString("abc", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Parse(input)
	}
}
