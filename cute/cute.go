package cute

import (
	"github.com/flowdev/comb"
	"github.com/flowdev/comb/cmb"
)

// C is a shortened version of `cmb.Char`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/flowdev/comb/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func C(char rune) comb.Parser[rune] {
	return cmb.Char(char)
}

// S is a shortened version of `cmb.Char`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/flowdev/comb/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func S(token string) comb.Parser[string] {
	return cmb.String(token)
}

// OneOfRunes is a shortened version of `cmb.OneOfRunes`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/flowdev/comb/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func OneOfRunes(collection ...rune) comb.Parser[rune] {
	return cmb.OneOfRunes(collection...)

}

// OneOf is a shortened version of `cmb.OneOf`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/flowdev/comb/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func OneOf(collection ...string) comb.Parser[string] {
	return cmb.OneOf(collection...)

}

// SaveSpot is the shortened version of `comb.SafeSpot`.
// This should encourage its use because SaveSpot is the backbone of the
// error handling mechanism.
//
// So please use it!
//
// Parsers that only accept keywords or special tokens are excellent candidates.
// Especially `C` and `S` but also `OneOf` and `OneOfRunes` from this package.
//
// SaveSpot is meant to be used without the package name with an import like:
//
//	import . "github.com/flowdev/comb/cute"
func SaveSpot[Output any](parse comb.Parser[Output]) comb.Parser[Output] {
	return comb.SafeSpot[Output](parse)
}

// ZeroOf returns the zero value of some type.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/flowdev/comb/cute"
func ZeroOf[T any]() T {
	return comb.ZeroOf[T]()
}
