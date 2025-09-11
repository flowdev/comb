package cmb_test

import (
	"math"
	"testing"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
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
			parser:        cmb.Int64(false, 10),
			input:         "123",
			wantErr:       false,
			wantOutput:    123,
			wantRemaining: "",
		}, {
			name:          "parsing negative integer should succeed",
			parser:        cmb.Int64(true, 10),
			input:         "-123",
			wantErr:       false,
			wantOutput:    -123,
			wantRemaining: "",
		}, {
			name:          "parsing positive integer prefix should succeed",
			parser:        cmb.Int64(false, 0),
			input:         "01_2_3abc",
			wantErr:       false,
			wantOutput:    01_2_3,
			wantRemaining: "abc",
		}, {
			name:          "parsing negative integer prefix should succeed",
			parser:        cmb.Int64(true, 0),
			input:         "-0x1a3ghi",
			wantErr:       false,
			wantOutput:    -0x1a3,
			wantRemaining: "ghi",
		}, {
			name:          "parsing binary integer should succeed",
			parser:        cmb.Int64(true, 0),
			input:         "-0b101",
			wantErr:       false,
			wantOutput:    -0b101,
			wantRemaining: "",
		}, {
			name:          "parsing octal integer should succeed",
			parser:        cmb.Int64(true, 0),
			input:         "-0o171",
			wantErr:       false,
			wantOutput:    -0o171,
			wantRemaining: "",
		}, {
			name:          "parsing binary integer should succeed",
			parser:        cmb.Int64(true, 2),
			input:         "-10",
			wantErr:       false,
			wantOutput:    -2,
			wantRemaining: "",
		}, {
			name:          "parsing octal integer should succeed",
			parser:        cmb.Int64(true, 8),
			input:         "+13",
			wantErr:       false,
			wantOutput:    11,
			wantRemaining: "",
		}, {
			name:          "parsing hex integer should succeed",
			parser:        cmb.Int64(true, 16),
			input:         "+1f",
			wantErr:       false,
			wantOutput:    31,
			wantRemaining: "",
		}, {
			name:          "parsing integer of base 20 should succeed",
			parser:        cmb.Int64(true, 20),
			input:         "+1j",
			wantErr:       false,
			wantOutput:    39,
			wantRemaining: "",
		}, {
			name:          "parsing overflowing integer should fail",
			parser:        cmb.Int64(true, 10),
			input:         "9223372036854775808", // max int64 + 1
			wantErr:       true,
			wantOutput:    9223372036854775807,
			wantRemaining: "",
		}, {
			name:          "parsing integer with invalid leading sign should fail",
			parser:        cmb.Int64(true, 10),
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

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
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
	parser := cmb.Int64(false, 10)
	input := comb.NewFromString("123", 0)

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
			parser:        cmb.UInt64(false, 0),
			input:         "253",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "",
		},
		{
			name:          "parsing positive integer prefix should succeed",
			parser:        cmb.UInt64(true, 10),
			input:         "253abc",
			wantErr:       false,
			wantOutput:    253,
			wantRemaining: "abc",
		},
		{
			name:          "parsing overflowing integer should fail",
			parser:        cmb.UInt64(true, 10),
			input:         "18446744073709551616", // max uint64 + 1
			wantErr:       true,
			wantOutput:    math.MaxUint64,
			wantRemaining: "",
		},
		{
			name:          "parsing empty input should fail",
			parser:        cmb.UInt64(true, 10),
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

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
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
	parser := cmb.UInt64(false, 10)
	input := comb.NewFromString("253", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}

func TestFloat64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[float64]
		input         string
		wantErr       bool
		wantOutput    float64
		wantRemaining string
	}{
		{
			name:          "parsing positive float should succeed",
			parser:        cmb.Float64(false, 10, true),
			input:         "12.3",
			wantErr:       false,
			wantOutput:    12.3,
			wantRemaining: "",
		}, {
			name:          "parsing negative float should succeed",
			parser:        cmb.Float64(true, 10, true),
			input:         "-.123",
			wantErr:       false,
			wantOutput:    -.123,
			wantRemaining: "",
		}, {
			name:          "parsing integer in non-strict mode should succeed",
			parser:        cmb.Float64(false, 10, false),
			input:         "123_",
			wantErr:       false,
			wantOutput:    123.0,
			wantRemaining: "_",
		}, {
			name:          "parsing integer in strict mode should fail",
			parser:        cmb.Float64(false, 10, true),
			input:         "123",
			wantErr:       true,
			wantOutput:    0.0,
			wantRemaining: "123",
		}, {
			name:          "parsing positive float prefix should succeed",
			parser:        cmb.Float64(false, 0, true),
			input:         "0x1_2.p3abc",
			wantErr:       false,
			wantOutput:    0x1_2.p3,
			wantRemaining: "abc",
		}, {
			name:          "parsing negative float prefix should succeed",
			parser:        cmb.Float64(true, 0, true),
			input:         "-1.2_3e4abc",
			wantErr:       false,
			wantOutput:    -1.2_3e4,
			wantRemaining: "abc",
		}, {
			name:          "parsing wild hex float should succeed",
			parser:        cmb.Float64(true, 16, true),
			input:         "-.2p3",
			wantErr:       false,
			wantOutput:    -0x.2p3,
			wantRemaining: "",
		}, {
			name:          "parsing wilder hex float should succeed",
			parser:        cmb.Float64(true, 0, true),
			input:         "-0x.2p3",
			wantErr:       false,
			wantOutput:    -0x.2p3,
			wantRemaining: "",
		}, {
			name:          "parsing overflowing float should fail",
			parser:        cmb.Float64(true, 10, false),
			input:         "1.79769313486231589999999999999999999999999e+308", // max float64 + a little
			wantErr:       true,
			wantOutput:    math.Inf(1),
			wantRemaining: "1.79769313486231589999999999999999999999999e+308",
		}, {
			name:          "parsing float with invalid leading sign should fail",
			parser:        cmb.Float64(true, 10, false),
			input:         "!1.27",
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "!1.27",
		}, {
			name:          "sign only should fail",
			parser:        cmb.Float64(true, 10, false),
			input:         "-",
			wantErr:       true,
			wantOutput:    0,
			wantRemaining: "-",
		}, {
			name:          "empty input should fail",
			parser:        cmb.Float64(true, 10, false),
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

			newState, gotResult, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			t.Logf("got error: %v", gotErr)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("got error %v, want error: %t", gotErr, tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %f, want output %f", gotResult, tc.wantOutput)
			}

			remainingString := newState.CurrentString()
			if remainingString != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", remainingString, tc.wantRemaining)
			}
		})
	}
}

func BenchmarkFloat64(b *testing.B) {
	parser := cmb.Float64(false, 10, true)
	input := comb.NewFromString("1.23", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parser.Parse(input)
	}
}
