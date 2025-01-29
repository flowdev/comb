package pcb

import (
	"github.com/oleiade/gomme"
)

// MapN is a helper for easily implementing Map like parsers.
// It is not meant for writing grammars, but only for implementing parsers.
// Only the `fn`n function has to be provided.
// All other `fn`X functions are expected to be `nil`.
// Only parsers up to `p`n have to be provided.
// All higher numbered parsers are expected to be nil.
func MapN[PO1, PO2, PO3, PO4, PO5 any, MO any](
	expected string,
	p1 gomme.Parser[PO1], p2 gomme.Parser[PO2], p3 gomme.Parser[PO3], p4 gomme.Parser[PO4], p5 gomme.Parser[PO5],
	n int,
	fn1 func(PO1) (MO, error), fn2 func(PO1, PO2) (MO, error), fn3 func(PO1, PO2, PO3) (MO, error),
	fn4 func(PO1, PO2, PO3, PO4) (MO, error), fn5 func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) gomme.Parser[MO] {
	if p1 == nil {
		panic("MapN: p1 is nil")
	}
	if n >= 2 {
		if p2 == nil {
			panic("MapN: p2 is nil (n >= 2)")
		}
		if n >= 3 {
			if p3 == nil {
				panic("MapN: p3 is nil (n >= 3)")
			}
			if n >= 4 {
				if p4 == nil {
					panic("MapN: p4 is nil (n >= 4)")
				}
				if n >= 5 {
					if p5 == nil {
						panic("MapN: p5 is nil (n >= 5)")
					}
				}
			}
		}
	}

	switch n {
	case 1:
		if fn1 == nil {
			panic("MapN: fn1 is nil")
		}
	case 2:
		if fn2 == nil {
			panic("MapN: fn2 is nil")
		}
	case 3:
		if fn3 == nil {
			panic("MapN: fn3 is nil")
		}
	case 4:
		if fn4 == nil {
			panic("MapN: fn4 is nil")
		}
	default:
		if fn5 == nil {
			panic("MapN: fn5 is nil")
		}
	}

	md := &mapData[PO1, PO2, PO3, PO4, PO5, MO]{
		expected: expected,
		p1:       p1, p2: p2, p3: p3, p4: p4, p5: p5,
		n:   n,
		fn1: fn1, fn2: fn2, fn3: fn3, fn4: fn4, fn5: fn5,
	}

	return gomme.NewBranchParser[MO](expected, md.children, md.parseAfterChild)
}

type mapData[PO1, PO2, PO3, PO4, PO5 any, MO any] struct {
	expected string
	p1       gomme.Parser[PO1]
	p2       gomme.Parser[PO2]
	p3       gomme.Parser[PO3]
	p4       gomme.Parser[PO4]
	p5       gomme.Parser[PO5]
	n        int
	fn1      func(PO1) (MO, error)
	fn2      func(PO1, PO2) (MO, error)
	fn3      func(PO1, PO2, PO3) (MO, error)
	fn4      func(PO1, PO2, PO3, PO4) (MO, error)
	fn5      func(PO1, PO2, PO3, PO4, PO5) (MO, error)
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) children() []gomme.AnyParser {
	children := make([]gomme.AnyParser, md.n)
	children[0] = md.p1
	if md.n >= 2 {
		children[1] = md.p2
		if md.n >= 3 {
			children[2] = md.p3
			if md.n >= 4 {
				children[3] = md.p4
				if md.n >= 5 {
					children[4] = md.p5
				}
			}
		}
	}
	return children
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) parseAfterChild(childID int32, childResult gomme.ParseResult,
) gomme.ParseResult {
	var zero MO
	var zero1 PO1
	var zero2 PO2
	var zero3 PO3
	var zero4 PO4
	var zero5 PO5

	gomme.Debugf("MapN.parseAfterChild - childID=%d, pos=%d", childID, childResult.EndState.CurrentPos())

	if childResult.Error != nil {
		return childResult // we can't avoid any errors by going another path
	}

	state := childResult.EndState
	id := childID
	idErrResult := gomme.ParseResult{
		StartState: state,
		EndState:   state,
		Output:     zero,
		Error:      state.NewSemanticError("unable to parse after child with unknown ID %d", id),
	}
	if id >= 0 && id != md.p1.ID() {
		if md.n <= 1 {
			return idErrResult
		}
		if id != md.p2.ID() {
			if md.n <= 2 {
				return idErrResult
			}
			if id != md.p3.ID() {
				if md.n <= 3 {
					return idErrResult
				}
				if id != md.p4.ID() {
					if md.n <= 4 {
						return idErrResult
					}
					if id != md.p5.ID() {
						return idErrResult
					}
				}
			}
		}
	}

	state1, out1, err1 := state, zero1, (*gomme.ParserError)(nil)
	if id < 0 {
		state1, out1, err1 = md.p1.Parse(state)
		if err1 != nil {
			return gomme.ParseResult{StartState: state, EndState: state1, Output: out1, Error: err1}
		}
	} else if id == md.p1.ID() {
		state1 = childResult.EndState
		out1, _ = childResult.Output.(PO1)
		err1 = childResult.Error
		id = -1
	}

	if md.n > 1 {
		state2, out2, err2 := state, zero2, (*gomme.ParserError)(nil)
		if id == md.p2.ID() {
			state2 = childResult.EndState
			out2, _ = childResult.Output.(PO2)
			err2 = childResult.Error
			id = -1
		} else if id < 0 {
			state2, out2, err2 = md.p2.Parse(state1)
			if err2 != nil {
				return gomme.ParseResult{StartState: state1, EndState: state2, Output: out2, Error: err2}
			}
		}

		if md.n > 2 {
			state3, out3, err3 := state, zero3, (*gomme.ParserError)(nil)
			if id == md.p3.ID() {
				state3 = childResult.EndState
				out3, _ = childResult.Output.(PO3)
				err3 = childResult.Error
				id = -1
			} else if id < 0 {
				state3, out3, err3 = md.p3.Parse(state2)
				if err3 != nil {
					return gomme.ParseResult{StartState: state2, EndState: state3, Output: out3, Error: err3}
				}
			}

			if md.n > 3 {
				state4, out4, err4 := state, zero4, (*gomme.ParserError)(nil)
				if id == md.p4.ID() {
					state4 = childResult.EndState
					out4, _ = childResult.Output.(PO4)
					err4 = childResult.Error
				} else if id < 0 {
					state4, out4, err4 = md.p4.Parse(state3)
					if err4 != nil {
						return gomme.ParseResult{StartState: state3, EndState: state4, Output: out4, Error: err4}
					}
				}

				if md.n > 4 {
					state5, out5, err5 := state, zero5, (*gomme.ParserError)(nil)
					if id == md.p5.ID() {
						state5 = childResult.EndState
						out5, _ = childResult.Output.(PO5)
						err5 = childResult.Error
					} else if id < 0 {
						state5, out5, err5 = md.p5.Parse(state4)
						if err5 != nil {
							return gomme.ParseResult{StartState: state4, EndState: state5, Output: out5, Error: err5}
						}
					}

					out, err := md.fn5(out1, out2, out3, out4, out5)
					var pErr *gomme.ParserError
					if err != nil {
						pErr = state5.NewSemanticError(err.Error())
					}
					return gomme.ParseResult{StartState: state, EndState: state5, Output: out, Error: pErr}
				}

				out, err := md.fn4(out1, out2, out3, out4)
				var pErr *gomme.ParserError
				if err != nil {
					pErr = state4.NewSemanticError(err.Error())
				}
				return gomme.ParseResult{StartState: state, EndState: state4, Output: out, Error: pErr}
			}

			out, err := md.fn3(out1, out2, out3)
			var pErr *gomme.ParserError
			if err != nil {
				pErr = state3.NewSemanticError(err.Error())
			}
			return gomme.ParseResult{StartState: state, EndState: state3, Output: out, Error: pErr}
		}

		out, err := md.fn2(out1, out2)
		var pErr *gomme.ParserError
		if err != nil {
			pErr = state2.NewSemanticError(err.Error())
		}
		return gomme.ParseResult{StartState: state, EndState: state2, Output: out, Error: pErr}
	}

	out, err := md.fn1(out1)
	var pErr *gomme.ParserError
	if err != nil {
		pErr = state1.NewSemanticError(err.Error())
	}
	return gomme.ParseResult{StartState: state, EndState: state1, Output: out, Error: pErr}
}
