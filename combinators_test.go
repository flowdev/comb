package gomme

import (
	"errors"
	"strconv"
	"testing"
)

func TestOptional(t *testing.T) {
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

			gotResult := tc.args.p(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			if gotResult.Output != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult.Output, tc.wantOutput)
			}

			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %v, want remaining %v", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkOptional(b *testing.B) {
	p := Optional(CRLF())
	input := NewInputFromString("\r\n123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

func TestPeek(t *testing.T) {
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

func BenchmarkPeek(b *testing.B) {
	p := Peek(Alpha1())
	input := NewInputFromString("abcd;")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

func TestRecognize(t *testing.T) {
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
			input: "123abc",
			args: args{
				p: Recognize(Pair(Digit1(), Alpha1())),
			},
			wantErr:       false,
			wantOutput:    "123abc",
			wantRemaining: "",
		},
		{
			name:  "no prefix match should fail",
			input: "abc",
			args: args{
				p: Recognize(Pair(Digit1(), Alpha1())),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
		{
			name:  "no parser match should fail",
			input: "123",
			args: args{
				p: Recognize(Pair(Digit1(), Alpha1())),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: Recognize(Pair(Digit1(), Alpha1())),
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

func BenchmarkRecognize(b *testing.B) {
	p := Recognize(Pair(Digit1(), Alpha1()))
	input := NewInputFromString("123abc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

func TestAssign(t *testing.T) {
	t.Parallel()

	type args struct {
		p Parser[int]
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

func BenchmarkAssign(b *testing.B) {
	p := Assign(1234, Alpha1())
	input := NewInputFromString("abcd")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

func TestMap1(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Foo int
		Bar string
	}

	type args struct {
		parser Parser[TestStruct]
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
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
					left, _ := strconv.Atoi(p.Left)
					return TestStruct{Foo: left, Bar: string(p.Right)}, nil
				}),
			},
			wantErr:       false,
			wantOutput:    TestStruct{Foo: 1, Bar: "abc"},
			wantRemaining: "",
		},
		{
			name:  "failing parser should fail",
			input: "abc\r\n",
			args: args{
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
					left, _ := strconv.Atoi(p.Left)
					return TestStruct{Foo: left, Bar: string(p.Right)}, nil
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
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
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
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
					left, _ := strconv.Atoi(p.Left)
					return TestStruct{Foo: left, Bar: string(p.Right)}, nil
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

			gotResult := tc.args.parser(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			if gotResult.Output != tc.wantOutput {
				t.Errorf("got output %#v, want output %#v", gotResult.Output, tc.wantOutput)
			}
			remainingString := gotResult.Remaining.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkMap1(b *testing.B) {
	type TestStruct struct {
		Foo int
		Bar string
	}

	p := Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
		left, _ := strconv.Atoi(p.Left)
		return TestStruct{Foo: left, Bar: string(p.Right)}, nil
	})
	input := NewInputFromString("1abc\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}

func TestMap2(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Foo int
		Bar string
	}

	type args struct {
		parser Parser[TestStruct]
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
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
					left, _ := strconv.Atoi(p.Left)
					return TestStruct{Foo: left, Bar: string(p.Right)}, nil
				}),
			},
			wantErr:       false,
			wantOutput:    TestStruct{Foo: 1, Bar: "abc"},
			wantRemaining: "",
		},
		{
			name:  "failing parser should fail",
			input: "abc\r\n",
			args: args{
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
					left, _ := strconv.Atoi(p.Left)
					return TestStruct{Foo: left, Bar: string(p.Right)}, nil
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
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
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
				Map1(Pair(Digit1(), TakeUntil(CRLF())), func(p PairContainer[string, []byte]) (TestStruct, error) {
					left, _ := strconv.Atoi(p.Left)
					return TestStruct{Foo: left, Bar: string(p.Right)}, nil
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

			gotResult := tc.args.parser(NewInputFromString(tc.input))
			if (gotResult.Err != nil) != tc.wantErr {
				t.Errorf("got error %v, want error %v", gotResult.Err, tc.wantErr)
			}

			if gotResult.Output != tc.wantOutput {
				t.Errorf("got output %#v, want output %#v", gotResult.Output, tc.wantOutput)
			}
			remainingString := gotResult.Remaining.CurrentString()
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

	p := Map2(Digit1(), TakeUntil(CRLF()), func(s string, bs []byte) (TestStruct, error) {
		first, _ := strconv.Atoi(s)
		return TestStruct{Foo: first, Bar: string(bs)}, nil
	})
	input := NewInputFromString("1abc\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p(input)
	}
}
