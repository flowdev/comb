package pcb

import (
	"errors"
	"github.com/oleiade/gomme"
	"strconv"
	"testing"
)

func TestOptional(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[string]
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
			input: "\r\n123",
			args: args{
				parser: Optional(CRLF()),
			},
			wantErr:       false,
			wantOutput:    "\r\n",
			wantRemaining: "123",
		},
		{
			name:  "no match should succeed",
			input: "123",
			args: args{
				parser: Optional(CRLF()),
			},
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:  "empty input should succeed",
			input: "",
			args: args{
				parser: Optional(CRLF()),
			},
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
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

func BenchmarkOptional(b *testing.B) {
	parser := Optional(CRLF())
	input := gomme.NewFromString(1, nil, -1, "\r\n123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestPeek(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[string]
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
			input: "abcd;",
			args: args{
				parser: Peek(Alpha1()),
			},
			wantErr:       false,
			wantOutput:    "abcd",
			wantRemaining: "abcd;",
		},
		{
			name:  "non matching parser should fail",
			input: "123;",
			args: args{
				parser: Peek(Alpha1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123;",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
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

func BenchmarkPeek(b *testing.B) {
	parser := Peek(Alpha1())
	input := gomme.NewFromString(1, nil, -1, "abcd;")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestRecognize(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[[]byte]
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
			input: "123abc",
			args: args{
				parser: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)),
			},
			wantErr:       false,
			wantOutput:    "123abc",
			wantRemaining: "",
		},
		{
			name:  "no prefix match should fail",
			input: "abc",
			args: args{
				parser: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)),
			},
			wantErr:       true,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:  "no postfix match should fail",
			input: "123",
			args: args{
				parser: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				parser: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)),
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

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if string(gotResult) != tc.wantOutput {
				t.Errorf("got output %q, want output %q", string(gotResult), tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkRecognize(b *testing.B) {
	parser := Recognize(Map2(Digit1(), Alpha1(), pairMapFunc))
	input := gomme.NewFromString(1, nil, -1, "123abc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestAssign(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[int]
	}
	testCases := []struct {
		name          string
		args          args
		input         string
		wantErr       bool
		wantOutput    int
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "abcd",
			args: args{
				parser: Assign(1234, Alpha1()),
			},
			wantErr:       false,
			wantOutput:    1234,
			wantRemaining: "",
		},
		{
			name:  "non matching parser should fail",
			input: "123abcd;",
			args: args{
				parser: Assign(1234, Alpha1()),
			},
			wantErr:       true,
			wantOutput:    1234,
			wantRemaining: "123abcd;",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %d, want output %d", gotResult, tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkAssign(b *testing.B) {
	parser := Assign(1234, Alpha1())
	input := gomme.NewFromString(1, nil, -1, "abcd")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestDelimited(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[string]
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
				parser: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:  "no prefix match should fail",
			input: "1\r\n",
			args: args{
				parser: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       true,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:  "no parser match should fail",
			input: "+\r\n",
			args: args{
				parser: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "no suffix match should fail",
			input: "+1",
			args: args{
				parser: Delimited(Char('+'), Digit1(), CRLF()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				parser: Delimited(Char('+'), Digit1(), CRLF()),
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

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
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
	input := gomme.NewFromString(1, nil, -1, "+1\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestPrefixed(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[string]
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
				parser: Prefixed(Char('+'), Digit1()),
			},
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:  "no prefix match should fail",
			input: "+123",
			args: args{
				parser: Prefixed(Char('-'), Digit1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:  "no parser match should fail",
			input: "+",
			args: args{
				parser: Prefixed(Char('+'), Digit1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				parser: Prefixed(Char('+'), Digit1()),
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

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
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

func BenchmarkPrefixed(b *testing.B) {
	parser := Prefixed(Char('+'), Digit1())
	input := gomme.NewFromString(1, nil, -1, "+123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestSuffixed(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[string]
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
				parser: Suffixed(Digit1(), Char('+')),
			},
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "23",
		},
		{
			name:  "no suffix match should fail",
			input: "1-23",
			args: args{
				parser: Suffixed(Digit1(), Char('+')),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "-23",
		},
		{
			name:  "no parser match should fail",
			input: "+",
			args: args{
				parser: Suffixed(Digit1(), Char('+')),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				parser: Suffixed(Digit1(), Char('+')),
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

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
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
	parser := Suffixed(Digit1(), Char('+'))
	input := gomme.NewFromString(1, nil, -1, "123+")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	type args struct {
		parser gomme.Parser[int]
	}
	testCases := []struct {
		name          string
		input         string
		args          args
		wantErr       bool
		wantOutput    int
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "1abc\r\n",
			args: args{
				Map(Digit1(), func(digit string) (int, error) {
					i, _ := strconv.Atoi(digit)
					return i, nil
				}),
			},
			wantErr:       false,
			wantOutput:    1,
			wantRemaining: "abc\r\n",
		},
		{
			name:  "failing parser should fail",
			input: "abc\r\n",
			args: args{
				Map(Digit1(), func(digit string) (int, error) {
					i, _ := strconv.Atoi(digit)
					return i, nil
				}),
			},
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "abc\r\n",
		},
		{
			name:  "failing mapper should fail",
			input: "1abc\r\n",
			args: args{
				Map(Digit1(), func(digit string) (int, error) {
					return 0, errors.New("unexpected error")
				}),
			},
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "abc\r\n",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				Map(Digit1(), func(digit string) (int, error) {
					i, _ := strconv.Atoi(digit)
					return i, nil
				}),
			},
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %#v, want output %#v", gotResult, tc.wantOutput)
			}
			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkMap(b *testing.B) {
	parser := Map(Digit1(), func(digit string) (int, error) {
		i, _ := strconv.Atoi(digit)
		return i, nil
	})
	input := gomme.NewFromString(1, nil, -1, "123abc\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestMap2(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Foo int
		Bar string
	}

	type args struct {
		parser gomme.Parser[TestStruct]
	}
	testCases := []struct {
		name          string
		input         string
		args          args
		wantErr       bool
		wantOutput    TestStruct
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "1abc\r\n",
			args: args{
				Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
					left, _ := strconv.Atoi(digit)
					return TestStruct{Foo: left, Bar: alpha}, nil
				}),
			},
			wantErr:       false,
			wantOutput:    TestStruct{Foo: 1, Bar: "abc"},
			wantRemaining: "\r\n",
		},
		{
			name:  "failing parser should fail",
			input: "abc\r\n",
			args: args{
				Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
					left, _ := strconv.Atoi(digit)
					return TestStruct{Foo: left, Bar: alpha}, nil
				}),
			},
			wantErr:       true,
			wantOutput:    TestStruct{Bar: "abc"},
			wantRemaining: "\r\n",
		},
		{
			name:  "failing mapper should fail",
			input: "1abc\r\n",
			args: args{
				Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
					return TestStruct{}, errors.New("unexpected error")
				}),
			},
			wantErr:       true,
			wantOutput:    TestStruct{},
			wantRemaining: "\r\n",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
					left, _ := strconv.Atoi(digit)
					return TestStruct{Foo: left, Bar: alpha}, nil
				}),
			},
			wantErr:       true,
			wantOutput:    TestStruct{},
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult := gomme.RunOnState(gomme.NewFromString(-1, nil, -1, tc.input), tc.args.parser)
			if newState.HasError() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %#v, want output %#v", gotResult, tc.wantOutput)
			}
			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkMap2(b *testing.B) {
	type TestStruct struct {
		Foo int
		Bar string
	}

	parser := Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
		first, _ := strconv.Atoi(digit)
		return TestStruct{Foo: first, Bar: alpha}, nil
	})
	input := gomme.NewFromString(1, nil, -1, "1abc\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func pairMapFunc(_ string, _ string) (string, error) {
	return "", nil
}
