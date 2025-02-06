package pcb

import (
	"github.com/oleiade/gomme"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[[]string]
		input         string
		wantErr       bool
		wantOutput    []string
		wantRemaining string
	}{
		{
			name:          "parsing exact count should succeed",
			parser:        Count(String("abc"), 2),
			input:         "abcabc",
			wantErr:       false,
			wantOutput:    []string{"abc", "abc"},
			wantRemaining: "",
		},
		{
			name:          "parsing more than count should succeed",
			parser:        Count(String("abc"), 2),
			input:         "abcabcabc",
			wantErr:       false,
			wantOutput:    []string{"abc", "abc"},
			wantRemaining: "abc",
		},
		{
			name:          "parsing less than count should fail",
			parser:        Count(String("abc"), 2),
			input:         "abc123",
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "abc123",
		},
		{
			name:          "parsing no count should fail",
			parser:        Count(String("abc"), 2),
			input:         "123123",
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "123123",
		},
		{
			name:          "parsing empty input should fail",
			parser:        Count(String("abc"), 2),
			input:         "",
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "",
		},
	}

	gomme.SetDebug(true)
	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult, gotErr := gomme.RunOnString(tc.input, tc.parser)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			assert.Equal(t,
				tc.wantOutput,
				gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)
		})
	}
}

func BenchmarkCount(b *testing.B) {
	parser := Count(Char('#'), 3)
	state := gomme.NewFromString("###", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(state)
	}
}

func TestMany0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[[]rune]
		input         string
		wantErr       bool
		wantOutput    []rune
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "###",
			parser:        Many0(Char('#')),
			wantErr:       false,
			wantOutput:    []rune{'#', '#', '#'},
			wantRemaining: "",
		},
		{
			name:          "no match should succeed",
			input:         "abc",
			parser:        Many0(Char('#')),
			wantErr:       false,
			wantOutput:    []rune{},
			wantRemaining: "abc",
		},
		{
			name:          "empty input should succeed",
			input:         "",
			parser:        Many0(Char('#')),
			wantErr:       false,
			wantOutput:    []rune{},
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

			// testify makes it easier to compare slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)
		})
	}
}

func TestMany0DetectsInfiniteLoops(t *testing.T) {
	t.Parallel()

	// Digit0 accepts empty state, and would cause an infinite loop if not detected
	state := gomme.NewFromString("abcdef", true)
	parser := Many0(Digit0())

	newState, output, err := parser.Parse(state)

	assert.Error(t, err)
	assert.Equal(t, []string{""}, output)
	assert.Equal(t, state.CurrentString(), newState.CurrentString())
}

func BenchmarkMany0(b *testing.B) {
	parser := Many0(Char('#'))
	state := gomme.NewFromString("###", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(state)
	}
}

func TestMany1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[[]rune]
		input         string
		wantErr       bool
		wantOutput    []rune
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "###",
			parser:        Many1(Char('#')),
			wantErr:       false,
			wantOutput:    []rune{'#', '#', '#'},
			wantRemaining: "",
		},
		{
			name:          "matching at least once should succeed",
			input:         "#abc",
			parser:        Many1(Char('#')),
			wantErr:       false,
			wantOutput:    []rune{'#'},
			wantRemaining: "abc",
		},
		{
			name:          "not matching at least once should fail",
			input:         "a##",
			parser:        Many1(Char('#')),
			wantErr:       true,
			wantOutput:    []rune{'#', '#'},
			wantRemaining: "a##",
		},
		{
			name:          "no match should fail",
			input:         "abc",
			parser:        Many1(Char('#')),
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "abc",
		},
		{
			name:          "empty input should fail",
			input:         "",
			parser:        Many1(Char('#')),
			wantErr:       true,
			wantOutput:    nil,
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

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)
		})
	}
}

func TestMany1DetectsInfiniteLoops(t *testing.T) {
	t.Parallel()

	// Digit0 accepts empty state, and would cause an infinite loop if not detected
	state := gomme.NewFromString("abcdef", true)
	parser := Many1(Digit0())

	newState, output, err := parser.Parse(state)

	assert.Error(t, err)
	assert.Equal(t, []string{""}, output)
	assert.Equal(t, state.CurrentString(), newState.CurrentString())
}

func BenchmarkMany1(b *testing.B) {
	parser := Many1(Char('#'))
	state := gomme.NewFromString("###", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(state)
	}
}

func TestSeparated0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[[]string]
		input         string
		wantErr       bool
		wantOutput    []string
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "abc,abc,abc",
			parser:        Separated0(String("abc"), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{"abc", "abc", "abc"},
			wantRemaining: "",
		}, {
			name:          "matching parser and missing separator should succeed",
			input:         "abc123abc",
			parser:        Separated0(String("abc"), Char(','), true),
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: "123abc",
		}, {
			name:          "parser with separator but non-matching right side should succeed",
			input:         "abc,def",
			parser:        Separated0(String("abc"), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: ",def",
		}, {
			name:          "parser matching on the right of the separator should succeed",
			input:         "def,abc",
			parser:        Separated0(String("abc"), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{},
			wantRemaining: "def,abc",
		}, {
			name:          "empty input should succeed",
			input:         "",
			parser:        Separated0(String("abc"), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{},
			wantRemaining: "",
		}, {
			name:          "parsing input without separator should succeed",
			input:         "123",
			parser:        Separated0(Digit0(), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{"123"},
			wantRemaining: "",
		}, {
			name:          "parsing empty input with *0 parser should succeed",
			input:         "",
			parser:        Separated0(Digit1(), Char(','), true),
			wantErr:       false,
			wantOutput:    []string{},
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

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)
		})
	}
}

func BenchmarkSeparated0(t *testing.B) {
	parser := Separated0(Char('#'), Char(','), false)
	state := gomme.NewFromString("#,#,#", false)

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		_, _, _ = parser.Parse(state)
	}
}

func TestSeparated1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[[]string]
		input         string
		wantErr       bool
		wantOutput    []string
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "abc,abc,abc",
			parser:        Separated1(String("abc"), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{"abc", "abc", "abc"},
			wantRemaining: "",
		}, {
			name:          "matching parser and missing separator should succeed",
			input:         "abc123abc",
			parser:        Separated1(String("abc"), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: "123abc",
		}, {
			name:          "parser with separator but non-matching right side should succeed",
			input:         "abc,def",
			parser:        Separated1(String("abc"), Char(','), false),
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: ",def",
		}, {
			name:          "parser matching on the right of the separator should fail",
			input:         "def,abc",
			parser:        Separated1(String("abc"), Char(','), false),
			wantErr:       true,
			wantOutput:    []string{"abc"}, // one value after deleting 2 tokens
			wantRemaining: "",
		}, {
			name:          "empty input should fail",
			input:         "",
			parser:        Separated1(String("abc"), Char(','), false),
			wantErr:       true,
			wantOutput:    nil,
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

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)
		})
	}
}

func BenchmarkSeparated1(t *testing.B) {
	parser := Separated1(Char('#'), Char(','), false)
	state := gomme.NewFromString("#,#,#,#", false)

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		_, _, _ = parser.Parse(state)
	}
}
