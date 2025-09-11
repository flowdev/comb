package cmb_test

import (
	"math"
	"slices"
	"testing"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
)

func TestExpression_IntHappyPath(t *testing.T) {
	t.Parallel()

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
			input:         "123 ",
			wantOutput:    123,
			wantRemaining: " ",
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
			})).WithSpace(cmb.Whitespace0()).Parser(),
			input:         "- 123 abc",
			wantOutput:    -123,
			wantRemaining: " abc",
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
			})).WithExpected("infix expression").Parser(),
			input:         "123+456 !",
			wantOutput:    579,
			wantRemaining: " !",
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
			input:         "123++ ",
			wantOutput:    124,
			wantRemaining: " ",
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
			})).AddParentheses("(", ")", true).Parser(),
			input:         " \t( 1 + 3 ) * (\t 2 - 6 \t ) / 4",
			wantOutput:    -4,
			wantRemaining: "",
		}, {
			name: "space parser",
			parser: cmb.Expression(cmb.Int64(false, 10)).AddPrefixLevel(cmb.PrefixOp[int64]{
				Op:       "-",
				SafeSpot: false,
				Fn: func(i int64) int64 {
					return -i
				},
			}).AddPostfixLevel(cmb.PostfixOp[int64]{
				Op:       "++",
				SafeSpot: false,
				Fn: func(i int64) int64 {
					return i + 1
				},
			}).AddInfixLevel(cmb.InfixOp[int64]{
				Op:       "*",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a * b
				},
			}).AddParentheses("[", "]", true).WithSpace(cmb.Whitespace1()).Parser(),
			input:         " \t-  [\t 2 * 3\t] ++ ",
			wantOutput:    -5,
			wantRemaining: " ",
		}, {
			name: "all mixed up",
			parser: cmb.Expression(cmb.Int64(false, 10)).AddPrefixLevel(cmb.PrefixOp[int64]{
				Op:       "-",
				SafeSpot: false,
				Fn: func(i int64) int64 {
					return -i
				},
			}).AddPostfixLevel(cmb.PostfixOp[int64]{
				Op:       "--",
				SafeSpot: false,
				Fn: func(i int64) int64 {
					return i - 1
				},
			}, cmb.PostfixOp[int64]{
				Op:       "++",
				SafeSpot: false,
				Fn: func(i int64) int64 {
					return i + 1
				},
			}).AddPrefixLevel(cmb.PrefixOp[int64]{
				Op:       "!",
				SafeSpot: false,
				Fn: func(v int64) int64 {
					r := int64(1)
					for i := int64(1); i <= v; i++ {
						r *= i
					}
					return r
				},
			}).AddInfixLevel(cmb.InfixOp[int64]{
				Op:       "^",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					r := int64(1)
					for i := int64(0); i < b; i++ {
						r *= a
					}
					return r
				},
			}, cmb.InfixOp[int64]{
				Op:       "%",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a % b
				},
			}).AddInfixLevel(cmb.InfixOp[int64]{
				Op:       "*",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a * b
				},
			}, cmb.InfixOp[int64]{
				Op:       "/",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a / b
				},
			}).AddInfixLevel(cmb.InfixOp[int64]{
				Op:       "-",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a - b
				},
			}, cmb.InfixOp[int64]{
				Op:       "+",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a + b
				},
			}).AddParentheses("(", ")", true).AddParentheses("[", "]", true).Parser(),
			input:         "-  (\t ! 2 \t ++ + 3 --) * \t [ 2 ^ 2 - 12 % 6 ] / 4",
			wantOutput:    -8,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

func TestExpression_FloatHappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		parser        comb.Parser[float64]
		input         string
		wantOutput    float64
		wantRemaining string
	}{
		{
			name:          "just value",
			parser:        cmb.Expression(cmb.Float64(false, 16, false)).Parser(),
			input:         "12p3 ",
			wantOutput:    0x12p3,
			wantRemaining: " ",
		}, {
			name: "multi level infix ops",
			parser: cmb.Expression(cmb.Float64(false, 10, false), cmb.InfixLevel([]cmb.InfixOp[float64]{
				{
					Op:       "*",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a * b
					},
				}, {
					Op:       "/",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a / b
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[float64]{
				{
					Op:       "-",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a - b
					},
				}, {
					Op:       "+",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a + b
					},
				},
			})).Parser(),
			input:         " \t 1 + 3 * \t 2 - 6 / 3 ag",
			wantOutput:    5,
			wantRemaining: " ag",
		}, {
			name: "parentheses and infix ops with strict floats",
			parser: cmb.Expression(cmb.Float64(false, 10, true), cmb.InfixLevel([]cmb.InfixOp[float64]{
				{
					Op:       "*",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a * b
					},
				}, {
					Op:       "/",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a / b
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[float64]{
				{
					Op:       "-",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a - b
					},
				}, {
					Op:       "+",
					SafeSpot: true,
					Fn: func(a, b float64) float64 {
						return a + b
					},
				},
			})).AddParentheses("(", ")", true).Parser(),
			input:         " \t( 1.0 + 3. ) * (\t 2. - 6.0 \t ) * .25",
			wantOutput:    -4.0,
			wantRemaining: "",
		}, {
			name: "all mixed up",
			parser: cmb.Expression(cmb.Float64(false, 10, true)).AddPrefixLevel(cmb.PrefixOp[float64]{
				Op:       "-",
				SafeSpot: false,
				Fn: func(i float64) float64 {
					return -i
				},
			}).AddPostfixLevel(cmb.PostfixOp[float64]{
				Op:       "--",
				SafeSpot: false,
				Fn: func(i float64) float64 {
					return i - 1
				},
			}, cmb.PostfixOp[float64]{
				Op:       "++",
				SafeSpot: false,
				Fn: func(i float64) float64 {
					return i + 1
				},
			}).AddPrefixLevel(cmb.PrefixOp[float64]{
				Op:       "!",
				SafeSpot: false,
				Fn: func(v float64) float64 {
					r := float64(1)
					for i := float64(1); i <= v; i++ {
						r *= i
					}
					return r
				},
			}).AddInfixLevel(cmb.InfixOp[float64]{
				Op:       "^",
				SafeSpot: true,
				Fn: func(a, b float64) float64 {
					r := float64(1)
					for i := float64(0); i < b; i++ {
						r *= a
					}
					return r
				},
			}, cmb.InfixOp[float64]{
				Op:       "%",
				SafeSpot: true,
				Fn: func(a, b float64) float64 {
					return math.Remainder(a, b)
				},
			}).AddInfixLevel(cmb.InfixOp[float64]{
				Op:       "*",
				SafeSpot: true,
				Fn: func(a, b float64) float64 {
					return a * b
				},
			}, cmb.InfixOp[float64]{
				Op:       "/",
				SafeSpot: true,
				Fn: func(a, b float64) float64 {
					return a / b
				},
			}).AddInfixLevel(cmb.InfixOp[float64]{
				Op:       "-",
				SafeSpot: true,
				Fn: func(a, b float64) float64 {
					return a - b
				},
			}, cmb.InfixOp[float64]{
				Op:       "+",
				SafeSpot: true,
				Fn: func(a, b float64) float64 {
					return a + b
				},
			}).AddParentheses("(", ")", true).AddParentheses("[", "]", true).Parser(),
			input:         "-  (\t ! 2. \t ++ + 3.0 --) * \t [ 2.0 ^ 2. - 12.000 % 6. ] / 4.0",
			wantOutput:    -8,
			wantRemaining: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newState, gotOutput, gotErr := tc.parser.Parse(comb.NewFromString(tc.input, 10))
			if gotErr != nil {
				t.Errorf("found error %v", gotErr)
			}

			if gotOutput != tc.wantOutput {
				t.Errorf("got output %f, want output %f", gotOutput, tc.wantOutput)
			}

			gotRemaining := newState.CurrentString()
			if gotRemaining != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", gotRemaining, tc.wantRemaining)
			}
		})
	}
}

func TestExpression_IntErrorCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[[]int64]
		input      string
		wantOutput []int64
		wantErrors int
	}{
		{
			name:       "additional character before value",
			parser:     cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Int64(true, 10))).Parser()),
			input:      "] -123",
			wantOutput: []int64{-123},
			wantErrors: 1,
		}, {
			name: "prefix op",
			parser: cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PrefixLevel([]cmb.PrefixOp[int64]{
					{
						Op:       "-",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return -i
						},
					},
				})).Parser()),
			input:      "! - | 123",
			wantOutput: []int64{-123},
			wantErrors: 2,
		}, {
			name: "infix op",
			parser: cmb.Count(2, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.InfixLevel([]cmb.InfixOp[int64]{
					{
						Op:       "+",
						SafeSpot: true,
						Fn: func(a, b int64) int64 {
							return a + b
						},
					},
				})).Parser()),
			input:      "(123)+ !=456",
			wantOutput: []int64{123, 456},
			wantErrors: 3,
		}, {
			name: "postfix op",
			parser: cmb.Count(2, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PostfixLevel([]cmb.PostfixOp[int64]{
					{
						Op:       "++",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i + 1
						},
					},
				})).Parser()),
			input:      "{123 ]++",
			wantOutput: []int64{123, 1},
			wantErrors: 2,
		}, {
			name: "multi prefix ops",
			parser: cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PrefixLevel([]cmb.PrefixOp[int64]{
					{
						Op:       "--",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i - 1
						},
					}, {
						Op:       "**",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i * 2
						},
					},
				})).Parser()),
			input:      "! -- { ** | 123",
			wantOutput: []int64{245},
			wantErrors: 3,
		}, {
			name: "multi infix ops",
			parser: cmb.Count(3, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.InfixLevel([]cmb.InfixOp[int64]{
					{
						Op:       "+",
						SafeSpot: true,
						Fn: func(a, b int64) int64 {
							return a + b
						},
					}, {
						Op:       "-",
						SafeSpot: true,
						Fn: func(a, b int64) int64 {
							return a - b
						},
					},
				})).Parser()),
			input:      "(1)+ !=2 [- ] 8",
			wantOutput: []int64{1, 2, -8},
			wantErrors: 5,
		}, {
			name: "multi postfix ops",
			parser: cmb.Count(3, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PostfixLevel([]cmb.PostfixOp[int64]{
					{
						Op:       "++",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i + 1
						},
					}, {
						Op:       "**",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i * 2
						},
					},
				})).Parser()),
			input:      "{123 ]++ | ++ **",
			wantOutput: []int64{123, 1, 2},
			wantErrors: 3,
		}, {
			name: "multi level infix ops",
			parser: cmb.Count(5, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PostfixLevel([]cmb.PostfixOp[int64]{
					{
						Op:       "++",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i + 1
						},
					},
				}),
				cmb.InfixLevel([]cmb.InfixOp[int64]{
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
							if b == 0 {
								if a >= 0 {
									return 99999
								}
								return -99999
							}
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
				})).Parser()),
			input:      " \\ 1 | + : 3; ++ * & 2 , - . 6 ~ ++ ++ ++ / ' 3",
			wantOutput: []int64{1, 3, 2, -6, 1},
			wantErrors: 9,
		}, {
			name: "parentheses and infix ops",
			parser: cmb.Count(7, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PostfixLevel([]cmb.PostfixOp[int64]{
					{
						Op:       "++",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i + 1
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
							if b == 0 {
								if a >= 0 {
									return 99999
								}
								return -99999
							}
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
				})).AddParentheses("(", ")", true).Parser()),
			input:      " \\( | 1 < + > 3 , ) . ++ * ; ( : 2 ! - ? 6 ' ) @ ++ ++ / # 2",
			wantOutput: []int64{1, 3, 0, 2, -6, 0, 1},
			wantErrors: 13,
		}, {
			name: "missing closing parenthesis",
			parser: cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PostfixLevel([]cmb.PostfixOp[int64]{
					{
						Op:       "++",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i + 1
						},
					},
				})).AddParentheses("(", ")", true).Parser()),
			input:      "( 1 ++",
			wantOutput: []int64{2},
			wantErrors: 1,
		}, {
			name: "space parser",
			parser: cmb.Count(4, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10))).
				AddPrefixLevel(cmb.PrefixOp[int64]{
					Op:       "-",
					SafeSpot: true,
					Fn: func(i int64) int64 {
						return -i
					},
				}).AddPostfixLevel(cmb.PostfixOp[int64]{
				Op:       "++",
				SafeSpot: true,
				Fn: func(i int64) int64 {
					return i + 1
				},
			}).AddInfixLevel(cmb.InfixOp[int64]{
				Op:       "*",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a * b
				},
			}).AddParentheses("[", "]", true).WithSpace(cmb.Whitespace1()).Parser()),
			input:      "-a[2* 3]++",
			wantOutput: []int64{-2, 0, 0, 1},
			wantErrors: 6,
		}, {
			name: "parse space after value in parentheses",
			parser: cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10))).
				AddPrefixLevel(cmb.PrefixOp[int64]{
					Op:       "-",
					SafeSpot: true,
					Fn: func(i int64) int64 {
						return -i
					},
				}).AddPostfixLevel(cmb.PostfixOp[int64]{
				Op:       "++",
				SafeSpot: true,
				Fn: func(i int64) int64 {
					return i + 1
				},
			}).AddInfixLevel(cmb.InfixOp[int64]{
				Op:       "*",
				SafeSpot: true,
				Fn: func(a, b int64) int64 {
					return a * b
				},
			}).AddParentheses("[", "]", true).WithSpace(cmb.Whitespace1()).Parser()),
			input:      " - [ 3] ",
			wantOutput: []int64{-3},
			wantErrors: 1,
		}, {
			name: "all mixed up",
			parser: cmb.Count(12, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
				cmb.PrefixLevel([]cmb.PrefixOp[int64]{
					{
						Op:       "-",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return -i
						},
					},
				}), cmb.PostfixLevel([]cmb.PostfixOp[int64]{
					{
						Op:       "--",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i - 1
						},
					}, {
						Op:       "++",
						SafeSpot: true,
						Fn: func(i int64) int64 {
							return i + 1
						},
					},
				}), cmb.PrefixLevel([]cmb.PrefixOp[int64]{
					{
						Op:       "!",
						SafeSpot: true,
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
							if b == 0 {
								return 0
							}
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
							if b == 0 {
								if a >= 0 {
									return 99999
								}
								return -99999
							}
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
				})).AddParentheses("(", ")", true).AddParentheses("[", "]", true).Parser()),
			input:      " \\ - | ( ? ! ~ 2 ` ++ ' + ; 3 : -- . ) , * @ [ # 2 $ ++ ++ ^ & 2 { - } 12 < ++ % > 6 a ] b ++ ++ ++ ++ / c 4 d +( 3 )",
			wantOutput: []int64{2, 1, 3, -1, 0, 2, 4, -12, 1, 0, 1, 3},
			wantErrors: 22,
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotOutput, err := comb.RunOnState(comb.NewFromString(tc.input, 50), comb.NewPreparedParser(tc.parser))
			t.Logf("got error(s) %v", err)
			if slices.Compare(gotOutput, tc.wantOutput) != 0 {
				t.Errorf("got output %#v, want output %#v", gotOutput, tc.wantOutput)
			}
			if got, want := len(comb.UnwrapErrors(err)), tc.wantErrors; got != want {
				t.Errorf("err=%v, want errors=%d", err, want)
			}
		})
	}
}

func TestExpression_FloatErrorCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		parser     comb.Parser[[]float64]
		input      string
		wantOutput []float64
		wantErrors int
	}{
		{
			name:       "additional character before value",
			parser:     cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Float64(true, 10, false))).Parser()),
			input:      "] -123",
			wantOutput: []float64{-123},
			wantErrors: 1,
		}, {
			name:       "integer as strict float value",
			parser:     cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Float64(false, 10, true))).Parser()),
			input:      "123 125.",
			wantOutput: []float64{125.0},
			wantErrors: 1,
		}, {
			name: "prefix op",
			parser: cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Float64(false, 10, false)),
				cmb.PrefixLevel([]cmb.PrefixOp[float64]{
					{
						Op:       "-",
						SafeSpot: true,
						Fn: func(i float64) float64 {
							return -i
						},
					},
				})).Parser()),
			input:      "! - | .123",
			wantOutput: []float64{-.123},
			wantErrors: 2,
		}, {
			name: "infix op",
			parser: cmb.Count(2, cmb.Expression(comb.SafeSpot(cmb.Float64(false, 10, true)),
				cmb.InfixLevel([]cmb.InfixOp[float64]{
					{
						Op:       "+",
						SafeSpot: true,
						Fn: func(a, b float64) float64 {
							return a + b
						},
					},
				})).Parser()),
			input:      "(12e3)+ 4e- 456.",
			wantOutput: []float64{12e3, 456.},
			wantErrors: 3,
		}, {
			name: "all mixed up",
			parser: cmb.Count(12, cmb.Expression(comb.SafeSpot(cmb.Float64(false, 10, false)),
				cmb.PrefixLevel([]cmb.PrefixOp[float64]{
					{
						Op:       "-",
						SafeSpot: true,
						Fn: func(i float64) float64 {
							return -i
						},
					},
				}), cmb.PostfixLevel([]cmb.PostfixOp[float64]{
					{
						Op:       "--",
						SafeSpot: true,
						Fn: func(i float64) float64 {
							return i - 1
						},
					}, {
						Op:       "++",
						SafeSpot: true,
						Fn: func(i float64) float64 {
							return i + 1
						},
					},
				}), cmb.PrefixLevel([]cmb.PrefixOp[float64]{
					{
						Op:       "!",
						SafeSpot: true,
						Fn: func(v float64) float64 {
							r := float64(1)
							for i := float64(1); i <= v; i++ {
								r *= i
							}
							return r
						},
					},
				}), cmb.InfixLevel([]cmb.InfixOp[float64]{
					{
						Op:       "^",
						SafeSpot: true,
						Fn: func(a, b float64) float64 {
							r := float64(1)
							for i := float64(0); i < b; i++ {
								r *= a
							}
							return r
						},
					}, {
						Op:       "%",
						SafeSpot: true,
						Fn: func(a, b float64) float64 {
							if b == 0 {
								return 0
							}
							return math.Remainder(a, b)
						},
					},
				}), cmb.InfixLevel([]cmb.InfixOp[float64]{
					{
						Op:       "*",
						SafeSpot: true,
						Fn: func(a, b float64) float64 {
							return a * b
						},
					}, {
						Op:       "/",
						SafeSpot: true,
						Fn: func(a, b float64) float64 {
							if b == 0 {
								if a >= 0 {
									return 99999
								}
								return -99999
							}
							return a / b
						},
					},
				}), cmb.InfixLevel([]cmb.InfixOp[float64]{
					{
						Op:       "-",
						SafeSpot: true,
						Fn: func(a, b float64) float64 {
							return a - b
						},
					}, {
						Op:       "+",
						SafeSpot: true,
						Fn: func(a, b float64) float64 {
							return a + b
						},
					},
				})).AddParentheses("(", ")", true).AddParentheses("[", "]", true).Parser()),
			input:      " \\ - | ( ? ! ~ 2 ` ++ ' + ; 3 : -- . ) , * @ [ # 2 $ ++ ++ ^ & 2 { - } 12 < ++ % > 6 a ] b ++ ++ ++ ++ / c 4 d +( 3 )",
			wantOutput: []float64{2, 1, 3, -1, 0, 2, 4, -12, 1, 0, 1, 3},
			wantErrors: 22,
		},
	}

	for _, tc := range testCases {
		tc := tc // this is needed for t.Parallel() to work correctly (or the same test case will be executed N times)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotOutput, err := comb.RunOnState(comb.NewFromString(tc.input, 50), comb.NewPreparedParser(tc.parser))
			t.Logf("got error(s) %v", err)
			if slices.Compare(gotOutput, tc.wantOutput) != 0 {
				t.Errorf("got output %#v, want output %#v", gotOutput, tc.wantOutput)
			}
			if got, want := len(comb.UnwrapErrors(err)), tc.wantErrors; got != want {
				t.Errorf("err=%v, want errors=%d", err, want)
			}
		})
	}
}

func TestExpression_Panics(t *testing.T) {
	t.Parallel()

	t.Run("empty string prefix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op: "",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			})).Parser()
	})

	t.Run("nil prefix func", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op: "-",
					Fn: nil,
				},
			})).Parser()
	})

	t.Run("duplicate prefix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				}, {
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			})).Parser()
	})

	t.Run("empty string infix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op: "",
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			})).Parser()
	})

	t.Run("nil infix func", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op: "+",
					Fn: nil,
				},
			})).Parser()
	})

	t.Run("duplicate infix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op: "+",
					Fn: func(a, b int64) int64 {
						return a + b
					},
				}, {
					Op: "+",
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			})).Parser()
	})

	t.Run("empty string postfix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op: "",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			})).Parser()
	})

	t.Run("nil postfix func", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op: "-",
					Fn: nil,
				},
			})).Parser()
	})

	t.Run("duplicate postfix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				}, {
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			})).Parser()
	})

	t.Run("duplicate multi-level prefix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			}), cmb.PrefixLevel([]cmb.PrefixOp[int64]{
				{
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			})).Parser()
	})

	t.Run("duplicate multi-level infix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op: "+",
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			}), cmb.InfixLevel([]cmb.InfixOp[int64]{
				{
					Op: "+",
					Fn: func(a, b int64) int64 {
						return a + b
					},
				},
			})).Parser()
	})

	t.Run("duplicate multi-level postfix op", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			}), cmb.PostfixLevel([]cmb.PostfixOp[int64]{
				{
					Op: "-",
					Fn: func(i int64) int64 {
						return -i
					},
				},
			})).Parser()
	})

	t.Run("duplicate parentheses", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10)).
			AddParentheses("(", ")", false).
			AddParentheses("(", ")", false).Parser()
	})

	t.Run("empty postfix op level", func(t *testing.T) {
		t.Parallel()
		defer recoverFunc(t)()
		cmb.Expression(cmb.Int64(false, 10),
			cmb.PostfixLevel([]cmb.PostfixOp[int64]{})).Parser()
	})

}
func recoverFunc(t *testing.T) func() {
	return func() {
		// if the test panics, recover() returns a non nil value
		r := recover()
		t.Logf("panic: %v", r)
		if r == nil {
			t.Errorf("test should panic")
		}
	}
}
