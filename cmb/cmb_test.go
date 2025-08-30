package cmb_test

import (
	"testing"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
)

func TestEOF(t *testing.T) {
	state1 := comb.NewFromString("123", 0).MoveBy(2)
	state2 := state1.MoveBy(1)

	tests := []struct {
		name          string
		state         comb.State
		wantErr       bool
		wantRemaining string
	}{
		{
			name:          "1 char remaining",
			state:         state1,
			wantErr:       true,
			wantRemaining: "3",
		}, {
			name:          "0 chars remaining",
			state:         state2,
			wantErr:       false,
			wantRemaining: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endState, _, err := cmb.EOF().Parse(tt.state)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("got error %v, want error: %t", err, tt.wantErr)
			}

			gotRemaining := endState.CurrentString()
			if gotRemaining != tt.wantRemaining {
				t.Errorf("got remaining %q, want remaining: %q", gotRemaining, tt.wantRemaining)
			}
		})
	}
}
