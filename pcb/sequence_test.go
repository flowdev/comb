package pcb

import (
	"github.com/oleiade/gomme"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDelimited(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[string]
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
			input: "+1\r\n",
			args: args{
				p: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:  "no prefix match should fail",
			input: "1\r\n",
			args: args{
				p: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "1\r\n",
		},
		{
			name:  "no parser match should fail",
			input: "+\r\n",
			args: args{
				p: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "+\r\n",
		},
		{
			name:  "no suffix match should fail",
			input: "+1",
			args: args{
				p: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "+1",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Delimited(Char('+'), Digit1(), CRLF()),
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

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkDelimited(b *testing.B) {
	parser := Delimited(Char('+'), Digit1(), CRLF())
	input := gomme.NewFromString("+1\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(input)
	}
}

func TestPreceded(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[string]
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
			input: "+123",
			args: args{
				p: Preceded(Char('+'), Digit1()),
			},
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:  "no prefix match should fail",
			input: "+123",
			args: args{
				p: Preceded(Char('-'), Digit1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "+123",
		},
		{
			name:  "no parser match should succeed",
			input: "+",
			args: args{
				p: Preceded(Char('+'), Digit1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "+",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Preceded(Char('+'), Digit1()),
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

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkPreceded(b *testing.B) {
	parser := Preceded(Char('+'), Digit1())
	input := gomme.NewFromString("+123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(input)
	}
}

func TestSequence(t *testing.T) {
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
			name:  "matching parsers should succeed",
			input: "1a3",
			args: args{
				p: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       false,
			wantOutput:    []string{"1", "a", "3"},
			wantRemaining: "",
		},
		{
			name:  "matching parsers in longer input should succeed",
			input: "1a3bcd",
			args: args{
				p: Sequence(Digit1(), Alpha0(), Digit1()),
			},
			wantErr:       false,
			wantOutput:    []string{"1", "a", "3"},
			wantRemaining: "bcd",
		},
		{
			name:  "partially matching parsers should fail",
			input: "1a3",
			args: args{
				p: Sequence(Digit1(), Digit1(), Digit1()),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "1a3",
		},
		{
			name:  "too short input should fail",
			input: "12",
			args: args{
				p: Sequence(Digit1(), Digit1(), Digit1()),
			},
			wantErr:       true,
			wantOutput:    nil,
			wantRemaining: "12",
		},
		{
			name:  "empty input should succeed",
			input: "",
			args: args{
				p: Sequence(Digit1(), Digit1(), Digit1()),
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

func BenchmarkSequence(b *testing.B) {
	parser := Sequence(Digit1(), Alpha0(), Digit1())
	input := gomme.NewFromString("123A45")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(input)
	}
}

func TestTerminated(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[string]
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
			input: "1+23",
			args: args{
				p: Terminated(Digit1(), Char('+')),
			},
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "23",
		},
		{
			name:  "no suffix match should fail",
			input: "1-23",
			args: args{
				p: Terminated(Digit1(), Char('+')),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "1-23",
		},
		{
			name:  "no parser match should succeed",
			input: "+",
			args: args{
				p: Terminated(Digit1(), Char('+')),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "+",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Terminated(Digit1(), Char('+')),
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

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkTerminated(b *testing.B) {
	parser := Terminated(Digit1(), Char('+'))
	input := gomme.NewFromString("123+")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.It(input)
	}
}
