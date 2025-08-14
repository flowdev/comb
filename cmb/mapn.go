package cmb

import (
	"github.com/flowdev/comb"
)

// MapN is a helper for easily implementing Map like parsers.
// It is not meant for writing grammars, but only for implementing parsers.
// Only the `fn`n function has to be provided.
// All other `fn`X functions are expected to be `nil`.
// Only parsers up to `p`n have to be provided.
// All higher-numbered parsers are expected to be nil.
func MapN[PO1, PO2, PO3, PO4, PO5 any, MO any](
	expected string,
	p1 comb.Parser[PO1], p2 comb.Parser[PO2], p3 comb.Parser[PO3], p4 comb.Parser[PO4], p5 comb.Parser[PO5],
	n int,
	fn1 func(PO1) (MO, error), fn2 func(PO1, PO2) (MO, error), fn3 func(PO1, PO2, PO3) (MO, error),
	fn4 func(PO1, PO2, PO3, PO4) (MO, error), fn5 func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) comb.Parser[MO] {
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

	p := comb.NewBranchParser[MO](expected, md.children, md.parseAfterChild)
	md.id = p.ID
	return p
}

type mapData[PO1, PO2, PO3, PO4, PO5 any, MO any] struct {
	id       func() int32
	expected string
	p1       comb.Parser[PO1]
	p2       comb.Parser[PO2]
	p3       comb.Parser[PO3]
	p4       comb.Parser[PO4]
	p5       comb.Parser[PO5]
	n        int
	fn1      func(PO1) (MO, error)
	fn2      func(PO1, PO2) (MO, error)
	fn3      func(PO1, PO2, PO3) (MO, error)
	fn4      func(PO1, PO2, PO3, PO4) (MO, error)
	fn5      func(PO1, PO2, PO3, PO4, PO5) (MO, error)
}

// partialMapResult is internal to the parsing method and methods and functions called by it.
type partialMapResult[PO1, PO2, PO3, PO4 any] struct {
	out1 PO1
	out2 PO2
	out3 PO3
	out4 PO4
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) children() []comb.AnyParser {
	children := make([]comb.AnyParser, md.n)
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

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) parseAfterChild(
	childID int32,
	childStartState, childState comb.State,
	childOut interface{},
	childErr *comb.ParserError,
	data interface{},
) (comb.State, MO, *comb.ParserError, interface{}) {
	var zero MO
	var partRes partialMapResult[PO1, PO2, PO3, PO4]

	comb.Debugf("MapN.parseAfterChild - childID=%d, pos=%d", childID, childState.CurrentPos())

	if childID >= 0 { // on the way up: Fetch
		partRes, _ = data.(partialMapResult[PO1, PO2, PO3, PO4])
	}

	if childErr != nil {
		return childState, zero, childErr, partRes
	}

	id := childID // use a new variable to keep the original childID (for distinguishing way: up/down)
	idErr := childState.NewSemanticError("unable to parse after child with unknown ID %d", id)
	if id >= 0 && id != md.p1.ID() {
		if md.n <= 1 {
			return childState, zero, idErr, partRes
		}
		if id != md.p2.ID() {
			if md.n <= 2 {
				return childState, zero, idErr, partRes
			}
			if id != md.p3.ID() {
				if md.n <= 3 {
					return childState, zero, idErr, partRes
				}
				if id != md.p4.ID() {
					if md.n <= 4 {
						return childState, zero, idErr, partRes
					}
					if id != md.p5.ID() {
						return childState, zero, idErr, partRes
					}
				}
			}
		}
	}

	if id < 0 {
		childStartState = childState
		childState, childOut, childErr = md.p1.ParseAny(md.id(), childStartState)
		partRes.out1, _ = childOut.(PO1)
		if childErr != nil {
			return childState, zero, childErr, partRes
		}
	} else if id == md.p1.ID() {
		partRes.out1, _ = childOut.(PO1)
		id = -1
	}

	if md.n > 1 {
		if id < 0 {
			childStartState = childState
			childState, childOut, childErr = md.p2.ParseAny(md.id(), childStartState)
			partRes.out2, _ = childOut.(PO2)
			if childErr != nil {
				out, _ := md.fn(partRes)
				return childState, out, childErr, partRes
			}
		} else if id == md.p2.ID() {
			partRes.out2, _ = childOut.(PO2)
			id = -1
		}

		if md.n > 2 {
			if id < 0 {
				childStartState = childState
				childState, childOut, childErr = md.p3.ParseAny(md.id(), childStartState)
				partRes.out3, _ = childOut.(PO3)
				if childErr != nil {
					out, _ := md.fn(partRes)
					return childState, out, childErr, partRes
				}
			} else if id == md.p3.ID() {
				partRes.out3, _ = childOut.(PO3)
				id = -1
			}

			if md.n > 3 {
				if id < 0 {
					childStartState = childState
					childState, childOut, childErr = md.p4.ParseAny(md.id(), childStartState)
					partRes.out4, _ = childOut.(PO4)
					if childErr != nil {
						out, _ := md.fn(partRes)
						return childState, out, childErr, partRes
					}
				} else if id == md.p4.ID() {
					partRes.out4, _ = childOut.(PO4)
					id = -1
				}

				if md.n > 4 {
					var out5 PO5

					if id < 0 {
						childStartState = childState
						childState, childOut, childErr = md.p5.ParseAny(md.id(), childStartState)
						out5, _ = childOut.(PO5)
						if childErr != nil {
							out, _ := md.fn5(partRes.out1, partRes.out2, partRes.out3, partRes.out4, out5)
							return childState, out, childErr, partRes
						}
					} else {
						out5, _ = childOut.(PO5)
					}

					out, err := md.fn5(partRes.out1, partRes.out2, partRes.out3, partRes.out4, out5)
					if err != nil {
						childState = childState.SaveError(childState.NewSemanticError(err.Error()))
						return childState, out, nil, partRes
					}
					return childState, out, nil, nil
				}

				out, err := md.fn4(partRes.out1, partRes.out2, partRes.out3, partRes.out4)
				if err != nil {
					childState = childState.SaveError(childState.NewSemanticError(err.Error()))
					return childState, out, nil, partRes
				}
				return childState, out, nil, nil
			}

			out, err := md.fn3(partRes.out1, partRes.out2, partRes.out3)
			if err != nil {
				childState = childState.SaveError(childState.NewSemanticError(err.Error()))
				return childState, out, nil, partRes
			}
			return childState, out, nil, nil
		}

		out, err := md.fn2(partRes.out1, partRes.out2)
		if err != nil {
			childState = childState.SaveError(childState.NewSemanticError(err.Error()))
			return childState, out, nil, partRes
		}
		return childState, out, nil, nil
	}

	out, err := md.fn1(partRes.out1)
	if err != nil {
		childState = childState.SaveError(childState.NewSemanticError(err.Error()))
		return childState, out, nil, partRes
	}
	return childState, out, nil, nil
}

func (md *mapData[PO1, PO2, PO3, PO4, PO5, MO]) fn(partRes partialMapResult[PO1, PO2, PO3, PO4]) (MO, error) {
	switch md.n {
	case 1:
		return md.fn1(partRes.out1)
	case 2:
		return md.fn2(partRes.out1, partRes.out2)
	case 3:
		return md.fn3(partRes.out1, partRes.out2, partRes.out3)
	case 4:
		return md.fn4(partRes.out1, partRes.out2, partRes.out3, partRes.out4)
	case 5:
		return md.fn5(partRes.out1, partRes.out2, partRes.out3, partRes.out4, comb.ZeroOf[PO5]())
	}
	return comb.ZeroOf[MO](), nil // can't happen
}
