package gomme_test

/*
func TestSaveSpot(t *testing.T) {
	t.Parallel()

	type args struct {
		p gomme.Parser[string]
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
			name:  "head matching parser should succeed",
			input: "123",
			args: args{
				p: pcb.FirstSuccessful(pcb.Digit1(), gomme.SafeSpot(pcb.Alpha1())),
			},
			wantErr:       false,
			wantOutput:    "123",
			wantRemaining: "",
		},
		{
			name:  "tail matching parser should succeed",
			input: "abc",
			args: args{
				p: pcb.FirstSuccessful(gomme.SafeSpot(pcb.Digit1()), pcb.Alpha1()),
			},
			wantErr:       false,
			wantOutput:    "abc",
			wantRemaining: "",
		},
		{
			name:  "FirstSuccessful: tail matching parser after failing SafeSpot head parser should fail",
			input: "abc",
			args: args{
				p: pcb.FirstSuccessful(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1()), pcb.Alpha1()),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
		{
			name:  "Optional: tail matching parser after failing SafeSpot head parser should fail",
			input: "abc",
			args: args{
				p: pcb.Optional(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1())),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "bc",
		},
		{
			name:  "Many0: tail matching parser after failing SafeSpot head parser should fail",
			input: "abc",
			args: args{
				p: pcb.Map(pcb.Many0(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1())), func(tokens []string) (string, error) {
					return strings.Join(tokens, ""), nil
				}),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "abc",
		},
		{
			name:  "Seperated1: matching main parser after failing SafeSpot head parser should fail",
			input: "a,1",
			args: args{
				p: pcb.Map(pcb.Separated0(pcb.Prefixed(gomme.SafeSpot(pcb.String("a")), pcb.Digit1()), pcb.Char(','), false),
					func(tokens []string) (string, error) {
						return strings.Join(tokens, ""), nil
					},
				),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "a,1",
		},
		{
			name:  "no matching parser should fail",
			input: "$%^*",
			args: args{
				p: pcb.FirstSuccessful(gomme.SafeSpot(pcb.Digit1()), gomme.SafeSpot(pcb.Alpha1())),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "$%^*",
		},
		{
			name:  "empty input should fail",
			input: "",
			args: args{
				p: pcb.FirstSuccessful(gomme.SafeSpot(pcb.Digit1()), gomme.SafeSpot(pcb.Alpha1())),
			},
			wantErr:       true,
			wantOutput:    "",
			wantRemaining: "",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := gomme.NewFromString(-1, nil, -1, tc.input)
			newState, gotResult := tc.args.p.It(state)
			if newState.Failed() != tc.wantErr {
				t.Errorf("got error %v, want error %v", newState.Errors(), tc.wantErr)
			}

			if gotResult != tc.wantOutput {
				t.Errorf("got output %q, want %q", gotResult, tc.wantOutput)
			}

			if newState.CurrentString() != tc.wantRemaining {
				t.Errorf("got remaining %q, want remaining %q", newState.CurrentString(), tc.wantRemaining)
			}
		})
	}
}

func BenchmarkSaveSpot(b *testing.B) {
	p := gomme.SafeSpot(pcb.Char('1'))
	input := gomme.NewFromString(1, nil, -1, "123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.It(input)
	}
}
*/
