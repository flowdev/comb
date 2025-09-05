package cmb_test

import (
	"bytes"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
)

func TestChar(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing char from single char input should succeed",
			parser:        cmb.Char('a'),
			input:         "a",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "",
		}, {
			name:          "parsing valid char in longer input should succeed",
			parser:        cmb.Char('a'),
			input:         "abc",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "bc",
		}, {
			name:          "parsing wrong char input should fail",
			parser:        cmb.Char('a'),
			input:         "123",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "123",
		}, {
			name:          "parsing non-valid Unicode char should fail",
			parser:        cmb.Char('a'),
			input:         string([]byte{129, 65, 66, 67}),
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: string([]byte{129, 65, 66, 67}),
		}, {
			name:          "parsing empty input should fail",
			parser:        cmb.Char('a'),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotOutput, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			if gotOutput != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotOutput, tc.wantOutput)
			}

			gotRemaining := newState.CurrentString()
			if gotRemaining != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", gotRemaining, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkChar(b *testing.B) {
	parser := cmb.Char('a')
	input := comb.NewFromString("a", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestAnyChar(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing char from single char input should succeed",
			parser:        cmb.AnyChar(),
			input:         "a",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "",
		}, {
			name:          "parsing valid char in longer input should succeed",
			parser:        cmb.AnyChar(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "bc",
		}, {
			name:          "parsing non-valid Unicode char should fail",
			parser:        cmb.AnyChar(),
			input:         string([]byte{129, 65, 66, 67}),
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: string([]byte{129, 65, 66, 67}),
		}, {
			name:          "parsing empty input should fail",
			parser:        cmb.AnyChar(),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotOutput, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			if gotOutput != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotOutput, tc.wantOutput)
			}

			gotRemaining := newState.CurrentString()
			if gotRemaining != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", gotRemaining, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkAnyChar(b *testing.B) {
	parser := cmb.AnyChar()
	input := comb.NewFromString("a", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestByte(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[byte]
		input         []byte
		wantErr       bool
		wantOutput    byte
		wantRemaining []byte
	}{
		{
			name:          "parsing char from single char input should succeed",
			parser:        cmb.Byte(7),
			input:         []byte{7},
			wantErr:       false,
			wantOutput:    byte(7),
			wantRemaining: []byte{},
		}, {
			name:          "parsing byte in longer input should succeed",
			parser:        cmb.Byte(7),
			input:         []byte{7, 8, 9},
			wantErr:       false,
			wantOutput:    byte(7),
			wantRemaining: []byte{8, 9},
		}, {
			name:          "parsing wrong byte should fail",
			parser:        cmb.Byte(7),
			input:         []byte{8, 9},
			wantErr:       true,
			wantOutput:    byte(0),
			wantRemaining: []byte{8, 9},
		}, {
			name:          "parsing empty input should fail",
			parser:        cmb.Byte(7),
			input:         []byte{},
			wantErr:       true,
			wantOutput:    byte(0),
			wantRemaining: []byte{},
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotOutput, gotErr := tc.parser.Parse(comb.NewFromBytes(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			if gotOutput != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotOutput, tc.wantOutput)
			}

			gotRemaining := newState.CurrentBytes()
			if !bytes.Equal(gotRemaining, tc.wantRemaining) {
				t.Errorf("got remaining %#v, want remaining %#v", gotRemaining, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkByte(b *testing.B) {
	parser := cmb.Byte(7)
	input := comb.NewFromBytes([]byte{7}, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestAnyByte(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[byte]
		input         []byte
		wantErr       bool
		wantOutput    byte
		wantRemaining []byte
	}{
		{
			name:          "parsing char from single char input should succeed",
			parser:        cmb.AnyByte(),
			input:         []byte{7},
			wantErr:       false,
			wantOutput:    byte(7),
			wantRemaining: []byte{},
		}, {
			name:          "parsing valid char in longer input should succeed",
			parser:        cmb.AnyByte(),
			input:         []byte{7, 8, 9},
			wantErr:       false,
			wantOutput:    byte(7),
			wantRemaining: []byte{8, 9},
		}, {
			name:          "parsing empty input should fail",
			parser:        cmb.AnyByte(),
			input:         []byte{},
			wantErr:       true,
			wantOutput:    byte(0),
			wantRemaining: []byte{},
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotOutput, gotErr := tc.parser.Parse(comb.NewFromBytes(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			if gotOutput != tc.wantOutput {
				t.Errorf("got output %q, want output %q", gotOutput, tc.wantOutput)
			}

			gotRemaining := newState.CurrentBytes()
			if !bytes.Equal(gotRemaining, tc.wantRemaining) {
				t.Errorf("got remaining %#v, want remaining %#v", gotRemaining, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkAnyByte(b *testing.B) {
	parser := cmb.AnyByte()
	input := comb.NewFromBytes([]byte("a"), 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestAlpha0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alpha input should succeed",
			parser:        cmb.Alpha0(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alpha input should succeed",
			parser:        cmb.Alpha0(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars until terminating char should succeed",
			parser:        cmb.Alpha0(),
			input:         "abc123",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "123",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        cmb.Alpha0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non alpha chars should succeed",
			parser:        cmb.Alpha0(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "123",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkAlpha0(b *testing.B) {
	parser := cmb.Alpha0()
	input := comb.NewFromString("abc", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestAlpha1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alpha input should succeed",
			parser:        cmb.Alpha1(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alpha input should succeed",
			parser:        cmb.Alpha1(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars until terminating char should succeed",
			parser:        cmb.Alpha1(),
			input:         "abc123",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "123",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        cmb.Alpha1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with an alpha char should fail",
			parser:        cmb.Alpha1(),
			input:         "1c",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "1c",
		},
		{
			name:          "parsing non alpha chars should fail",
			parser:        cmb.Alpha1(),
			input:         "123",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkAlpha1(b *testing.B) {
	parser := cmb.Alpha1()
	input := comb.NewFromString("abc", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestDigit0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single digit char from single digit input should succeed",
			parser:        cmb.Digit0(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple digit input should succeed",
			parser:        cmb.Digit0(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        cmb.Digit0(),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        cmb.Digit0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non digit chars should succeed",
			parser:        cmb.Digit0(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "abc",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkDigit0(b *testing.B) {
	parser := cmb.Digit0()
	input := comb.NewFromString("123", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestDigit1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single digit char from single digit input should succeed",
			parser:        cmb.Digit1(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple digit input should succeed",
			parser:        cmb.Digit1(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        cmb.Digit1(),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        cmb.Digit1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with an digit char should fail",
			parser:        cmb.Digit1(),
			input:         "c1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "c1",
		},
		{
			name:          "parsing non digit chars should fail",
			parser:        cmb.Digit1(),
			input:         "abc",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkDigit1(b *testing.B) {
	parser := cmb.Digit1()
	input := comb.NewFromString("123", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestHexDigit0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single hex digit char from single hex digit input should succeed",
			parser:        cmb.HexDigit0(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars from multiple hex digit input should succeed",
			parser:        cmb.HexDigit0(),
			input:         "1f3",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars until terminating char should succeed",
			parser:        cmb.HexDigit0(),
			input:         "1f3z",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "z",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        cmb.HexDigit0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non hex digit chars should succeed",
			parser:        cmb.HexDigit0(),
			input:         "ghi",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "ghi",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkHexDigit0(b *testing.B) {
	parser := cmb.HexDigit0()
	input := comb.NewFromString("1f3", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestHexDigit1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single hex digit char from single hex digit input should succeed",
			parser:        cmb.HexDigit1(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars from multiple hex digit input should succeed",
			parser:        cmb.HexDigit1(),
			input:         "1f3",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "",
		},
		{
			name:          "parsing hex digit chars until terminating char should succeed",
			parser:        cmb.HexDigit1(),
			input:         "1f3ghi",
			wantErr:       false,
			wantOutput:    "1f3",
			wantRemaining: "ghi",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        cmb.HexDigit1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with a hex digit char should fail",
			parser:        cmb.HexDigit1(),
			input:         "h1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "h1",
		},
		{
			name:          "parsing non hex digit chars should fail",
			parser:        cmb.HexDigit1(),
			input:         "ghi",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "ghi",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkHexDigit1(b *testing.B) {
	parser := cmb.HexDigit1()
	input := comb.NewFromString("1f3", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestWhitespace0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single whitespace from single ' ' input should succeed",
			parser:        cmb.Whitespace0(),
			input:         " ",
			wantErr:       false,
			wantOutput:    " ",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\t' input should succeed",
			parser:        cmb.Whitespace0(),
			input:         "\t",
			wantErr:       false,
			wantOutput:    "\t",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\n' input should succeed",
			parser:        cmb.Whitespace0(),
			input:         "\n",
			wantErr:       false,
			wantOutput:    "\n",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\r' input should succeed",
			parser:        cmb.Whitespace0(),
			input:         "\r",
			wantErr:       false,
			wantOutput:    "\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars input should succeed",
			parser:        cmb.Whitespace0(),
			input:         " \t\n\r",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars with suffix input should succeed",
			parser:        cmb.Whitespace0(),
			input:         " \t\n\rabc",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        cmb.Whitespace0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing a single non-whitespace char input should succeed",
			parser:        cmb.Whitespace0(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "a",
		},
		{
			name:          "parsing input starting with a non-whitespace char should succeed",
			parser:        cmb.Whitespace0(),
			input:         "a \t\n\r",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "a \t\n\r",
		},
		{
			name:          "parsing non-whitespace chars should succeed",
			parser:        cmb.Whitespace0(),
			input:         "ghi",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "ghi",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkWhitespace0(b *testing.B) {
	b.ReportAllocs()
	parser := cmb.Whitespace0()
	input := comb.NewFromString(" \t\n\r", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestWhitespace1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single whitespace from single ' ' input should succeed",
			parser:        cmb.Whitespace1(),
			input:         " ",
			wantErr:       false,
			wantOutput:    " ",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\t' input should succeed",
			parser:        cmb.Whitespace1(),
			input:         "\t",
			wantErr:       false,
			wantOutput:    "\t",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\n' input should succeed",
			parser:        cmb.Whitespace1(),
			input:         "\n",
			wantErr:       false,
			wantOutput:    "\n",
			wantRemaining: "",
		},
		{
			name:          "parsing single whitespace from single '\r' input should succeed",
			parser:        cmb.Whitespace1(),
			input:         "\r",
			wantErr:       false,
			wantOutput:    "\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars input should succeed",
			parser:        cmb.Whitespace1(),
			input:         " \t\n\r",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple whitespace chars from multiple whitespace chars with suffix input should succeed",
			parser:        cmb.Whitespace1(),
			input:         " \t\n\rabc",
			wantErr:       false,
			wantOutput:    " \t\n\r",
			wantRemaining: "abc",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        cmb.Whitespace1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing a single non-whitespace char input should fail",
			parser:        cmb.Whitespace1(),
			input:         "a",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "a",
		},
		{
			name:          "parsing input starting with a non-whitespace char should fail",
			parser:        cmb.Whitespace1(),
			input:         "a \t\n\r",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "a \t\n\r",
		},
		{
			name:          "parsing non-whitespace chars should fail",
			parser:        cmb.Whitespace1(),
			input:         "ghi",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "ghi",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkWhitespace1(b *testing.B) {
	b.ReportAllocs()
	input := comb.NewFromString(" \t\n\r", 0)
	parser := cmb.Whitespace1()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestAlphanumeric0(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alphanumerical input should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing single digit char from single alphanumerical input should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alphanumerical input should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple alphanumerical input should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing multiple alphanumerical input should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "a1b2c3",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars until terminating char should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "abc$%^",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "123$%^",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing alphanumerical chars until terminating char should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "a1b2c3$%^",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing an empty input should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing non alphanumerical chars should succeed",
			parser:        cmb.Alphanumeric0(),
			input:         "$%^",
			wantErr:       false,
			wantOutput:    "",
			wantRemaining: "$%^",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkAlphanumeric0(b *testing.B) {
	parser := cmb.Alphanumeric0()
	input := comb.NewFromString("a1b2c3", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestAlphanumeric1(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char from single alphanumerical input should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "a",
			wantErr:       false,
			wantOutput:    "a",
			wantRemaining: "",
		},
		{
			name:          "parsing single digit char from single alphanumerical input should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "1",
			wantErr:       false,
			wantOutput:    "1",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars from multiple alphanumerical input should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "abc",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:          "parsing digit chars from multiple alphanumerical input should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "123",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:          "parsing alphanumerical chars from multiple alphanumerical input should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "a1b2c3",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "",
		},
		{
			name:          "parsing alpha chars until terminating char should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "abc$%^",
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing digit chars until terminating char should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "123$%^",
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing alphanumerical chars until terminating char should succeed",
			parser:        cmb.Alphanumeric1(),
			input:         "a1b2c3$%^",
			wantErr:       false,
			wantOutput:    "a1b2c3",
			wantRemaining: "$%^",
		},
		{
			name:          "parsing an empty input should fail",
			parser:        cmb.Alphanumeric1(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing input not starting with an alphanumeric char should fail",
			parser:        cmb.Alphanumeric1(),
			input:         "$1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$1",
		},
		{
			name:          "parsing non digit chars should fail",
			parser:        cmb.Alphanumeric1(),
			input:         "$%^",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$%^",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkAlphanumeric1(b *testing.B) {
	parser := cmb.Alphanumeric1()
	input := comb.NewFromString("a1b2c3", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestLF(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single line-feed char from single line-feed input should succeed",
			parser:        cmb.LF(),
			input:         "\n",
			wantErr:       false,
			wantOutput:    '\n',
			wantRemaining: "",
		},
		{
			name:          "parsing single line-feed char from multiple char input should succeed",
			parser:        cmb.LF(),
			input:         "\nabc",
			wantErr:       false,
			wantOutput:    '\n',
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        cmb.LF(),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
		{
			name:          "parsing single line-feed char from single non-line-feed input should fail",
			parser:        cmb.LF(),
			input:         "1",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "1",
		},
		{
			name:          "parsing single line-feed from multiple non-line-feed input should fail",
			parser:        cmb.LF(),
			input:         "123",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "123",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkLF(b *testing.B) {
	parser := cmb.LF()
	input := comb.NewFromString("\n", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestCR(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single carriage-return char from single carriage-return input should succeed",
			parser:        cmb.CR(),
			input:         "\r",
			wantErr:       false,
			wantOutput:    '\r',
			wantRemaining: "",
		},
		{
			name:          "parsing single carriage-return char from multiple char input should succeed",
			parser:        cmb.CR(),
			input:         "\rabc",
			wantErr:       false,
			wantOutput:    '\r',
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        cmb.CR(),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
		{
			name:          "parsing single carriage-return char from single non-carriage-return input should fail",
			parser:        cmb.CR(),
			input:         "1",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "1",
		},
		{
			name:          "parsing single carriage-return from multiple non-carriage-return input should fail",
			parser:        cmb.CR(),
			input:         "123",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "123",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkCR(b *testing.B) {
	parser := cmb.CR()
	input := comb.NewFromString("\r", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestCRLF(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing single CRLF char from single CRLF input should succeed",
			parser:        cmb.CRLF(),
			input:         "\r\n",
			wantErr:       false,
			wantOutput:    "\r\n",
			wantRemaining: "",
		},
		{
			name:          "parsing single CRLF char from multiple char input should succeed",
			parser:        cmb.CRLF(),
			input:         "\r\nabc",
			wantErr:       false,
			wantOutput:    "\r\n",
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        cmb.CRLF(),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:          "parsing incomplete CRLF input should fail",
			parser:        cmb.CRLF(),
			input:         "\r",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "\r",
		},
		{
			name:          "parsing single CRLF char from single non-CRLF input should fail",
			parser:        cmb.CRLF(),
			input:         "1",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "1",
		},
		{
			name:          "parsing single CRLF from multiple non-CRLF input should fail",
			parser:        cmb.CRLF(),
			input:         "123",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "123",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkCRLF(b *testing.B) {
	parser := cmb.CRLF()
	input := comb.NewFromString("\r\n", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestOneOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing matched char should succeed",
			parser:        cmb.OneOfRunes('a', '1', '+'),
			input:         "+",
			wantErr:       false,
			wantOutput:    '+',
			wantRemaining: "",
		},
		{
			name:          "parsing input not containing any of the sought chars should fail",
			parser:        cmb.OneOfRunes('a', '1', '+'),
			input:         "b",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "b",
		},
		{
			name:          "parsing empty input should fail",
			parser:        cmb.OneOfRunes('a', '1', '+'),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkOneOf(b *testing.B) {
	parser := cmb.OneOfRunes('a', '1', '+')
	input := comb.NewFromString("+", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestSatisfy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single alpha char satisfying constraint should succeed",
			parser:        cmb.Satisfy("letter", unicode.IsLetter),
			input:         "a",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "",
		},
		{
			name:          "parsing alpha char satisfying constraint from mixed input should succeed",
			parser:        cmb.Satisfy("letter", unicode.IsLetter),
			input:         "a1",
			wantErr:       false,
			wantOutput:    'a',
			wantRemaining: "1",
		},
		{
			name:          "parsing char not satisfying constraint should succeed",
			parser:        cmb.Satisfy("letter", unicode.IsLetter),
			input:         "1",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "1",
		},
		{
			name:          "parsing empty input should succeed",
			parser:        cmb.Satisfy("letter", unicode.IsLetter),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkSatisfy(b *testing.B) {
	parser := cmb.Satisfy("letter", unicode.IsLetter)
	input := comb.NewFromString("a", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestSpace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single space char from single space input should succeed",
			parser:        cmb.Space(),
			input:         " ",
			wantErr:       false,
			wantOutput:    ' ',
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from multiple char input should succeed",
			parser:        cmb.Space(),
			input:         " abc",
			wantErr:       false,
			wantOutput:    ' ',
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        cmb.Space(),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from single non-space input should fail",
			parser:        cmb.Space(),
			input:         "1",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "1",
		},
		{
			name:          "parsing single space from multiple non-space input should fail",
			parser:        cmb.Space(),
			input:         "123",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "123",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkSpace(b *testing.B) {
	parser := cmb.Space()
	input := comb.NewFromString(" ", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestTab(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[rune]
		input         string
		wantErr       bool
		wantOutput    rune
		wantRemaining string
	}{
		{
			name:          "parsing single space char from single space input should succeed",
			parser:        cmb.Tab(),
			input:         "\t",
			wantErr:       false,
			wantOutput:    '\t',
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from multiple char input should succeed",
			parser:        cmb.Tab(),
			input:         "\tabc",
			wantErr:       false,
			wantOutput:    '\t',
			wantRemaining: "abc",
		},
		{
			name:          "parsing empty input should fail",
			parser:        cmb.Tab(),
			input:         "",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "",
		},
		{
			name:          "parsing single space char from single non-space input should fail",
			parser:        cmb.Tab(),
			input:         "1",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "1",
		},
		{
			name:          "parsing single space from multiple non-space input should fail",
			parser:        cmb.Tab(),
			input:         "123",
			wantErr:       true,
			wantOutput:    utf8.RuneError,
			wantRemaining: "123",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkTab(b *testing.B) {
	parser := cmb.Tab()
	input := comb.NewFromString("\t", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[string]
		input         string
		wantErr       bool
		wantOutput    string
		wantRemaining string
	}{
		{
			name:          "parsing a token from an input starting with it should succeed",
			parser:        cmb.String("Bonjour"),
			input:         "Bonjour tout le monde",
			wantErr:       false,
			wantOutput:    "Bonjour",
			wantRemaining: " tout le monde",
		},
		{
			name:          "parsing a token from a non-matching input should fail",
			parser:        cmb.String("Bonjour"),
			input:         "Hello tout le monde",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "Hello tout le monde",
		},
		{
			name:          "parsing a token from an empty input should fail",
			parser:        cmb.String("Bonjour"),
			input:         "",
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkToken(b *testing.B) {
	parser := cmb.String("Bonjour")
	input := comb.NewFromString("Bonjour tout le monde", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestSatisfyMN(t *testing.T) {
	t.Parallel()

	type args struct {
		p comb.Parser[string]
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
				p: cmb.SatisfyMN("letter", 3, 6, unicode.IsLetter),
			},
			wantErr:       false,
			wantOutput:    "latin",
			wantRemaining: "123",
		},
		{
			name:  "parsing input longer than atLeast and atMost should succeed",
			input: "lengthy",
			args: args{
				p: cmb.SatisfyMN("letter", 3, 6, unicode.IsLetter),
			},
			wantErr:       false,
			wantOutput:    "length",
			wantRemaining: "y",
		},
		{
			name:  "parsing input longer than atLeast and shorter than atMost should succeed",
			input: "latin",
			args: args{
				p: cmb.SatisfyMN("letter", 3, 6, unicode.IsLetter),
			},
			wantErr:       false,
			wantOutput:    "latin",
			wantRemaining: "",
		},
		{
			name:  "parsing empty input should fail",
			input: "",
			args: args{
				p: cmb.SatisfyMN("letter", 3, 6, unicode.IsLetter),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
		{
			name:  "parsing too short input should fail",
			input: "ed",
			args: args{
				p: cmb.SatisfyMN("letter", 3, 6, unicode.IsLetter),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "ed",
		},
		{
			name:  "parsing with non-matching predicate should fail",
			input: "12345",
			args: args{
				p: cmb.SatisfyMN("letter", 3, 6, unicode.IsLetter),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "12345",
		},
	}
	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.args.p.Parse(comb.NewFromString(tc.input, 10))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkSatisfyMN(b *testing.B) {
	p := cmb.SatisfyMN("letter", 3, 6, cmb.IsDigit)
	input := comb.NewFromString("13579", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = p.Parse(input)
	}
}
