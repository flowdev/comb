package cmb_test

import (
	"slices"
	"testing"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
)

func TestExpression_HappyPath(t *testing.T) {
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

func TestExpression_ErrorCases(t *testing.T) {
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
			parser:     cmb.Count(1, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10))).Parser()),
			input:      "] 123",
			wantOutput: []int64{123},
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
			parser: cmb.Count(11, cmb.Expression(comb.SafeSpot(cmb.Int64(false, 10)),
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
			input:      " \\ - | ( ? ! ~ 2 ` ++ ' + ; 3 : -- . ) , * @ [ # 2 $ ++ ++ ^ & 2 { - } 12 < ++ % > 6 a ] b ++ ++ ++ ++ / c 4",
			wantOutput: []int64{2, 1, 3, -1, 0, 2, 4, -12, 1, 0, 1},
			wantErrors: 21,
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
