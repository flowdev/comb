package gomme

import (
	"testing"
	"unicode"
)

func TestChar(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing char from single char input should succeed",
			parser:        Char('a'),
			input:         "a",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "",
		},
		{
			name:          "parsing valid char in longer input should succeed",
			parser:        Char('a'),
			input:         "abc",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "bc",
		},
		{
			name:          "parsing single non-char input should fail",
			parser:        Char('a'),
			input:         "123",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "123",
		},
		{
			name:          "parsing empty input should fail",
			parser:        Char('a'),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
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

func BenchmarkChar(b *testing.B) {
	parser := Char('a')
	input := NewInputFromString("a")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestAnyChar(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing any char from single entry input should succeed",
			parser:        AnyChar(),
			input:         "a",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "",
		},
		{
			name:          "parsing valid any char from longer input should succeed",
			parser:        AnyChar(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "bc",
		},
		{
			name:          "parsing any char from empty input should fail",
			parser:        AnyChar(),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
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

func BenchmarkAnyChar(b *testing.B) {
	parser := AnyChar()
	input := NewInputFromString("a")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestAlpha0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alpha input should succeed",
			parser:        Alpha0(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alpha input should succeed",
			parser:        Alpha0(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars until terminating char should succeed",
			parser:        Alpha0(),
			input:         "abc123",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "123",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        Alpha0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non alpha chars should succeed",
			parser:        Alpha0(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "123",
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

func BenchmarkAlpha0(b *testing.B) {
	parser := Alpha0()
	input := NewInputFromString("abc")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestAlpha1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alpha input should succeed",
			parser:        Alpha1(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alpha input should succeed",
			parser:        Alpha1(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars until terminating char should succeed",
			parser:        Alpha1(),
			input:         "abc123",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "123",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        Alpha1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with an alpha char should fail",
			parser:        Alpha1(),
			input:         "1c",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "1c",
		},
		{
			name:          "parsing non alpha chars should fail",
			parser:        Alpha1(),
			input:         "123",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
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

func BenchmarkAlpha1(b *testing.B) {
	parser := Alpha1()
	input := NewInputFromString("abc")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestDigit0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single digit char from single digit input should succeed",
			parser:        Digit0(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple digit input should succeed",
			parser:        Digit0(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        Digit0(),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        Digit0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non digit chars should succeed",
			parser:        Digit0(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "abc",
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

func BenchmarkDigit0(b *testing.B) {
	parser := Digit0()
	input := NewInputFromString("123")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestDigit1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single digit char from single digit input should succeed",
			parser:        Digit1(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple digit input should succeed",
			parser:        Digit1(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        Digit1(),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        Digit1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with an digit char should fail",
			parser:        Digit1(),
			input:         "c1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "c1",
		},
		{
			name:          "parsing non digit chars should fail",
			parser:        Digit1(),
			input:         "abc",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
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

func BenchmarkDigit1(b *testing.B) {
	parser := Digit1()
	input := NewInputFromString("123")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestHexDigit0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single hex digit char from single hex digit input should succeed",
			parser:        HexDigit0(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars from multiple hex digit input should succeed",
			parser:        HexDigit0(),
			input:         "1f3",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars until terminating char should succeed",
			parser:        HexDigit0(),
			input:         "1f3z",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "z",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        HexDigit0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non hex digit chars should succeed",
			parser:        HexDigit0(),
			input:         "ghi",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "ghi",
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

func BenchmarkHexDigit0(b *testing.B) {
	parser := HexDigit0()
	input := NewInputFromString("1f3")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestHexDigit1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single hex digit char from single hex digit input should succeed",
			parser:        HexDigit1(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars from multiple hex digit input should succeed",
			parser:        HexDigit1(),
			input:         "1f3",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars until terminating char should succeed",
			parser:        HexDigit1(),
			input:         "1f3ghi",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "ghi",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        HexDigit1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with a hex digit char should fail",
			parser:        HexDigit1(),
			input:         "h1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "h1",
		},
		{
			name:          "parsing non hex digit chars should fail",
			parser:        HexDigit1(),
			input:         "ghi",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "ghi",
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

func BenchmarkHexDigit1(b *testing.B) {
	parser := HexDigit1()
	input := NewInputFromString("1f3")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestWhitespace0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single whitespace from single ' ' input should succeed",
			parser:        Whitespace0(),
			input:         " ",
			wantErr:       false,
			wantOutput:    " ",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\t' input should succeed",
			parser:        Whitespace0(),
			input:         "\t",
			wantErr:       false,
			wantOutput:    "\t",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\n' input should succeed",
			parser:        Whitespace0(),
			input:         "\n",
			wantErr:       false,
			wantOutput:    "\n",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\r' input should succeed",
			parser:        Whitespace0(),
			input:         "\r",
			wantErr:       false,
			wantOutput:    "\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars input should succeed",
			parser:        Whitespace0(),
			input:         " \t\n\r",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars with suffix input should succeed",
			parser:        Whitespace0(),
			input:         " \t\n\rabc",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        Whitespace0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing a single non-whitespace char input should succeed",
			parser:        Whitespace0(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "a",
		},
		{
			name:          "parsing input starting with a non-whitespace char should succeed",
			parser:        Whitespace0(),
			input:         "a \t\n\r",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "a \t\n\r",
		},
		{
			name:          "parsing non-whitespace chars should succeed",
			parser:        Whitespace0(),
			input:         "ghi",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "ghi",
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

func BenchmarkWhitespace0(b *testing.B) {
	b.ReportAllocs()
	parser := Whitespace0()
	input := NewInputFromString(" \t\n\r")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestWhitespace1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single whitespace from single ' ' input should succeed",
			parser:        Whitespace1(),
			input:         " ",
			wantErr:       false,
			wantOutput:    " ",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\t' input should succeed",
			parser:        Whitespace1(),
			input:         "\t",
			wantErr:       false,
			wantOutput:    "\t",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\n' input should succeed",
			parser:        Whitespace1(),
			input:         "\n",
			wantErr:       false,
			wantOutput:    "\n",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\r' input should succeed",
			parser:        Whitespace1(),
			input:         "\r",
			wantErr:       false,
			wantOutput:    "\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars input should succeed",
			parser:        Whitespace1(),
			input:         " \t\n\r",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars with suffix input should succeed",
			parser:        Whitespace1(),
			input:         " \t\n\rabc",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        Whitespace1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing a single non-whitespace char input should fail",
			parser:        Whitespace1(),
			input:         "a",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "a",
		},
		{
			name:          "parsing input starting with a non-whitespace char should fail",
			parser:        Whitespace1(),
			input:         "a \t\n\r",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "a \t\n\r",
		},
		{
			name:          "parsing non-whitespace chars should fail",
			parser:        Whitespace1(),
			input:         "ghi",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "ghi",
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

func BenchmarkWhitespace1(b *testing.B) {
	b.ReportAllocs()
	input := NewInputFromString(" \t\n\r")

	parser := Whitespace1()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestAlphanumeric0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alphanumerical input should succeed",
			parser:        Alphanumeric0(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing single digit char from single alphanumerical input should succeed",
			parser:        Alphanumeric0(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alphanumerical input should succeed",
			parser:        Alphanumeric0(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple alphanumerical input should succeed",
			parser:        Alphanumeric0(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple alphanumerical input should succeed",
			parser:        Alphanumeric0(),
			input:         "a1b2c3",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "",
		},
		{
			name:          "parsing alph chars until terminating char should succeed",
			parser:        Alphanumeric0(),
			input:         "abc$%^",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        Alphanumeric0(),
			input:         "123$%^",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing alphanumerical chars until terminating char should succeed",
			parser:        Alphanumeric0(),
			input:         "a1b2c3$%^",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        Alphanumeric0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non alphanumerical chars should succeed",
			parser:        Alphanumeric0(),
			input:         "$%^",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "$%^",
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

func BenchmarkAlphanumeric0(b *testing.B) {
	parser := Alphanumeric0()
	input := NewInputFromString("a1b2c3")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestAlphanumeric1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alphanumerical input should succeed",
			parser:        Alphanumeric1(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing single digit char from single alphanumerical input should succeed",
			parser:        Alphanumeric1(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alphanumerical input should succeed",
			parser:        Alphanumeric1(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple alphanumerical input should succeed",
			parser:        Alphanumeric1(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing alphanumerical chars from multiple alphanumerical input should succeed",
			parser:        Alphanumeric1(),
			input:         "a1b2c3",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars until terminating char should succeed",
			parser:        Alphanumeric1(),
			input:         "abc$%^",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        Alphanumeric1(),
			input:         "123$%^",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing alphanumerical chars until terminating char should succeed",
			parser:        Alphanumeric1(),
			input:         "a1b2c3$%^",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        Alphanumeric1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with an alphanumeric char should fail",
			parser:        Alphanumeric1(),
			input:         "$1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$1",
		},
		{
			name:          "parsing non digit chars should fail",
			parser:        Alphanumeric1(),
			input:         "$%^",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$%^",
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

func BenchmarkAlphanumeric1(b *testing.B) {
	parser := Alphanumeric1()
	input := NewInputFromString("a1b2c3")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestLF(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single line-feed char from single line-feed input should succeed",
			parser:        LF(),
			input:         "\n",
			wantErr:       false,
			wantOutput:    rune('\n'),
			wantRemaining: "",
		},
		{
			name:          "parsing single line-feed char from multiple char input should succeed",
			parser:        LF(),
			input:         "\nabc",
			wantErr:       false,
			wantOutput:    rune('\n'),
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        LF(),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "",
		},
		{
			name:          "parsing single line-feed char from single non-line-feed input should fail",
			parser:        LF(),
			input:         "1",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "1",
		},
		{
			name:          "parsing single line-feed from multiple non-line-feed input should fail",
			parser:        LF(),
			input:         "123",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "123",
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

func BenchmarkLF(b *testing.B) {
	parser := LF()
	input := NewInputFromString("\n")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestCR(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single carriage-return char from single carriage-return input should succeed",
			parser:        CR(),
			input:         "\r",
			wantErr:       false,
			wantOutput:    rune('\r'),
			wantRemaining: "",
		},
		{
			name:          "parsing single carriage-return char from multiple char input should succeed",
			parser:        CR(),
			input:         "\rabc",
			wantErr:       false,
			wantOutput:    rune('\r'),
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        CR(),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "",
		},
		{
			name:          "parsing single carriage-return char from single non-carriage-return input should fail",
			parser:        CR(),
			input:         "1",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "1",
		},
		{
			name:          "parsing single carriage-return from multiple non-carriage-return input should fail",
			parser:        CR(),
			input:         "123",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "123",
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

func BenchmarkCR(b *testing.B) {
	parser := CR()
	input := NewInputFromString("\r")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestCRLF(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single CRLF char from single CRLF input should succeed",
			parser:        CRLF(),
			input:         "\r\n",
			wantErr:       false,
			wantOutput:    "\r\n",
			wantRemaining: "",
		},
		{
			name:          "parsing single CRLF char from multiple char input should succeed",
			parser:        CRLF(),
			input:         "\r\nabc",
			wantErr:       false,
			wantOutput:    "\r\n",
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        CRLF(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing incomplete CRLF input should fail",
			parser:        CRLF(),
			input:         "\r",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "\r",
		},
		{
			name:          "parsing single CRLF char from single non-CRLF input should fail",
			parser:        CRLF(),
			input:         "1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "1",
		},
		{
			name:          "parsing single CRLF from multiple non-CRLF input should fail",
			parser:        CRLF(),
			input:         "123",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
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

func BenchmarkCRLF(b *testing.B) {
	parser := CRLF()
	input := NewInputFromString("\r\n")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestOneOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing matched char should succeed",
			parser:        OneOf('a', '1', '+'),
			input:         "+",
			wantErr:       false,
			wantOutput:    '+',
			wantRemaining: "",
		},
		{
			name:          "parsing input not containing any of the sought chars should fail",
			parser:        OneOf('a', '1', '+'),
			input:         "b",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "b",
		},
		{
			name:          "parsing empty input should fail",
			parser:        OneOf('a', '1', '+'),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
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

func BenchmarkOneOf(b *testing.B) {
	parser := OneOf('a', '1', '+')
	input := NewInputFromString("+")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestSatisfy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char satisfying constraint should succeed",
			parser:        Satisfy(unicode.IsLetter),
			input:         "a",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "",
		},
		{
			name:          "parsing alpha char satisfying constraint from mixed input should succeed",
			parser:        Satisfy(unicode.IsLetter),
			input:         "a1",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "1",
		},
		{
			name:          "parsing char not satisfying constraint should succeed",
			parser:        Satisfy(unicode.IsLetter),
			input:         "1",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "1",
		},
		{
			name:          "parsing empty input should succeed",
			parser:        Satisfy(unicode.IsLetter),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
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

func BenchmarkSatisfy(b *testing.B) {
	parser := Satisfy(unicode.IsLetter)
	input := NewInputFromString("a")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestSpace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single space char from single space input should succeed",
			parser:        Space(),
			input:         " ",
			wantErr:       false,
			wantOutput:    rune(' '),
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from multiple char input should succeed",
			parser:        Space(),
			input:         " abc",
			wantErr:       false,
			wantOutput:    rune(' '),
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        Space(),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from single non-space input should fail",
			parser:        Space(),
			input:         "1",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "1",
		},
		{
			name:          "parsing single space from multiple non-space input should fail",
			parser:        Space(),
			input:         "123",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "123",
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

func BenchmarkSpace(b *testing.B) {
	parser := Space()
	input := NewInputFromString(" ")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestTab(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single space char from single space input should succeed",
			parser:        Tab(),
			input:         "\t",
			wantErr:       false,
			wantOutput:    rune('\t'),
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from multiple char input should succeed",
			parser:        Tab(),
			input:         "\tabc",
			wantErr:       false,
			wantOutput:    rune('\t'),
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        Tab(),
			input:         "",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from single non-space input should fail",
			parser:        Tab(),
			input:         "1",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "1",
		},
		{
			name:          "parsing single space from multiple non-space input should fail",
			parser:        Tab(),
			input:         "123",
			wantErr:       true,
			wantOutput:    rune(0),
			wantRemaining: "123",
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

func BenchmarkTab(b *testing.B) {
	parser := Tab()
	input := NewInputFromString("\t")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[int64]
		input         string
		wantErr       bool
		wantOutput    int64
		wantRemaining string
	}{
		{
			name:          "parsing positive integer should succeed",
			parser:        Int64(),
			input:         "123",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "",
		},
		{
			name:          "parsing negative integer should succeed",
			parser:        Int64(),
			input:         "-123",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        Int64(),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing negative integer should succeed",
			parser:        Int64(),
			input:         "-123abc",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        Int64(),
			input:         "9223372036854775808", // max int64 + 1
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "9223372036854775808",
		},
		{
			name:          "parsing integer with invalid leading sign should fail",
			parser:        Int64(),
			input:         "!127",
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "!127",
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

func BenchmarkInt64(b *testing.B) {
	parser := Int64()
	input := NewInputFromString("123")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestInt8(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[int8]
		input         string
		wantErr       bool
		wantOutput    int8
		wantRemaining string
	}{
		{
			name:          "parsing positive integer should succeed",
			parser:        Int8(),
			input:         "123",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "",
		},
		{
			name:          "parsing negative integer should succeed",
			parser:        Int8(),
			input:         "-123",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        Int8(),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing negative integer should succeed",
			parser:        Int8(),
			input:         "-123abc",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        Int8(),
			input:         "128", // max int8 + 1
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "128",
		},
		{
			name:          "parsing integer with invalid leading sign should fail",
			parser:        Int8(),
			input:         "!127",
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "!127",
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

func BenchmarkInt8(b *testing.B) {
	parser := Int8()
	input := NewInputFromString("123")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}

func TestUInt8(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        Parser[uint8]
		input         string
		wantErr       bool
		wantOutput    uint8
		wantRemaining string
	}{
		{
			name:          "parsing positive integer should succeed",
			parser:        UInt8(),
			input:         "253",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        UInt8(),
			input:         "253abc",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        UInt8(),
			input:         "256", // max uint8 + 1
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "256",
		},
		{
			name:          "parsing empty input should succeed",
			parser:        UInt8(),
			input:         "",
			wantErr:       true,
			wantOutput:    0,
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

func BenchmarkUInt8(b *testing.B) {
	parser := UInt8()
	input := NewInputFromString("253")

	for i := 0; i < b.N; i++ {
		parser(input)
	}
}
