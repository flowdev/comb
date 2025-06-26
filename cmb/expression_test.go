package cmb_test

import (
	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
	"testing"
)

func TestExpression_HappyPath(t *testing.T) {
	testCases := []struct {
		name          string
		parser        comb.Parser[int64]
		input         string
		wantOutput    int64
		wantRemaining string
	}{
		{
			name:          "just value",
			parser:        cmb.Expression(cmb.Int64(false, 10)).Parser(),
			input:         "123",
			wantOutput:    123,
			wantRemaining: "",
		}, {
			name: "prefix op",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op:       "-",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return -i
					},
				},
			})).Parser(),
			input:         "-123abc",
			wantOutput:    -123,
			wantRemaining: "abc",
		}, {
			name: "infix op",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "+",
					SafeSpot: false,
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			})).Parser(),
			input:         "123+456!",
			wantOutput:    579,
			wantRemaining: "!",
		}, {
			name: "postfix op",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op:       "++",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return i + 1
					},
				},
			})).Parser(),
			input:         "123++",
			wantOutput:    124,
			wantRemaining: "",
		}, {
			name: "multi prefix ops",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op:       "-",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return -i
					},
				}, {
					Op:       "+",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return i + 1
					},
				},
			})).Parser(),
			input:         " + - 123",
			wantOutput:    -122,
			wantRemaining: "",
		}, {
			name: "multi infix ops",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "+",
					SafeSpot: false,
					Fn: func(a, b int64) int64 {
						return a + b
					},
				}, {
					Op:       "-",
					SafeSpot: false,
					Fn: func(a, b int64) int64 {
						return a - b
					},
				},
			})).Parser(),
			input:         " 1 + 2 - 3 + 4",
			wantOutput:    4,
			wantRemaining: "",
		}, {
			name: "multi postfix ops",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op:       "-",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return -i
					},
				}, {
					Op:       "+",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return i + 1
					},
				},
			})).Parser(),
			input:         " \t 123 - \t +",
			wantOutput:    -122,
			wantRemaining: "",
		}, {
			name: "multi level infix ops",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "*",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a * b
					},
				}, {
					Op:       "/",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a / b
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "-",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a - b
					},
				}, {
					Op:       "+",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			})).Parser(),
			input:         " \t 1 + 3 * \t 2 - 6 / 3 ag",
			wantOutput:    5,
			wantRemaining: " ag",
		}, {
			name: "parentheses and infix ops",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "*",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a * b
					},
				}, {
					Op:       "/",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a / b
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "-",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a - b
					},
				}, {
					Op:       "+",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			})).AddParentheses("(", ")").Parser(),
			input:         " \t( 1 + 3 ) * (\t 2 - 6 \t ) / 4",
			wantOutput:    -4,
			wantRemaining: "",
		}, {
			name: "all mixed up",
			parser: cmb.Expression(cmb.Int64(false, 10), cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op:       "-",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return -i
					},
				},
			}), cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op:       "--",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return i - 1
					},
				}, {
					Op:       "++",
					SafeSpot: false,
					Fn: func(i int64) int64 {
						return i + 1
					},
				},
			}), cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op:       "!",
					SafeSpot: false,
					Fn: func(v int64) int64 {
						r := int64(1)
						for i := int64(1); i <= v; i++ {
							r *= i
						}
						return r
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "^",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						r := int64(1)
						for i := int64(0); i < b; i++ {
							r *= a
						}
						return r
					},
				}, {
					Op:       "%",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a % b
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "*",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a * b
					},
				}, {
					Op:       "/",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a / b
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op:       "-",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a - b
					},
				}, {
					Op:       "+",
					SafeSpot: true,
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			})).AddParentheses("(", ")").AddParentheses("[", "]").Parser(),
			input:         "-  (\t ! 2 \t ++ + 3 --) * \t [ 2 ^ 2 - 12 % 6 ] / 4",
			wantOutput:    -8,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newState, gotOutput, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if gotErr != nil {
				t.Errorf("found error %v", gotErr)
			}

			if gotOutput != tc.wantOutput {
				t.Errorf("got output %d, want output %d", gotOutput, tc.wantOutput)
			}

			gotRemaining := newState.CurrentString()
			if gotRemaining != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", gotRemaining, tc.wantRemaining)
			}
		})
	}
}
