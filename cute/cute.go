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
// So this cute function name will make your grammar look nicer.
func C(char rune) gomme.Parser[rune] {
	return pcb.Char(char)
}

// S is a shortened version of `pcb.Char`.
// It is meant to be used without the package name with an import like:
//
//	import . "github.com/oleiade/gomme/cute"
//
// So this cute function name will make your grammar look nicer.
func S(token string) gomme.Parser[string] {
	return pcb.String(token)
}
