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
			wantOutput:    []string{},
			wantRemaining: "abc123",
		},
		{
			name:          "parsing no count should fail",
			parser:        Count(String("abc"), 2),
			input:         "123123",
			wantErr:       true,
			wantOutput:    []string{},
			wantRemaining: "123123",
		},
		{
			name:          "parsing empty input should fail",
			parser:        Count(String("abc"), 2),
			input:         "",
			wantErr:       true,
			wantOutput:    []string{},
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := tc.parser.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			assert.Equal(t,
				tc.wantOutput,
				gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkCount(b *testing.B) {
	parser := Count(Char('#'), 3)
	state := gomme.NewFromString("###")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(state)
	}
}

func TestMany0(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[[]rune]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    []rune
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "###",
			args: args{
				p: Many0(Char('#')),
			},
			wantErr:       false,
			wantOutput:    []rune{'#', '#', '#'},
			wantRemaining: "",
		},
		{
			name:  "no match should succeed",
			input: "abc",
			args: args{
				p: Many0(Char('#')),
			},
			wantErr:       false,
			wantOutput:    []rune{},
			wantRemaining: "abc",
		},
		{
			name:  "empty input should succeed",
			input: "",
			args: args{
				p: Many0(Char('#')),
			},
			wantErr:       false,
			wantOutput:    []rune{},
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func TestMany0DetectsInfiniteLoops(t *testing.T) {
	t.Parallel()

	// Digit0 accepts empty state, and would cause an infinite loop if not detected
	state := gomme.NewFromString("abcdef")
	parser := Many0(Digit0())

	newState, output := parser.It(state)

	assert.Error(t, newState)
	assert.Empty(t, output)
	assert.Equal(t, state.CurrentString(), newState.CurrentString())
}

func BenchmarkMany0(b *testing.B) {
	parser := Many0(Char('#'))
	state := gomme.NewFromString("###")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(state)
	}
}

func TestMany1(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[[]rune]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    []rune
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "###",
			args: args{
				p: Many1(Char('#')),
			},
			wantErr:       false,
			wantOutput:    []rune{'#', '#', '#'},
			wantRemaining: "",
		},
		{
			name:  "matching at least once should succeed",
			input: "#abc",
			args: args{
				p: Many1(Char('#')),
			},
			wantErr:       false,
			wantOutput:    []rune{'#'},
			wantRemaining: "abc",
		},
		{
			name:  "not matching at least once should fail",
			input: "a##",
			args: args{
				p: Many1(Char('#')),
			},
			wantErr:       true,
			wantOutput:    []rune{},
			wantRemaining: "a##",
		},
		{
			name:  "no match should fail",
			input: "abc",
			args: args{
				p: Many1(Char('#')),
			},
			wantErr:       true,
			wantOutput:    []rune{},
			wantRemaining: "abc",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Many1(Char('#')),
			},
			wantErr:       true,
			wantOutput:    []rune{},
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func TestMany1DetectsInfiniteLoops(t *testing.T) {
	t.Parallel()

	// Digit0 accepts empty state, and would cause an infinite loop if not detected
	state := gomme.NewFromString("abcdef")
	parser := Many1(Digit0())

	newState, output := parser.It(state)

	assert.Error(t, newState)
	assert.Empty(t, output)
	assert.Equal(t, state.CurrentString(), newState.CurrentString())
}

func BenchmarkMany1(b *testing.B) {
	parser := Many1(Char('#'))
	state := gomme.NewFromString("###")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(state)
	}
}

func TestSeparated0(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[[]string]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    []string
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "abc,abc,abc",
			args: args{
				p: Separated0(String("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc", "abc", "abc"},
			wantRemaining: "",
		},
		{
			name:  "matching parser and missing separator should succeed",
			input: "abc123abc",
			args: args{
				p: Separated0(String("abc"), Char(','), true),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: "123abc",
		},
		{
			name:  "parser with separator but non-matching right side should succeed",
			input: "abc,def",
			args: args{
				p: Separated0(String("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: ",def",
		},
		{
			name:  "parser matching on the right of the separator should succeed",
			input: "def,abc",
			args: args{
				p: Separated0(String("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{},
			wantRemaining: "def,abc",
		},
		{
			name:  "empty input should succeed",
			input: "",
			args: args{
				p: Separated0(String("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{},
			wantRemaining: "",
		},
		{
			name:  "parsing input without separator should succeed",
			input: "123",
			args: args{
				p: Separated0(Digit0(), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"123"},
			wantRemaining: "",
		},
		{
			name:  "parsing empty input with *0 parser should succeed",
			input: "",
			args: args{
				p: Separated0(Digit0(), Char(','), true),
			},
			wantErr:       false,
			wantOutput:    []string{},
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkSeparated0(t *testing.B) {
	parser := Separated0(Char('#'), Char(','), false)
	state := gomme.NewFromString("#,#,#")

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		_, _ = parser.It(state)
	}
}

func TestSeparated1(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[[]string]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    []string
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "abc,abc,abc",
			args: args{
				p: Separated1(String("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc", "abc", "abc"},
			wantRemaining: "",
		},
		{
			name:  "matching parser and missing separator should succeed",
			input: "abc123abc",
			args: args{
				p: Separated1(String("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: "123abc",
		},
		{
			name:  "parser with separator but non-matching right side should succeed",
			input: "abc,def",
			args: args{
				p: Separated1(String("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: ",def",
		},
		{
			name:  "parser matching on the right of the separator should fail",
			input: "def,abc",
			args: args{
				p: Separated1(String("abc"), Char(','), false),
			},
			wantErr:       true,
			wantOutput:    []string{},
			wantRemaining: "def,abc",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Separated1(String("abc"), Char(','), false),
			},
			wantErr:       true,
			wantOutput:    []string{},
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult,
				"got output %v, want output %v", gotResult, tc.wantOutput,
			)

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkSeparated1(t *testing.B) {
	parser := Separated1(Char('#'), Char(','), false)
	state := gomme.NewFromString("#,#,#,#")

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		_, _ = parser.It(state)
	}
}
