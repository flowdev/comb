package pcb

import (
	"github.com/oleiade/gomme"
	"testing"
)

func TestInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[int64]
		input         string
		wantErr       bool
		wantOutput    int64
		wantRemaining string
	}{
		{
			name:          "parsing positive integer should succeed",
			parser:        Int64(false, 10),
			input:         "123",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "",
		},
		{
			name:          "parsing negative integer should succeed",
			parser:        Int64(true, 10),
			input:         "-123",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        Int64(false, 0),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing negative integer prefix should succeed",
			parser:        Int64(true, 0),
			input:         "-123abc",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        Int64(true, 10),
			input:         "9223372036854775808", // max int64 + 1
			wantErr:       true,
			wantOutput:    9223372036854775807,
			wantRemaining: "",
		},
		{
			name:          "parsing integer with invalid leading sign should fail",
			parser:        Int64(true, 10),
			input:         "!127",
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "!127",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(gomme.NewFromString(tc.input, true))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkInt64(b *testing.B) {
	parser := Int64(false, 10)
	input := gomme.NewFromString("123", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestInt8(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[int8]
		input         string
		wantErr       bool
		wantOutput    int8
		wantRemaining string
	}{
		{
			name:          "parsing positive integer should succeed",
			parser:        Int8(false, 10),
			input:         "123",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "",
		},
		{
			name:          "parsing negative integer should succeed",
			parser:        Int8(true, 10),
			input:         "-123",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        Int8(false, 0),
			input:         "123abc",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing negative integer should succeed",
			parser:        Int8(true, 0),
			input:         "-123abc",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        Int8(true, 10),
			input:         "128", // max int8 + 1
			wantErr:       true,
			wantOutput:    127,
			wantRemaining: "",
		},
		{
			name:          "parsing integer with invalid leading sign should fail",
			parser:        Int8(true, 10),
			input:         "!127",
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "!127",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(gomme.NewFromString(tc.input, true))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkInt8(b *testing.B) {
	parser := Int8(false, 10)
	input := gomme.NewFromString("123", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestUInt8(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        gomme.Parser[uint8]
		input         string
		wantErr       bool
		wantOutput    uint8
		wantRemaining string
	}{
		{
			name:          "parsing positive integer should succeed",
			parser:        UInt8(false, 0),
			input:         "253",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        UInt8(true, 10),
			input:         "253abc",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        UInt8(true, 10),
			input:         "256", // max uint8 + 1
			wantErr:       true,
			wantOutput:    255,
			wantRemaining: "",
		},
		{
			name:          "parsing empty input should fail",
			parser:        UInt8(true, 10),
			input:         "",
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotResult, gotErr := tc.parser.Parse(gomme.NewFromString(tc.input, true))
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
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

func BenchmarkUInt8(b *testing.B) {
	parser := UInt8(false, 10)
	input := gomme.NewFromString("253", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}
