package gomme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[[]string]
		input         string
		wantErr       bool
		wantOutput    []string
		wantRemaining string
	}{
		{
			name:          "parsing exact count should succeed",
			parser:        Count(Token("abc"), 2),
			input:         "abcabc",
			wantErr:       false,
			wantOutput:    []string{"abc", "abc"},
			wantRemaining: "",
		},
		{
			name:          "parsing more than count should succeed",
			parser:        Count(Token("abc"), 2),
			input:         "abcabcabc",
			wantErr:       false,
			wantOutput:    []string{"abc", "abc"},
			wantRemaining: "abc",
		},
		{
			name:          "parsing less than count should fail",
			parser:        Count(Token("abc"), 2),
			input:         "abc123",
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "abc123",
		},
		{
			name:          "parsing no count should fail",
			parser:        Count(Token("abc"), 2),
			input:         "123123",
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "123123",
		},
		{
			name:          "parsing empty input should fail",
			parser:        Count(Token("abc"), 2),
			input:         "",
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult := tc.parser(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			assert.Equal(t,
				tc.wantOutput,
				gotResult.Output,
				"got output %v, want output %v", gotResult.Output, tc.wantOutput,
			)

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkCount(b *testing.B) {
	parser := Count(Char('#'), 3)
	input := NewInputFromString("###")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestMany0(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[[]rune]
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

			gotResult := tc.args.p(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult.Output,
				"got output %v, want output %v", gotResult.Output, tc.wantOutput,
			)

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func TestMany0DetectsInfiniteLoops(t *testing.T) {
	t.Parallel()

	// Digit0 accepts empty input, and would cause an infinite loop if not detected
	input := NewInputFromString("abcdef")
	parser := Many0(Digit0())

	result := parser(input)

	assert.Error(t, result.Err)
	assert.Nil(t, result.Output)
	assert.Equal(t, input, result.Remaining)
}

func BenchmarkMany0(b *testing.B) {
	parser := Many0(Char('#'))
	input := NewInputFromString("###")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestMany1(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[[]rune]
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
			wantOutput:    nil,
			wantRemaining: "a##",
		},
		{
			name:  "no match should fail",
			input: "abc",
			args: args{
				p: Many1(Char('#')),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "abc",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Many1(Char('#')),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult := tc.args.p(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult.Output,
				"got output %v, want output %v", gotResult.Output, tc.wantOutput,
			)

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func TestMany1DetectsInfiniteLoops(t *testing.T) {
	t.Parallel()

	// Digit0 accepts empty input, and would cause an infinite loop if not detected
	input := NewInputFromString("abcdef")
	parser := Many1(Digit0())

	result := parser(input)

	assert.Error(t, result.Err)
	assert.Nil(t, result.Output)
	assert.Equal(t, input, result.Remaining)
}

func BenchmarkMany1(b *testing.B) {
	parser := Many1(Char('#'))
	input := NewInputFromString("###")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestSeparatedList0(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[[]string]
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
				p: SeparatedList0(Token("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc", "abc", "abc"},
			wantRemaining: "",
		},
		{
			name:  "matching parser and missing separator should succeed",
			input: "abc123abc",
			args: args{
				p: SeparatedList0(Token("abc"), Char(','), true),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: "123abc",
		},
		{
			name:  "parser with separator but non-matching right side should succeed",
			input: "abc,def",
			args: args{
				p: SeparatedList0(Token("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: ",def",
		},
		{
			name:  "parser matching on the right of the separator should succeed",
			input: "def,abc",
			args: args{
				p: SeparatedList0(Token("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{},
			wantRemaining: "def,abc",
		},
		{
			name:  "empty input should succeed",
			input: "",
			args: args{
				p: SeparatedList0(Token("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{},
			wantRemaining: "",
		},
		{
			name:  "parsing input without separator should succeed",
			input: "123",
			args: args{
				p: SeparatedList0(Digit0(), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"123"},
			wantRemaining: "",
		},
		{
			name:  "using a parser accepting empty input should fail",
			input: "",
			args: args{
				p: SeparatedList0(Digit0(), Char(','), true),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult := tc.args.p(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult.Output,
				"got output %v, want output %v", gotResult.Output, tc.wantOutput,
			)

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkSeparatedList0(t *testing.B) {
	parser := SeparatedList0(Char('#'), Char(','), false)
	input := NewInputFromString("#,#,#")

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		parser(input)
	}
}

func TestSeparatedList1(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[[]string]
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
				p: SeparatedList1(Token("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc", "abc", "abc"},
			wantRemaining: "",
		},
		{
			name:  "matching parser and missing separator should succeed",
			input: "abc123abc",
			args: args{
				p: SeparatedList1(Token("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: "123abc",
		},
		{
			name:  "parser with separator but non-matching right side should succeed",
			input: "abc,def",
			args: args{
				p: SeparatedList1(Token("abc"), Char(','), false),
			},
			wantErr:       false,
			wantOutput:    []string{"abc"},
			wantRemaining: ",def",
		},
		{
			name:  "parser matching on the right of the separator should fail",
			input: "def,abc",
			args: args{
				p: SeparatedList1(Token("abc"), Char(','), false),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "def,abc",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: SeparatedList1(Token("abc"), Char(','), false),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotResult := tc.args.p(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			// testify makes it easier comparing slices
			assert.Equal(t,
				tc.wantOutput, gotResult.Output,
				"got output %v, want output %v", gotResult.Output, tc.wantOutput,
			)

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkSeparatedList1(t *testing.B) {
	parser := SeparatedList1(Char('#'), Char(','), false)
	input := NewInputFromString("#,#,#,#")

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		parser(input)
	}
}
