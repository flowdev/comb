package pcb

import (
	"github.com/oleiade/gomme"
	"strings"
)

// MapN is a helper for easily implementing Map like parsers.
// It is not meant for writing grammars, but only for implementing parsers.
// Only the `fn`n function has to be provided.
// All other `fn`X functions are expected to be `nil`.
// Only parsers up to `p`n have to be provided.
// All higher numbered parsers are expected to be nil.
func MapN[PO1, PO2, PO3, PO4, PO5 any, MO any](
	p1 gomme.Parser[PO1], p2 gomme.Parser[PO2], p3 gomme.Parser[PO3], p4 gomme.Parser[PO4], p5 gomme.Parser[PO5],
	n int,
	fn1 func(PO1) (MO, error), fn2 func(PO1, PO2) (MO, error), fn3 func(PO1, PO2, PO3) (MO, error),
	fn4 func(PO1, PO2, PO3, PO4) (MO, error), fn5 func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) gomme.Parser[MO] {
	var zeroMO MO
	id := gomme.NewBranchParserID()

	expected := strings.Builder{}
	expected.WriteString(p1.Expected())
	if n > 1 {
		expected.WriteString(" + ")
		expected.WriteString(p2.Expected())
		if n > 2 {
			expected.WriteString(" + ")
			expected.WriteString(p3.Expected())
			if n > 3 {
				expected.WriteString(" + ")
				expected.WriteString(p4.Expected())
				if n > 4 {
					expected.WriteString(" + ")
					expected.WriteString(p5.Expected())
				}
			}
		}
	}

	containsNoWayBack := p1.ContainsNoWayBack()
	if n > 1 {
		containsNoWayBack = max(containsNoWayBack, p2.ContainsNoWayBack())
		if n > 2 {
			containsNoWayBack = max(containsNoWayBack, p3.ContainsNoWayBack())
			if n > 3 {
				containsNoWayBack = max(containsNoWayBack, p4.ContainsNoWayBack())
				if n > 4 {
					containsNoWayBack = max(containsNoWayBack, p5.ContainsNoWayBack())
				}
			}
		}
	}

	// Construct myNoWayBackRecoverer from the sub-parsers
	subRecoverers := make([]gomme.Recoverer, 0, 5)
	if p1.ContainsNoWayBack() > gomme.TernaryNo {
		subRecoverers = append(subRecoverers, p1.NoWayBackRecoverer)
	}
	if n > 1 {
		if p2.ContainsNoWayBack() > gomme.TernaryNo {
			subRecoverers = append(subRecoverers, p2.NoWayBackRecoverer)
		}
		if n > 2 {
			if p3.ContainsNoWayBack() > gomme.TernaryNo {
				subRecoverers = append(subRecoverers, p3.NoWayBackRecoverer)
			}
			if n > 3 {
				if p4.ContainsNoWayBack() > gomme.TernaryNo {
					subRecoverers = append(subRecoverers, p4.NoWayBackRecoverer)
				}
				if n > 4 {
					if p5.ContainsNoWayBack() > gomme.TernaryNo {
						subRecoverers = append(subRecoverers, p5.NoWayBackRecoverer)
					}
				}
			}
		}
	}
	myNoWayBackRecoverer := gomme.NewCombiningRecoverer(subRecoverers...)

	mapParse := func(state gomme.State) (gomme.State, MO) {
		newState1, output1 := p1.It(state)
		if newState1.Failed() {
			return gomme.IWitnessed(state, id, 0, newState1), zeroMO
		}

		if n > 1 {
			newState2, output2 := p2.It(newState1)
			if newState2.Failed() {
				return gomme.IWitnessed(state, id, 1, newState2), zeroMO
			}

			if n > 2 {
				newState3, output3 := p3.It(newState2)
				if newState3.Failed() {
					return gomme.IWitnessed(state, id, 2, newState3), zeroMO
				}

				if n > 3 {
					newState4, output4 := p4.It(newState3)
					if newState4.Failed() {
						return gomme.IWitnessed(state, id, 3, newState4), zeroMO
					}

					if n > 4 {
						newState5, output5 := p5.It(newState4)
						if newState5.Failed() {
							return gomme.IWitnessed(state, id, 4, newState5), zeroMO
						}

						mapped, err := fn5(output1, output2, output3, output4, output5)
						if err != nil {
							return state.NewSemanticError(err.Error()), zeroMO
						}
						return newState5, mapped
					}
					mapped, err := fn4(output1, output2, output3, output4)
					if err != nil {
						return state.NewSemanticError(err.Error()), zeroMO
					}
					return newState4, mapped
				}
				mapped, err := fn3(output1, output2, output3)
				if err != nil {
					return state.NewSemanticError(err.Error()), zeroMO
				}
				return newState3, mapped
			}
			mapped, err := fn2(output1, output2)
			if err != nil {
				return state.NewSemanticError(err.Error()), zeroMO
			}
			return newState2, mapped
		}
		mapped, err := fn1(output1)
		if err != nil {
			return state.NewSemanticError(err.Error()), zeroMO
		}
		return newState1, mapped
	}

	return gomme.NewParser[MO](
		expected.String(),
		mapParse,
		true,
		BasicRecovererFunc(mapParse),
		containsNoWayBack,
		myNoWayBackRecoverer.Recover,
	)
}
