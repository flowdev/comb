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
	parserIDs    []int32
}

func (pl PrecedenceLevel[Output]) children() []comb.AnyParser {
	return nil
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
	id                func() int32
	expected          string
	value             comb.Parser[Output]
	space             comb.Parser[string]
	levels            []PrecedenceLevel[Output]
	parens            []parens
	openParenParser   comb.Parser[string]
	closeParenParsers map[string]comb.Parser[string]
	saveSpot          bool
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
// The first level added, binds the strongest (e.g., unary sign operator) and
// the last level added binds the least (e.g., assignment operator).
// It's also possible to later add (multiple) pairs of parentheses.
func Expression[Output any](valueParser comb.Parser[Output], levels ...PrecedenceLevel[Output]) *expr[Output] {
	e := &expr[Output]{
		value:  valueParser,
		levels: levels,
	}
	return e
}
func (e *expr[Output]) AddPrefixLevel(level ...PrefixOp[Output]) *expr[Output] {
	e.levels = append(e.levels, PrecedenceLevel[Output]{prefixLevel: level})
	return e
}
func (e *expr[Output]) AddInfixLevel(level ...InfixOp[Output]) *expr[Output] {
	e.levels = append(e.levels, PrecedenceLevel[Output]{infixLevel: level})
	return e
}
func (e *expr[Output]) AddPostfixLevel(level ...PostfixOp[Output]) *expr[Output] {
	e.levels = append(e.levels, PrecedenceLevel[Output]{postfixLevel: level})
	return e
}
func (e *expr[Output]) AddParentheses(open, close string) *expr[Output] {
	e.parens = append(e.parens, parens{open: open, close: close})
	return e
}

// WithSpace sets the parser for handling spaces between tokens in the expression and
// returns the updated expression object.
// If no parser is explicitly set, Whitespace0 is the default.
func (e *expr[Output]) WithSpace(spaceParser comb.Parser[string]) *expr[Output] {
	e.space = spaceParser
	return e
}

// WithExpected sets what kind of expression is expected and
// returns the updated expression object.
// This is used by other parsers embedding this one, like the `Not` parser.
// If nothing is explicitly set, 'expression' is the default.
func (e *expr[Output]) WithExpected(expected string) *expr[Output] {
	e.expected = expected
	return e
}

// Parser performs the last checks and returns the functional expression parser.
// It will panic in the following cases:
//   - double opening parentheses
//   - double operators of the same type (prefix, infix or postfix)
func (e *expr[Output]) Parser() comb.Parser[Output] {
	var p comb.Parser[Output]
	safeSpot := e.checkOperators()
	e.prepareParens()
	e.saveSpot = safeSpot
	if e.space == nil {
		e.space = Whitespace0()
	}
	if e.expected == "" {
		e.expected = "expression"
	}
	e.id = func() int32 { return p.ID() }
	p = comb.NewParserWithData(e.expected, e.parseWithData, e.recover)
	return p
}
func (e *expr[Output]) checkOperators() bool {
	prefixCheck := make(map[string]struct{})
	infixCheck := make(map[string]struct{})
	postfixCheck := make(map[string]struct{})
	safeSpot := false

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
func (e *expr[Output]) prepareParens() {
	if len(e.parens) == 0 {
		return
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
}

// recover finds the operator with minimal waste that has the highest priority.
func (e *expr[Output]) recover(state comb.State, data interface{}) (int, interface{}) {
	return comb.RecoverWasteTooMuch, data
}

func (e *expr[Output]) parseWithData(state comb.State, data interface{}) (comb.State, Output, *comb.ParserError, interface{}) {
	return e.parseLevelAfterError(len(e.levels)-1, state, data)
}
func (e *expr[Output]) parseLevelAfterError(
	l int, state comb.State, data interface{},
) (comb.State, Output, *comb.ParserError, interface{}) {
	var out Output
	var aOut interface{}

	nState, err := e.parseSpace(state)
	if err != nil {
		return nState, out, err, nil
	}
	state = nState

	if l < 0 { // parse value or parentheses
		if e.openParenParser != nil {
			nState, aOut, err = e.openParenParser.ParseAny(e.id(), state)
		}
		if err != nil || e.openParenParser == nil {
			nState, aOut, err = e.value.ParseAny(e.id(), state)
			out, _ = aOut.(Output)
			return nState, out, comb.ClaimError(err), data // TODO: in case of error: return temp data
		}
		state = nState
		openParen, _ := aOut.(string)

		nState, aOut, err, data = e.parseLevelAfterError(len(e.levels)-1, state, data)
		out, _ = aOut.(Output)
		if err != nil {
			return nState, out, err, data
		}
		state = nState

		nState, err = e.parseSpace(state)
		if err != nil {
			return nState, out, err, data
		}
		state = nState

		nState, aOut, err = e.closeParenParsers[openParen].ParseAny(e.id(), state)
		return nState, out, err, data // TODO: in case of error: return temp data
	}

	level := e.levels[l]
	switch {
	case level.prefixLevel != nil:
		return e.parsePrefixLevelAfterError(l, level, state, data)
	case level.infixLevel != nil:
		return e.parseInfixLevelAfterError(l, level, state, data)
	default:
		return e.parsePostfixLevelAfterError(l, level, state, data)
	}
}
func (e *expr[Output]) parsePrefixLevelAfterError(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	data interface{},
) (comb.State, Output, *comb.ParserError, interface{}) {
	var out Output

	state := startState
	nState, aOut, err := level.opParser.ParseAny(e.id(), state)
	if err != nil {
		return e.parseLevelAfterError(l-1, startState, data)
	}
	state = nState
	op, _ := aOut.(string)

	nState, err = e.parseSpace(state)
	if err != nil {
		return nState, out, err, data
	}
	state = nState

	// go recursive to support: '-- ++ a'
	nState, out, err, data = e.parseLevelAfterError(l, state, data)
	if err != nil {
		return nState, out, err, data
	}

	out = level.opFn1s[op](out)
	if level.opSafeSpots[op] {
		nState = nState.MoveSafeSpot()
	}
	return nState, out, nil, data
}
func (e *expr[Output]) parseInfixLevelAfterError(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	data interface{},
) (comb.State, Output, *comb.ParserError, interface{}) {
	var aOut interface{}

	state := startState
	nState, out, err, data2 := e.parseLevelAfterError(l-1, state, data)
	if err != nil {
		return nState, out, err, data2
	}
	state = nState

	for {
		nState, err = e.parseSpace(state)
		if err != nil {
			return nState, out, err, data
		}
		state = nState
		startState = state

		nState, aOut, err = level.opParser.ParseAny(e.id(), state)
		if err != nil {
			return state, out, nil, data
		}
		state = nState
		op, _ := aOut.(string)

		val1 := out
		nState, out, err, data2 = e.parseLevelAfterError(l-1, state, data)
		if err != nil {
			return nState, level.opFn2s[op](val1, out), err, data2
		}
		state = nState

		out = level.opFn2s[op](val1, out)
		if level.opSafeSpots[op] {
			state = nState.MoveSafeSpot()
		}
	}
}
func (e *expr[Output]) parsePostfixLevelAfterError(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	data interface{},
) (comb.State, Output, *comb.ParserError, interface{}) {
	var aOut interface{}

	state := startState
	nState, out, err, data2 := e.parseLevelAfterError(l-1, state, data)
	if err != nil {
		return nState, out, err, data2
	}
	state = nState

	for {
		nState, err = e.parseSpace(state)
		if err != nil {
			return nState, out, err, data
		}
		state = nState

		nState, aOut, err = level.opParser.ParseAny(e.id(), state)
		if err != nil {
			return state, out, nil, data
		}
		state = nState
		op, _ := aOut.(string)

		out = level.opFn1s[op](out)
		if level.opSafeSpots[op] {
			nState = nState.MoveSafeSpot()
		}
	}
}

func (e *expr[Output]) parseSpace(state comb.State) (comb.State, *comb.ParserError) {
	nState, _, err := e.space.ParseAny(e.id(), state)
	if err != nil {
		return state, comb.ClaimError(err)
	}
	return nState, nil
}
