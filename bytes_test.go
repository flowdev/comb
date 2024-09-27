package gomme

import (
	"fmt"
	"testing"
	"unicode"
)

func TestTake(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[[]byte]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:  "taking less than input size should succeed",
			input: "1234567",
			args: args{
				p: Take(6),
			},
			wantErr:       false,
			wantOutput:    "123456",
			wantRemaining: "7",
		},
		{
			name:  "taking exact input size should succeed",
			input: "123456",
			args: args{
				p: Take(6),
			},
			wantErr:       false,
			wantOutput:    "123456",
			wantRemaining: "",
		},
		{
			name:  "taking more than input size should fail",
			input: "123",
			args: args{
				p: Take(6),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:  "taking from empty input should fail",
			input: "",
			args: args{
				p: Take(6),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := NewInputFromString(tc.input)
			gotResult := tc.args.p(input)
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			gotOutput := string(gotResult.Output)
			if gotOutput != tc.wantOutput {
				t.Errorf("got output %v, want output %v", gotResult.Output, tc.wantOutput)
			}

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %v, want remaining %v", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkTake(b *testing.B) {
	p := Take(6)
	input := NewInputFromString("123456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

func TestTakeUntil(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[[]byte]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "abc123",
			args: args{
				p: TakeUntil(Digit1()),
			},
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:  "immediately matching parser should succeed",
			input: "123",
			args: args{
				p: TakeUntil(Digit1()),
			},
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "no match should fail",
			input: "abcdef",
			args: args{
				p: TakeUntil(Digit1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abcdef",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: TakeUntil(Digit1()),
			},
			wantErr:       true,
			wantOutput:    "",
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

			if string(gotResult.Output) != tc.wantOutput {
				t.Errorf("got output %v, want output %v", string(gotResult.Output), tc.wantOutput)
			}

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkTakeUntil(b *testing.B) {
	p := TakeUntil(Digit1())
	input := NewInputFromString("abc123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

func TestTakeWhileMN(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[[]byte]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:  "parsing input with enough characters and partially matching predicated should succeed",
			input: "latin123",
			args: args{
				p: SatisfyMN(3, 6, unicode.IsLetter),
			},
			wantErr:       false,
			wantOutput:    "latin",
			wantRemaining: "123",
		},
		{
			name:  "parsing input longer than atLeast and atMost should succeed",
			input: "lengthy",
			args: args{
				p: SatisfyMN(3, 6, unicode.IsLetter),
			},
			wantErr:       false,
			wantOutput:    "length",
			wantRemaining: "y",
		},
		{
			name:  "parsing input longer than atLeast and shorter than atMost should succeed",
			input: "latin",
			args: args{
				p: SatisfyMN(3, 6, unicode.IsLetter),
			},
			wantErr:       false,
			wantOutput:    "latin",
			wantRemaining: "",
		},
		{
			name:  "parsing empty input should fail",
			input: "",
			args: args{
				p: SatisfyMN(3, 6, unicode.IsLetter),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "parsing too short input should fail",
			input: "ed",
			args: args{
				p: SatisfyMN(3, 6, unicode.IsLetter),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "ed",
		},
		{
			name:  "parsing with non-matching predicate should fail",
			input: "12345",
			args: args{
				p: SatisfyMN(3, 6, unicode.IsLetter),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "12345",
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

			if string(gotResult.Output) != tc.wantOutput {
				t.Errorf("got output %q, want output %q", string(gotResult.Output), tc.wantOutput)
			}

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkTakeWhileMN(b *testing.B) {
	p := SatisfyMN(3, 6, IsDigit)
	input := NewInputFromString("13579")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

// TakeWhileOneOf parses any number of characters present in the
// provided collection of runes.
func TakeWhileOneOf(collection ...rune) Parser[string] {
	index := make(map[rune]struct{}, len(collection))

	for _, r := range collection {
		index[r] = struct{}{}
	}

	expected := fmt.Sprintf("chars(%v)", string(collection))

	return func(input State) (State, string) {
		if input.AtEnd() {
			return Failure[string](NewError(input, expected), input)
		}

		pos := 0
		var r rune
		for pos, r = range input.CurrentString() {
			_, exists := index[r]
			if !exists {
				if pos == 0 {
					return Failure[string](NewError(input, expected), input)
				}

				break
			}
		}

		nextInput := input.MoveBy(uint(pos))
		return Success(input.StringTo(nextInput), nextInput)
	}
}

func TestTakeWhileOneOf(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[string]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "abc123",
			args: args{
				p: TakeWhileOneOf('a', 'b', 'c'),
			},
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "123",
		},
		{
			name:  "no match should fail",
			input: "123",
			args: args{
				p: TakeWhileOneOf('a', 'b', 'c'),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: TakeWhileOneOf('a', 'b', 'c'),
			},
			wantErr:       true,
			wantOutput:    "",
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

			if gotResult.Output != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult.Output, tc.wantOutput)
			}

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkTakeWhileOneOf(b *testing.B) {
	p := TakeWhileOneOf('a', 'b', 'c')
	input := NewInputFromString("abc123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}
