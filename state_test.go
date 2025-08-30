package comb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesTo(t *testing.T) {
	t.Parallel()

	textState1 := NewFromString("12345678", 0)
	textState2 := textState1.MoveBy(4)
	binaryState1 := NewFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 0)
	binaryState2 := binaryState1.MoveBy(4)

	testCases := []struct {
		name      string
		state1    State
		state2    State
		wantBytes []byte
	}{
		{
			name:      "text moved forward",
			state1:    textState1,
			state2:    textState2,
			wantBytes: []byte("1234"),
		}, {
			name:      "text moved backward",
			state1:    textState2,
			state2:    textState1,
			wantBytes: []byte{},
		}, {
			name:      "binary moved forward",
			state1:    binaryState1,
			state2:    binaryState2,
			wantBytes: []byte{1, 2, 3, 4},
		}, {
			name:      "binary moved backward",
			state1:    binaryState2,
			state2:    binaryState1,
			wantBytes: []byte{},
		},
	}
	for _, tt := range testCases {
		tt := tt // needed for not testing the same case N times
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantBytes, tt.state1.BytesTo(tt.state2))
		})
	}
}
