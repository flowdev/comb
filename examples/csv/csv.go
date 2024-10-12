// Package csv implements a parser for CSV files.
//
// It is a simple, incomplete, example of how to use the gomme
// parser combinator library to build a parser targeting the
// format described in [RFC4180].
//
// [RFC4180]: https://tools.ietf.org/html/rfc4180
package csv

import (
	"github.com/oleiade/gomme"
	"github.com/oleiade/gomme/pcb"
)

func ParseCSV(input string) ([][]string, error) {
	parser := pcb.Separated1(
		pcb.Separated1(
			gomme.FirstSuccessful(
				pcb.Alphanumeric1(),
				pcb.Delimited(pcb.Char('"'), pcb.Alphanumeric1(), pcb.Char('"')),
			),
			pcb.Char(','),
			false,
		),
		pcb.CRLF(),
		false,
	)

	return gomme.RunOnString(0, nil, input, parser)
}
