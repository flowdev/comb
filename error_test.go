package comb

import (
	"testing"
)

func TestPatchMessage(t *testing.T) {
	t.Parallel()

	state := NewFromString("source", 0)
	tailMsg := " [1:1] ▶source"

	tests := []struct {
		name    string
		err     *ParserError
		subMsg  string
		wantMsg string
	}{
		{
			name:    "syntax error",
			err:     state.NewSyntaxError("source"),
			subMsg:  "patch ",
			wantMsg: "expected patch source",
		}, {
			name:    "semantic error",
			err:     state.NewSemanticError("no source"),
			subMsg:  "patch ",
			wantMsg: "patch no source",
		}, {
			name:    "double patch",
			err:     state.NewSyntaxError("no patch source"),
			subMsg:  "patch ",
			wantMsg: "expected no patch source",
		},
	}
	for _, tt := range tests {
		tt := tt // needed for truly different test cases!
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.err.PatchMessage(tt.subMsg)

			if got, want := tt.err.Error(), tt.wantMsg+tailMsg; got != want {
				t.Errorf("got message %q, want: %q", got, want)
			}
		})
	}
}

func TestClaimError(t *testing.T) {
	t.Parallel()

	state := NewFromString("source", 0)

	tests := []struct {
		name         string
		gotParserID  int32
		wantParserID int32
	}{
		{
			name:         "parser ID 0",
			gotParserID:  0,
			wantParserID: -1,
		}, {
			name:         "parser ID 1",
			gotParserID:  1,
			wantParserID: -1,
		}, {
			name:         "parser ID -1",
			gotParserID:  -1,
			wantParserID: -1,
		},
	}
	for _, tt := range tests {
		tt := tt // needed for truly different test cases!
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := state.NewSyntaxError("no source")
			err.parserID = tt.gotParserID
			gotErr := ClaimError(err)
			if got, want := gotErr.parserID, int32(-1); got != want {
				t.Errorf("got parser ID %d, want: %d", got, want)
			}
		})
	}
}
func TestFirstNRunes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{
			name: "first ASCII char",
			s:    "abc",
			n:    1,
			want: "a",
		}, {
			name: "multiple ASCII char",
			s:    "abc",
			n:    2,
			want: "ab",
		}, {
			name: "all ASCII char",
			s:    "abc",
			n:    3,
			want: "abc",
		}, {
			name: "more than all chars",
			s:    "abc",
			n:    4,
			want: "abc",
		}, {
			name: "first non-ASCII char",
			s:    "öabc",
			n:    1,
			want: "ö",
		}, {
			name: "multiple non-ASCII char",
			s:    "äöüabc",
			n:    3,
			want: "äöü",
		}, {
			name: "all non-ASCII char",
			s:    "äöü",
			n:    3,
			want: "äöü",
		},
	}
	for _, tt := range tests {
		tt := tt // needed for truly different test cases!
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := firstNRunes(tt.s, tt.n), tt.want; got != want {
				t.Errorf("got runes %q, want: %q", got, want)
			}
		})
	}
}

func TestLastNRunes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{
			name: "last ASCII char",
			s:    "abc",
			n:    1,
			want: "c",
		}, {
			name: "multiple ASCII char",
			s:    "abc",
			n:    2,
			want: "bc",
		}, {
			name: "all ASCII char",
			s:    "abc",
			n:    3,
			want: "abc",
		}, {
			name: "more than all chars",
			s:    "abc",
			n:    4,
			want: "abc",
		}, {
			name: "last non-ASCII char",
			s:    "abcö",
			n:    1,
			want: "ö",
		}, {
			name: "multiple non-ASCII char",
			s:    "abcäöü",
			n:    3,
			want: "äöü",
		}, {
			name: "all non-ASCII char",
			s:    "äöü",
			n:    3,
			want: "äöü",
		},
	}
	for _, tt := range tests {
		tt := tt // needed for truly different test cases!
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := lastNRunes(tt.s, tt.n), tt.want; got != want {
				t.Errorf("got runes %q, want: %q", got, want)
			}
		})
	}
}
