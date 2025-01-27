package cute

import (
	"github.com/oleiade/gomme"
	"github.com/oleiade/gomme/pcb"
)

// C is a shortened version of `pcb.Char`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/oleiade/gomme/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func C(char rune) gomme.Parser[rune] {
	return pcb.Char(char)
}

// S is a shortened version of `pcb.Char`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/oleiade/gomme/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func S(token string) gomme.Parser[string] {
	return pcb.String(token)
}

// OneOfRunes is a shortened version of `pcb.OneOfRunes`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/oleiade/gomme/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func OneOfRunes(collection ...rune) gomme.Parser[rune] {
	return pcb.OneOfRunes(collection...)

}

// OneOf is a shortened version of `pcb.OneOf`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/oleiade/gomme/cute"
//
// This parser is a good candidate for SaveSpot and has an optimized recoverer.
func OneOf(collection ...string) gomme.Parser[string] {
	return pcb.OneOf(collection...)

}

// SaveSpot is the shortened version of `pcb.SafeSpot`.
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
//	import . "github.com/oleiade/gomme/cute"
func SaveSpot[Output any](parse gomme.Parser[Output]) gomme.Parser[Output] {
	return gomme.SafeSpot[Output](parse)
}

// FirstSuccessful is a shortened version of `gomme.FirstSuccessful`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/oleiade/gomme/cute"
func FirstSuccessful[Output any](parsers ...gomme.Parser[Output]) gomme.Parser[Output] {
	return pcb.FirstSuccessful[Output](parsers...)
}

// ZeroOf returns the zero value of some type.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/oleiade/gomme/cute"
func ZeroOf[T any]() T {
	return gomme.ZeroOf[T]()
}
