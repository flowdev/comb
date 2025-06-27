package cmb

import (
	"fmt"

	"github.com/flowdev/comb"
)

// ============================================================================
// Parse (Mathematical) Expressions
//

type PrefixOp[Output any] struct {
	Op       string
	SafeSpot bool
	Fn       func(Output) Output
}
type InfixOp[Output any] struct {
	Op       string
	SafeSpot bool
	Fn       func(Output, Output) Output
}
type PostfixOp[Output any] struct {
	Op       string
	SafeSpot bool
	Fn       func(Output) Output
}

type PrecedenceLevel[Output any] struct {
	prefixLevel  []PrefixOp[Output]
	infixLevel   []InfixOp[Output]
	postfixLevel []PostfixOp[Output]
	opParser     comb.Parser[string]
	opFn1s       map[string]func(Output) Output
	opFn2s       map[string]func(Output, Output) Output
	opSafeSpots  map[string]bool
}

// PrefixLevel returns a precedence level for evaluating expressions that
// consists of prefix operators.
// It will panic in the following cases:
//   - empty string for the operator
//   - nil function for the output mapping
//   - double operators
func PrefixLevel[Output any](ops []PrefixOp[Output]) PrecedenceLevel[Output] {
	fn1map := make(map[string]func(Output) Output)
	sops := make([]string, len(ops))
	safeSpots := make(map[string]bool, len(ops))
	for i, op := range ops {
		if op.Op == "" {
			panic(fmt.Sprintf("prefix operation with index %d has no operator", i))
		}
		if op.Fn == nil {
			panic(fmt.Sprintf("prefix operation %q (index %d) has no mapping function", op.Op, i))
		}
		if _, ok := fn1map[op.Op]; ok {
			panic(fmt.Sprintf("prefix operation %q (index %d) is a duplicate", op.Op, i))
		}
		sops[i] = op.Op
		fn1map[op.Op] = op.Fn
		safeSpots[op.Op] = op.SafeSpot
	}
	return PrecedenceLevel[Output]{prefixLevel: ops, opParser: OneOf(sops...), opFn1s: fn1map, opSafeSpots: safeSpots}
}

// InfixLevel returns a precedence level for evaluating expressions that
// consists of infix operators.
// It will panic in the following cases:
//   - empty string for the operator
//   - nil function for the output mapping
//   - double operators
func InfixLevel[Output any](ops []InfixOp[Output]) PrecedenceLevel[Output] {
	fn2map := make(map[string]func(Output, Output) Output)
	sops := make([]string, len(ops))
	safeSpots := make(map[string]bool, len(ops))
	for i, op := range ops {
		if op.Op == "" {
			panic(fmt.Sprintf("infix operation with index %d has no operator", i))
		}
		if op.Fn == nil {
			panic(fmt.Sprintf("infix operation %q (index %d) has no mapping function", op.Op, i))
		}
		if _, ok := fn2map[op.Op]; ok {
			panic(fmt.Sprintf("prefix operation %q (index %d) is a duplicate", op.Op, i))
		}
		sops[i] = op.Op
		fn2map[op.Op] = op.Fn
		safeSpots[op.Op] = op.SafeSpot
	}
	if len(fn2map) < len(ops) {
		panic(fmt.Sprintf("unable to use double infix operator: got %q, only %d are unique", sops, len(fn2map)))
	}
	return PrecedenceLevel[Output]{infixLevel: ops, opParser: OneOf(sops...), opFn2s: fn2map, opSafeSpots: safeSpots}
}

// PostfixLevel returns a precedence level for evaluating expressions that
// consists of postfix operators.
// It will panic in the following cases:
//   - empty string for the operator
//   - nil function for the output mapping
//   - double operators
func PostfixLevel[Output any](ops []PostfixOp[Output]) PrecedenceLevel[Output] {
	fn1map := make(map[string]func(Output) Output)
	sops := make([]string, len(ops))
	safeSpots := make(map[string]bool, len(ops))
	for i, op := range ops {
		if op.Op == "" {
			panic(fmt.Sprintf("postfix operation with index %d has no operator", i))
		}
		if op.Fn == nil {
			panic(fmt.Sprintf("postfix operation %q (index %d) has no mapping function", op.Op, i))
		}
		if _, ok := fn1map[op.Op]; ok {
			panic(fmt.Sprintf("postfix operation %q (index %d) is a duplicate", op.Op, i))
		}
		sops[i] = op.Op
		fn1map[op.Op] = op.Fn
		safeSpots[op.Op] = op.SafeSpot
	}
	if len(fn1map) < len(ops) {
		panic(fmt.Sprintf("unable to use double postfix operator: got %q, only %d are unique", sops, len(fn1map)))
	}
	return PrecedenceLevel[Output]{postfixLevel: ops, opParser: OneOf(sops...), opFn1s: fn1map, opSafeSpots: safeSpots}
}

type expr[Output any] struct {
	value             comb.Parser[Output]
	space             comb.Parser[string]
	mustSpace         bool
	levels            []PrecedenceLevel[Output]
	parens            []parens
	openParenParser   comb.Parser[string]
	closeParenParsers map[string]comb.Parser[string]
	subParser         comb.Parser[levelIdx] // do we use this, or many sub-parsers instead?
}
type parens struct {
	open, close string
}
type levelIdx struct {
	level int
	idx   int
}

// Expression returns a branch parser for parsing (mathematical) expressions
// with prefix, infix and postfix operators.
// PrecedenceLevel s can be set in this function call or added one by one later.
// Each PrecedenceLevel can only contain either all prefix or all infix or all postfix operators.
// Within each level evaluation is always from left to right.
// The order of the levels matters and is similar to FirstSuccessful.
// The first level added, binds the strongest (e.g. unary sign operator) and
// the last level added binds the least (e.g. assignment operator).
// It's also possible to later add (multiple) pairs of parentheses.
func Expression[Output any](valueParser comb.Parser[Output], levels ...PrecedenceLevel[Output]) expr[Output] {
	e := expr[Output]{
		value:  valueParser,
		levels: levels,
	}
	return e
}
func (e expr[Output]) AddPrefixLevel(level ...PrefixOp[Output]) expr[Output] {
	e.levels = append(e.levels, PrecedenceLevel[Output]{prefixLevel: level})
	return e
}
func (e expr[Output]) AddInfixLevel(level ...InfixOp[Output]) expr[Output] {
	e.levels = append(e.levels, PrecedenceLevel[Output]{infixLevel: level})
	return e
}
func (e expr[Output]) AddPostfixLevel(level ...PostfixOp[Output]) expr[Output] {
	e.levels = append(e.levels, PrecedenceLevel[Output]{postfixLevel: level})
	return e
}
func (e expr[Output]) AddParentheses(open, close string) expr[Output] {
	e.parens = append(e.parens, parens{open: open, close: close})
	return e
}
func (e expr[Output]) SetSpace(spaceParser comb.Parser[string], mandatory bool) expr[Output] {
	e.space = spaceParser
	e.mustSpace = mandatory
	return e
}

// Parser performs the last checks and returns the functional expression parser.
// It will panic in the following cases:
//   - double opening parentheses
//   - double operators of the same type (prefix, infix or postfix)
func (e expr[Output]) Parser() comb.Parser[Output] {
	safeSpot := e.checkOperators()
	e = e.prepareParens()
	e.subParser = comb.NewParser[levelIdx]("operator", e.subParse, e.recover)
	if safeSpot {
		e.subParser = comb.SafeSpot(e.subParser)
	}
	if e.space == nil {
		e.space = Whitespace0()
	}
	return comb.NewBranchParser[Output]("Expression", e.children, e.parseAfterChild)
}
func (e expr[Output]) checkOperators() bool {
	prefixCheck := make(map[string]struct{})
	infixCheck := make(map[string]struct{})
	postfixCheck := make(map[string]struct{})
	safeSpot := e.value.IsSaveSpot()

	for _, level := range e.levels {
		switch {
		case level.prefixLevel != nil:
			for _, op := range level.prefixLevel {
				if _, ok := prefixCheck[op.Op]; ok {
					panic(fmt.Sprintf("prefix operation %q is a duplicate", op.Op))
				}
				prefixCheck[op.Op] = struct{}{}
				if op.SafeSpot {
					safeSpot = true
				}
			}
		case level.infixLevel != nil:
			for _, op := range level.infixLevel {
				if _, ok := infixCheck[op.Op]; ok {
					panic(fmt.Sprintf("infix operation %q is a duplicate", op.Op))
				}
				infixCheck[op.Op] = struct{}{}
				if op.SafeSpot {
					safeSpot = true
				}
			}
		default:
			for _, op := range level.postfixLevel {
				if _, ok := postfixCheck[op.Op]; ok {
					panic(fmt.Sprintf("postfix operation %q is a duplicate", op.Op))
				}
				postfixCheck[op.Op] = struct{}{}
				if op.SafeSpot {
					safeSpot = true
				}
			}
		}
	}
	return safeSpot
}
func (e expr[Output]) prepareParens() expr[Output] {
	if len(e.parens) == 0 {
		return e
	}
	opens := make([]string, len(e.parens))
	parsers := make(map[string]comb.Parser[string], len(e.parens))
	check := make(map[string]struct{}, len(e.parens))

	for i, paren := range e.parens {
		if _, ok := check[paren.open]; ok {
			panic(fmt.Sprintf("opening parentheses %q (index %d) is already defined", paren.open, i))
		}
		check[paren.open] = struct{}{}
		opens[i] = paren.open
		parsers[paren.open] = String(paren.close)
	}
	e.openParenParser = OneOf(opens...)
	e.closeParenParsers = parsers
	return e
}

func (e expr[Output]) children() []comb.AnyParser {
	return []comb.AnyParser{e.value, e.subParser, e.space}
}

// subParse is only used during error recovery.
// We don't know which operator to use. But we know that one operator matches immediately.
// We just have to try them in the same order as the recover method.
func (e expr[Output]) subParse(state comb.State) (comb.State, levelIdx, *comb.ParserError) {
	return state, levelIdx{}, nil
}

// recover finds the operator with minimal waste that has the highest priority.
func (e expr[Output]) recover(pe *comb.ParserError, state comb.State) int {
	return comb.RecoverWasteTooMuch
}

func (e expr[Output]) parseAfterChild(id int32, result comb.ParseResult) comb.ParseResult {
	if id >= 0 {
		// TODO: implement partial parsing
		return result
	}
	return e.parseLevel(len(e.levels)-1, result)
}
func (e expr[Output]) parseLevel(l int, result comb.ParseResult) comb.ParseResult {
	nResult := e.parseSpace(result)
	if nResult.Error != nil {
		return nResult
	}
	if l < 0 { // parse value or parentheses
		if e.openParenParser == nil {
			return comb.RunParser(e.value, nResult)
		}
		oResult := comb.RunParser(e.openParenParser, nResult)
		if oResult.Error != nil {
			return comb.RunParser(e.value, nResult)
		}
		openParen, _ := oResult.Output.(string)
		pResult := e.parseLevel(len(e.levels)-1, oResult)
		if pResult.Error != nil {
			return pResult
		}
		qResult := e.parseSpace(pResult)
		if qResult.Error != nil {
			return qResult
		}
		rResult := comb.RunParser(e.closeParenParsers[openParen], qResult)
		if qResult.Error != nil {
			// TODO: partial result?
			return rResult
		}
		rResult.Output = pResult.Output
		return rResult
	}

	level := e.levels[l]
	switch {
	case level.prefixLevel != nil:
		return e.parsePrefixLevel(l, level, result.EndState, nResult)
	case level.infixLevel != nil:
		return e.parseInfixLevel(l, level, result.EndState, nResult)
	default:
		return e.parsePostfixLevel(l, level, result.EndState, nResult)
	}
}
func (e expr[Output]) parsePrefixLevel(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	nResult comb.ParseResult,
) comb.ParseResult {
	oResult := comb.RunParser(level.opParser, nResult)
	if oResult.Error != nil {
		return e.parseLevel(l-1, nResult)
	}
	op, _ := oResult.Output.(string)
	pResult := e.parseLevel(l, oResult)
	if pResult.Error != nil {
		return e.parseLevel(l-1, nResult)
	}
	val, _ := pResult.Output.(Output)
	pResult.Output = level.opFn1s[op](val)
	pResult.StartState = startState
	if level.opSafeSpots[op] {
		pResult.EndState = pResult.EndState.MoveSafeSpot()
	}
	return pResult
}
func (e expr[Output]) parseInfixLevel(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	nResult comb.ParseResult,
) comb.ParseResult {
	oResult := e.parseLevel(l-1, nResult)
	for oResult.Error == nil {
		pResult := e.parseSpace(oResult)
		if pResult.Error != nil {
			return oResult
		}
		qResult := comb.RunParser(level.opParser, pResult)
		if qResult.Error != nil {
			return oResult
		}
		op, _ := qResult.Output.(string)
		rResult := e.parseLevel(l-1, qResult)
		if rResult.Error != nil {
			// TODO: save partial result
			return rResult
		}
		val1, _ := oResult.Output.(Output)
		val2, _ := rResult.Output.(Output)
		rResult.Output = level.opFn2s[op](val1, val2)
		rResult.StartState = startState
		if level.opSafeSpots[op] {
			rResult.EndState = rResult.EndState.MoveSafeSpot()
		}
		oResult = rResult
	}
	return oResult
}
func (e expr[Output]) parsePostfixLevel(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	nResult comb.ParseResult,
) comb.ParseResult {
	oResult := e.parseLevel(l-1, nResult)
	if oResult.Error != nil {
		return oResult
	}
	for oResult.Error == nil {
		oResult = e.parseSpace(oResult)
		if oResult.Error != nil {
			return oResult
		}

		pResult := comb.RunParser(level.opParser, oResult)
		if pResult.Error != nil {
			return oResult
		}
		op, _ := pResult.Output.(string)
		val, _ := oResult.Output.(Output)
		pResult.Output = level.opFn1s[op](val)
		if level.opSafeSpots[op] {
			pResult.EndState = pResult.EndState.MoveSafeSpot()
		}
		oResult = pResult
	}
	oResult.StartState = startState
	return oResult
}
func (e expr[Output]) parseSpace(result comb.ParseResult) comb.ParseResult {
	state, _, err := e.space.Parse(result.EndState)
	if err == nil {
		result.StartState = result.EndState
		result.EndState = state
	} else if e.mustSpace {
		result.StartState = result.EndState
		result.Error = err
	}
	return result
}
