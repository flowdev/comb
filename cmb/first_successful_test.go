package cmb

import (
	"testing"

	"github.com/flowdev/comb"
)

func TestFirstSuccessful(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		input      string
		parser     comb.Parser[string]
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "head matching parser should succeed",
			input:      "123",
			parser:     FirstSuccessful(Digit1(), Alpha0()),
			wantErr:    false,
			wantOutput: "123",
		}, {
			name:       "tail matching parser should succeed",
			input:      "abc",
			parser:     FirstSuccessful(Digit1(), Alpha0()),
			wantErr:    false,
			wantOutput: "abc",
		}, {
			name:       "no matching parser should fail",
			input:      "$%^*",
			parser:     FirstSuccessful(Digit1(), Alpha1()),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "empty input should fail",
			input:      "",
			parser:     FirstSuccessful(Digit1(), Alpha1()),
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

func BenchmarkFirstSuccessful(b *testing.B) {
	p := FirstSuccessful(Char('b'), Char('a'))
	input := comb.NewFromString("abc", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Parse(input)
	}
}
