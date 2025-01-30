package pcb

import (
	"errors"
	"github.com/oleiade/gomme"
	"strconv"
	"testing"
)

func TestOptional(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "\r\n123",
			parser:        Optional(CRLF()),
			wantErr:       false,
			wantOutput:    "\r\n",
			wantRemaining: "123",
		},
		{
			name:          "no match should succeed",
			input:         "123",
			parser:        Optional(CRLF()),
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:          "empty input should succeed",
			input:         "",
			parser:        Optional(CRLF()),
			wantErr:       false,
			wantOutput:    "",
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkOptional(b *testing.B) {
	parser := Optional(CR())
	input := gomme.NewFromString("\r123", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestPeek(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "abcd;",
			parser:        Peek(Alpha1()),
			wantErr:       false,
			wantOutput:    "abcd",
			wantRemaining: "abcd;",
		},
		{
			name:          "non matching parser should fail",
			input:         "123;",
			parser:        Peek(Alpha1()),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123;",
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkPeek(b *testing.B) {
	parser := Peek(Alpha1())
	input := gomme.NewFromString("abcd;", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestAssign(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[int]
		input         string
		wantErr       bool
		wantOutput    int
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "abcd",
			parser:        Assign(1234, Alpha1()),
			wantErr:       false,
			wantOutput:    1234,
			wantRemaining: "",
		},
		{
			name:          "non matching parser should fail",
			input:         "123abcd;",
			parser:        Assign(1234, Alpha1()),
			wantErr:       true,
			wantOutput:    1234,
			wantRemaining: "123abcd;",
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %d, want output %d", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkAssign(b *testing.B) {
	parser := Assign(1234, Alpha1())
	input := gomme.NewFromString("abcd", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestDelimited(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "+1\r\n",
			parser:        Delimited(Char('+'), Digit1(), CRLF()),
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		}, {
			name:          "no prefix match should fail",
			input:         "1\r\n",
			parser:        Delimited(Char('+'), gomme.SafeSpot(Digit1()), CRLF()),
			wantErr:       true,
			wantOutput:    "1",
			wantRemaining: "",
		}, {
			name:          "no parser match should fail",
			input:         "+\r\n",
			parser:        Delimited(Char('+'), Digit1(), CRLF()),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		}, {
			name:          "no suffix match should fail",
			input:         "+1",
			parser:        Delimited(Char('+'), Digit1(), CRLF()),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		}, {
			name:          "empty input should fail",
			input:         "",
			parser:        Delimited(Char('+'), Digit1(), CRLF()),
			wantErr:       true,
			wantOutput:    "",
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkDelimited(b *testing.B) {
	parser := Delimited(Char('+'), Digit1(), CRLF())
	input := gomme.NewFromString("+1\r\n", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestPrefixed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "+123",
			parser:        Prefixed(Char('+'), Digit1()),
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "no prefix match should fail",
			input:         "+123",
			parser:        Prefixed(Char('-'), Digit1()),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
		{
			name:          "no parser match should fail",
			input:         "+",
			parser:        Prefixed(Char('+'), Digit1()),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "empty input should fail",
			input:         "",
			parser:        Prefixed(Char('+'), Digit1()),
			wantErr:       true,
			wantOutput:    "",
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkPrefixed(b *testing.B) {
	parser := Prefixed(Char('+'), Digit1())
	input := gomme.NewFromString("+123", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestSuffixed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "matching parser should succeed",
			input:         "1+23",
			parser:        Suffixed(Digit1(), Char('+')),
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "23",
		},
		{
			name:          "no suffix match should fail",
			input:         "1-23",
			parser:        Suffixed(Digit1(), Char('+')),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "-23",
		},
		{
			name:          "no parser match should fail",
			input:         "+",
			parser:        Suffixed(Digit1(), Char('+')),
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "empty input should fail",
			input:         "",
			parser:        Suffixed(Digit1(), Char('+')),
			wantErr:       true,
			wantOutput:    "",
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

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkTerminated(b *testing.B) {
	parser := Suffixed(Digit1(), Char('+'))
	input := gomme.NewFromString("123+", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		parser        gomme.Parser[int]
		wantErr       bool
		wantOutput    int
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "1abc\r\n",
			parser: Map(Digit1(), func(digit string) (int, error) {
				i, _ := strconv.Atoi(digit)
				return i, nil
			}),
			wantErr:       false,
			wantOutput:    1,
			wantRemaining: "abc\r\n",
		},
		{
			name:  "failing parser should fail",
			input: "abc\r\n",
			parser: Map(Digit1(), func(digit string) (int, error) {
				i, _ := strconv.Atoi(digit)
				return i, nil
			}),
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "abc\r\n",
		},
		{
			name:  "failing mapper should fail",
			input: "1abc\r\n",
			parser: Map(Digit1(), func(digit string) (int, error) {
				return 0, errors.New("unexpected error")
			}),
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "abc\r\n",
		},
		{
			name:  "empty input should fail",
			input: "",
			parser: Map(Digit1(), func(digit string) (int, error) {
				i, _ := strconv.Atoi(digit)
				return i, nil
			}),
			wantErr:       true,
			wantOutput:    0,
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
			if gotResult != tc.wantOutput {
				t.Errorf("got output %#v, want output %#v", gotResult, tc.wantOutput)
			}
		})
	}
}

func BenchmarkMap(b *testing.B) {
	parser := Map(Digit1(), func(digit string) (int, error) {
		i, _ := strconv.Atoi(digit)
		return i, nil
	})
	input := gomme.NewFromString("123abc\r\n", false)

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

	testCases := []struct {
		name          string
		input         string
		parser        gomme.Parser[TestStruct]
		wantErr       bool
		wantOutput    TestStruct
		wantRemaining string
	}{
		{
			name:  "matching parser should succeed",
			input: "1abc\r\n",
			parser: Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
				left, _ := strconv.Atoi(digit)
				return TestStruct{Foo: left, Bar: alpha}, nil
			}),
			wantErr:       false,
			wantOutput:    TestStruct{Foo: 1, Bar: "abc"},
			wantRemaining: "\r\n",
		}, {
			name:  "failing parser should fail",
			input: "abc\r\n",
			parser: Map2(Digit1(), gomme.SafeSpot(Alpha1()), func(digit string, alpha string) (TestStruct, error) {
				left, _ := strconv.Atoi(digit)
				return TestStruct{Foo: left, Bar: alpha}, nil
			}),
			wantErr:       true,
			wantOutput:    TestStruct{Bar: "abc"},
			wantRemaining: "\r\n",
		}, {
			name:  "failing mapper should fail",
			input: "1abc\r\n",
			parser: Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
				return TestStruct{}, errors.New("unexpected error")
			}),
			wantErr:       true,
			wantOutput:    TestStruct{},
			wantRemaining: "\r\n",
		}, {
			name:  "empty input should fail",
			input: "",
			parser: Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
				left, _ := strconv.Atoi(digit)
				return TestStruct{Foo: left, Bar: alpha}, nil
			}),
			wantErr:       true,
			wantOutput:    TestStruct{},
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

	parser := Map2(Digit1(), Alpha1(), func(digit string, alpha string) (TestStruct, error) {
		first, _ := strconv.Atoi(digit)
		return TestStruct{Foo: first, Bar: alpha}, nil
	})
	input := gomme.NewFromString("1abc\r\n", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gomme.RunOnState(input, parser)
	}
}
