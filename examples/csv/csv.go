// Package csv implements a parser for CSV files.
//
// It is a simple, incomplete, example of how to use the gomme
// parser combinator library to build a parser targeting the
// format described in [RFC4180].
//
// [RFC4180]: https://tools.ietf.org/html/rfc4180
package csv

import "github.com/oleiade/gomme"

func ParseCSV(input string) ([][]string, error) {
	parser := gomme.SeparatedList1(
		gomme.SeparatedList1(
			gomme.Alternative(
				gomme.Alphanumeric1(),
				gomme.Delimited(gomme.Char('"'), gomme.Alphanumeric1(), gomme.Char('"')),
			),
			gomme.Char(','),
			false,
		),
		gomme.CRLF(),
		false,
	)

	newState, output := parser(gomme.NewInputFromString(input))
	if newState.Failed() {
		return nil, result.Err
	}

	return output, nil
}
