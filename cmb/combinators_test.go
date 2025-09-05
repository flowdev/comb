package cmb_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
)

func TestOptional(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[string]
		input      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "matching parser should succeed",
			input:      "\r\n123",
			parser:     cmb.Optional(cmb.CRLF()),
			wantErr:    false,
			wantOutput: "\r\n",
		},
		{
			name:       "no match should succeed",
			input:      "123",
			parser:     cmb.Optional(cmb.CRLF()),
			wantErr:    false,
			wantOutput: "",
		},
		{
			name:       "empty input should succeed",
			input:      "",
			parser:     cmb.Optional(cmb.CRLF()),
			wantErr:    false,
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
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkOptional(b *testing.B) {
	parser := cmb.Optional(cmb.CR())
	input := comb.NewFromString("\r123", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestPeek(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[string]
		input      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "matching parser should succeed",
			input:      "abcd;",
			parser:     cmb.Peek(cmb.Alpha1()),
			wantErr:    false,
			wantOutput: "abcd",
		}, {
			name:       "non matching parser should fail",
			input:      "123;",
			parser:     cmb.Peek(cmb.Alpha1()),
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
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkPeek(b *testing.B) {
	parser := cmb.Peek(cmb.Alpha1())
	input := comb.NewFromString("abcd;", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestNot(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[bool]
		input      string
		wantErr    bool
		wantOutput bool
	}{
		{
			name:       "not matching parser should succeed",
			input:      "123;",
			parser:     cmb.Not(cmb.Alpha1()),
			wantErr:    false,
			wantOutput: true,
		}, {
			name:       "matching parser should fail",
			input:      "abcd;",
			parser:     cmb.Not(cmb.Alpha1()),
			wantErr:    true,
			wantOutput: false,
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
				t.Errorf("got output %t, want output %t", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkNot(b *testing.B) {
	parser := cmb.Not(cmb.Alpha1())
	input := comb.NewFromString("abcd;", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestStringUntil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[string]
		input      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "immediately matching parser should succeed",
			input:      "abcd;",
			parser:     cmb.StringUntil(cmb.Alpha1()),
			wantErr:    false,
			wantOutput: "",
		}, {
			name:       "non matching parser should fail",
			input:      "123;",
			parser:     cmb.StringUntil(cmb.Alpha1()),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "matching parser should succeed",
			input:      "123abcd;",
			parser:     cmb.StringUntil(cmb.Alpha1()),
			wantErr:    false,
			wantOutput: "123",
		}, {
			name:       "empty input should fail",
			input:      "",
			parser:     cmb.StringUntil(cmb.Alpha1()),
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
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkStringUntil(b *testing.B) {
	parser := cmb.Peek(cmb.Alpha1())
	input := comb.NewFromString("abcd;", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestAssign(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[int]
		input      string
		wantErr    bool
		wantOutput int
	}{
		{
			name:       "matching parser should succeed",
			input:      "abcd",
			parser:     cmb.Assign(1234, cmb.Alpha1()),
			wantErr:    false,
			wantOutput: 1234,
		}, {
			name:       "non matching parser should fail",
			input:      "123abcd;",
			parser:     cmb.Assign(1234, cmb.Alpha1()),
			wantErr:    true,
			wantOutput: 1234,
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
				t.Errorf("got output %d, want output %d", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkAssign(b *testing.B) {
	parser := cmb.Assign(1234, cmb.Alpha1())
	input := comb.NewFromString("abcd", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestDelimited(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[string]
		input      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "matching parser should succeed",
			input:      "+1\r\n",
			parser:     cmb.Delimited(cmb.Char('+'), cmb.Digit1(), cmb.CRLF()),
			wantErr:    false,
			wantOutput: "1",
		}, {
			name:       "no prefix match should fail",
			input:      "1\r\n",
			parser:     cmb.Delimited(cmb.Char('+'), comb.SafeSpot(cmb.Digit1()), cmb.CRLF()),
			wantErr:    true,
			wantOutput: "1",
		}, {
			name:       "no parser match should fail",
			input:      "+\r\n",
			parser:     cmb.Delimited(cmb.Char('+'), cmb.Digit1(), cmb.CRLF()),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "no suffix match should fail",
			input:      "+1",
			parser:     cmb.Delimited(cmb.Char('+'), cmb.Digit1(), cmb.CRLF()),
			wantErr:    true,
			wantOutput: "1",
		}, {
			name:       "empty input should fail",
			input:      "",
			parser:     cmb.Delimited(cmb.Char('+'), cmb.Digit1(), cmb.CRLF()),
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
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkDelimited(b *testing.B) {
	parser := cmb.Delimited(cmb.Char('+'), cmb.Digit1(), cmb.CRLF())
	input := comb.NewFromString("+1\r\n", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestPrefixed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[string]
		input      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "matching parser should succeed",
			input:      "+123",
			parser:     cmb.Prefixed(cmb.Char('+'), cmb.Digit1()),
			wantErr:    false,
			wantOutput: "123",
		}, {
			name:       "no prefix match should fail",
			input:      "+123",
			parser:     cmb.Prefixed(cmb.Char('-'), cmb.Digit1()),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "no parser match should fail",
			input:      "+",
			parser:     cmb.Prefixed(cmb.Char('+'), cmb.Digit1()),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "empty input should fail",
			input:      "",
			parser:     cmb.Prefixed(cmb.Char('+'), cmb.Digit1()),
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
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkPrefixed(b *testing.B) {
	parser := cmb.Prefixed(cmb.Char('+'), cmb.Digit1())
	input := comb.NewFromString("+123", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestSuffixed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[string]
		input      string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "matching parser should succeed",
			input:      "1+23",
			parser:     cmb.Suffixed(cmb.Digit1(), cmb.Char('+')),
			wantErr:    false,
			wantOutput: "1",
		}, {
			name:       "no suffix match should fail",
			input:      "1-23",
			parser:     cmb.Suffixed(cmb.Digit1(), cmb.Char('+')),
			wantErr:    true,
			wantOutput: "1",
		}, {
			name:       "no parser match should fail",
			input:      "+",
			parser:     cmb.Suffixed(cmb.Digit1(), cmb.Char('+')),
			wantErr:    true,
			wantOutput: "",
		}, {
			name:       "empty input should fail",
			input:      "",
			parser:     cmb.Suffixed(cmb.Digit1(), cmb.Char('+')),
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
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkTerminated(b *testing.B) {
	parser := cmb.Suffixed(cmb.Digit1(), cmb.Char('+'))
	input := comb.NewFromString("123+", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		input      string
		parser     comb.Parser[int]
		wantErr    bool
		wantOutput int
	}{
		{
			name:  "matching parser should succeed",
			input: "1abc\r\n",
			parser: cmb.Map(cmb.Digit1(), func(digit string) (int, error) {
				i, _ := strconv.Atoi(digit)
				return i, nil
			}),
			wantErr:    false,
			wantOutput: 1,
		}, {
			name:  "failing parser should fail",
			input: "abc\r\n",
			parser: cmb.Map(cmb.Digit1(), func(digit string) (int, error) {
				i, _ := strconv.Atoi(digit)
				return i, nil
			}),
			wantErr:    true,
			wantOutput: 0,
		}, {
			name:  "failing mapper should fail",
			input: "1abc\r\n",
			parser: cmb.Map(cmb.Digit1(), func(digit string) (int, error) {
				return 0, errors.New("unexpected error")
			}),
			wantErr:    true,
			wantOutput: 0,
		}, {
			name:  "empty input should fail",
			input: "",
			parser: cmb.Map(cmb.Digit1(), func(digit string) (int, error) {
				i, _ := strconv.Atoi(digit)
				return i, nil
			}),
			wantErr:    true,
			wantOutput: 0,
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
				t.Errorf("got output %#v, want output %#v", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkMap(b *testing.B) {
	parser := cmb.Map(cmb.Digit1(), func(digit string) (int, error) {
		i, _ := strconv.Atoi(digit)
		return i, nil
	})
	input := comb.NewFromString("123abc\r\n", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestMap2(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Foo int
		Bar string
	}

	testCases := []struct {
		name       string
		input      string
		parser     comb.Parser[TestStruct]
		wantErr    bool
		wantOutput TestStruct
	}{
		{
			name:  "matching parser should succeed",
			input: "1abc\r\n",
			parser: cmb.Map2(cmb.Digit1(), cmb.Alpha1(), func(digit string, alpha string) (TestStruct, error) {
				left, _ := strconv.Atoi(digit)
				return TestStruct{Foo: left, Bar: alpha}, nil
			}),
			wantErr:    false,
			wantOutput: TestStruct{Foo: 1, Bar: "abc"},
		}, {
			name:  "failing parser should fail",
			input: "abc\r\n",
			parser: cmb.Map2(cmb.Digit1(), comb.SafeSpot(cmb.Alpha1()), func(digit string, alpha string) (TestStruct, error) {
				left, _ := strconv.Atoi(digit)
				return TestStruct{Foo: left, Bar: alpha}, nil
			}),
			wantErr:    true,
			wantOutput: TestStruct{Bar: "abc"},
		}, {
			name:  "failing mapper should fail",
			input: "1abc\r\n",
			parser: cmb.Map2(cmb.Digit1(), cmb.Alpha1(), func(digit string, alpha string) (TestStruct, error) {
				return TestStruct{}, errors.New("unexpected error")
			}),
			wantErr:    true,
			wantOutput: TestStruct{},
		}, {
			name:  "empty input should fail",
			input: "",
			parser: cmb.Map2(cmb.Digit1(), cmb.Alpha1(), func(digit string, alpha string) (TestStruct, error) {
				left, _ := strconv.Atoi(digit)
				return TestStruct{Foo: left, Bar: alpha}, nil
			}),
			wantErr:    true,
			wantOutput: TestStruct{},
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
				t.Errorf("got output %#v, want output %#v", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkMap2(b *testing.B) {
	type TestStruct struct {
		Foo int
		Bar string
	}

	parser := cmb.Map2(cmb.Digit1(), cmb.Alpha1(), func(digit string, alpha string) (TestStruct, error) {
		first, _ := strconv.Atoi(digit)
		return TestStruct{Foo: first, Bar: alpha}, nil
	})
	input := comb.NewFromString("1abc\r\n", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}
