package cmb

import (
	"github.com/flowdev/comb"
	"math"
	"testing"
)

func TestInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[int64]
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

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, true, 0))
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
	input := comb.NewFromString("123", false, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestUInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[uint64]
		input         string
		wantErr       bool
		wantOutput    uint64
		wantRemaining string
	}{
		{
			name:          "parsing positive integer should succeed",
			parser:        UInt64(false, 0),
			input:         "253",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        UInt64(true, 10),
			input:         "253abc",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        UInt64(true, 10),
			input:         "18446744073709551616", // max uint64 + 1
			wantErr:       true,
			wantOutput:    math.MaxUint64,
			wantRemaining: "",
		},
		{
			name:          "parsing empty input should fail",
			parser:        UInt64(true, 10),
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

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, true, 0))
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

func BenchmarkUInt64(b *testing.B) {
	parser := UInt64(false, 10)
	input := comb.NewFromString("253", false, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}
