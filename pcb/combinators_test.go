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
			input: "\r\n123",
			args: args{
				p: Optional(CRLF()),
			},
			wantErr:       false,
			wantOutput:    "\r\n",
			wantRemaining: "123",
		},
		{
			name:  "no match should succeed",
			input: "123",
			args: args{
				p: Optional(CRLF()),
			},
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:  "empty input should succeed",
			input: "",
			args: args{
				p: Optional(CRLF()),
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

			newState, gotResult := tc.args.p.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Error(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %v, want remaining %v", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkOptional(b *testing.B) {
	p := Optional(CRLF())
	input := gomme.NewFromString("\r\n123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
	}
}

func TestPeek(t *testing.T) {
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
			input: "abcd;",
			args: args{
				p: Peek(Alpha1()),
			},
			wantErr:       false,
			wantOutput:    "abcd",
			wantRemaining: "abcd;",
		},
		{
			name:  "non matching parser should fail",
			input: "123;",
			args: args{
				p: Peek(Alpha1()),
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

func BenchmarkPeek(b *testing.B) {
	p := Peek(Alpha1())
	input := gomme.NewFromString("abcd;")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
	}
}

func TestRecognize(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[[]byte]
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
				p: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)), /*func(po1 string, po2 string) (string, error) {
					return po1 + po2, nil
				})),*/
			},
			wantErr:       false,
			wantOutput:    "123abc",
			wantRemaining: "",
		},
		{
			name:  "no prefix match should fail",
			input: "abc",
			args: args{
				p: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
		{
			name:  "no parser match should fail",
			input: "123",
			args: args{
				p: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Recognize(Map2(Digit1(), Alpha1(), pairMapFunc)),
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

			if string(gotResult) != tc.wantOutput {
				t.Errorf("got output %v, want output %v", string(gotResult), tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkRecognize(b *testing.B) {
	p := Recognize(Map2(Digit1(), Alpha1(), pairMapFunc))
	input := gomme.NewFromString("123abc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
	}
}

func TestAssign(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[int]
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
				p: Assign(1234, Alpha1()),
			},
			wantErr:       false,
			wantOutput:    1234,
			wantRemaining: "",
		},
		{
			name:  "non matching parser should fail",
			input: "123abcd;",
			args: args{
				p: Assign(1234, Alpha1()),
			},
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "123abcd;",
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

func BenchmarkAssign(b *testing.B) {
	p := Assign(1234, Alpha1())
	input := gomme.NewFromString("abcd")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
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
			wantRemaining: "1abc\r\n",
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

			newState, gotResult := tc.args.parser.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
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

func BenchmarkMap1(b *testing.B) {
	p := Map(Digit1(), func(digit string) (int, error) {
		i, _ := strconv.Atoi(digit)
		return i, nil
	})
	input := gomme.NewFromString("123abc\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
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
			wantOutput:    TestStruct{},
			wantRemaining: "abc\r\n",
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
			wantRemaining: "1abc\r\n",
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

			newState, gotResult := tc.args.parser.It(gomme.NewFromString(tc.input))
			if newState.Failed() != tc.wantErr {
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

	p := Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
		first, _ := strconv.Atoi(digit)
		return TestStruct{Foo: first, Bar: alpha}, nil
	})
	input := gomme.NewFromString("1abc\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
	}
}

func pairMapFunc(_ string, _ string) (string, error) {
	return "", nil
}
