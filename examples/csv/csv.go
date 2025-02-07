// Package csv implements a parser for CSV files.
//
// It is a simple, incomplete, example of how to use the gomme
// parser combinator library to build a parser targeting the
// format described in [RFC4180].
//
// [RFC4180]: https://tools.ietf.org/html/rfc4180
package csv

import (
	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
	. "github.com/flowdev/comb/cute"
)

func ParseCSV(input string) ([][]string, error) {
	parser := cmb.Separated1(
		cmb.Separated1(
			FirstSuccessful(
				cmb.Alphanumeric1(),
				cmb.Delimited(cmb.Char('"'), cmb.Alphanumeric1(), cmb.Char('"')),
			),
			cmb.Char(','),
			false,
		),
		cmb.CRLF(),
		false,
	)

	return comb.RunOnString(input, parser)
}
