package json

import (
	"fmt"
	"testing"

	"github.com/flowdev/comb"
)

func TestParseRESPMessage(t *testing.T) {
	t.Parallel()

	testJSON := `{
  "abc": 123,
  "entries": [
    {
      "name": "John",
      "age": 30
    },
    {
      "name": "Jane",
      "age": 25
    }
  ]
}
`

	output, err := comb.RunOnString(testJSON, valuep)

	//if err != nil {
	if got, want := err, error(nil); got != want {
		t.Errorf("No error expexted but got err=%v", err)
	}

	if got, want := fmt.Sprintln(output),
		"map[abc:123 entries:[map[age:30 name:John] map[age:25 name:Jane]]]\n"; got != want {
		t.Errorf("got JSON output=%q, want=%q", got, want)
	}
}
