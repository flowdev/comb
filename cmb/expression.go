package cmb

import (
	"fmt"
	"math"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/flowdev/comb"
	"github.com/flowdev/comb/x/omap"
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
	opsText      string
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
	return PrecedenceLevel[Output]{
		prefixLevel: ops,
		opFn1s:      fn1map,
		opSafeSpots: safeSpots,
		opsText:     fmt.Sprintf("%q", sops),
	}
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
			panic(fmt.Sprintf("infix operation %q (index %d) is a duplicate", op.Op, i))
		}
		sops[i] = op.Op
		fn2map[op.Op] = op.Fn
		safeSpots[op.Op] = op.SafeSpot
	}
	return PrecedenceLevel[Output]{
		infixLevel:  ops,
		opFn2s:      fn2map,
		opSafeSpots: safeSpots,
		opsText:     fmt.Sprintf("%q", sops),
	}
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
	return PrecedenceLevel[Output]{
		postfixLevel: ops,
		opFn1s:       fn1map,
		opSafeSpots:  safeSpots,
		opsText:      fmt.Sprintf("%q", sops),
	}
}

type expr[Output any] struct {
	id                func() int32
	expected          string
	value             comb.Parser[Output]
	space             comb.Parser[string]
	levels            []PrecedenceLevel[Output]
	parens            []parens
	openParenParser   comb.Parser[string]
	closeParenParser  comb.Parser[string]
	closeParenParsers map[string]comb.Parser[string]
	safeSpots         []safeSpot
	recoverCache      *omap.OrderedMap[int, safeSpot]
}
type parens struct {
	open, close string
	safeSpot    bool
}

type safeSpot struct {
	op  string
	l   int
	rec comb.AnyParser
}

type recoverData[Output any] struct {
	lData         []levelData[Output]
	safeSpotLevel int
	safeSpotOp    string
}

// levelData stores partial output and other data of each level.
// It's used as []levelData in practice with a length of len(levels) + 1
// because of the level -1 for values and parentheses.
// A value of 0 for exit signals that there is no data for the level.
type levelData[Output any] struct {
	out    Output
	op     string
	preOps []string
	exit   int
}

// Expression returns a branch parser for parsing (mathematical) expressions
// with prefix, infix and postfix operators.
// The valueParser should be a SafeSpot parser if reasonable.
// It's also very good to turn all operators into safe spots, as long as they aren't used in other contexts, too.
// The valueParser MUST be a simple parser that doesn't need any data for error recovery.
//
// PrecedenceLevel s can be set in this function call or added one by one later.
// Each PrecedenceLevel can only contain either all prefix or all infix or all postfix operators.
// Within each level evaluation is always from left to right.
// The order of the levels matters and is similar to FirstSuccessful.
// The first level added, binds the strongest (e.g., unary sign operator) and
// the last level added binds the least (e.g., assignment operator).
// It's also possible to later add (multiple) pairs of parentheses.
//
// The Expression parser is a safe spot parser iff the valueParser is or
// one of its operators is marked as a safe spot.
//
// The Expression parser will panic in the following cases:
//   - empty string for any operator
//   - nil function for output calculation
//   - double operators of the same type (prefix, infix or postfix)
//   - double opening parentheses
func Expression[Output any](valueParser comb.Parser[Output], levels ...PrecedenceLevel[Output]) expr[Output] {
	e := expr[Output]{
		value:  valueParser,
		levels: levels,
	}
	return e
}
func (e expr[Output]) AddPrefixLevel(level ...PrefixOp[Output]) expr[Output] {
	e.levels = append(e.levels, PrefixLevel(level))
	return e
}
func (e expr[Output]) AddInfixLevel(level ...InfixOp[Output]) expr[Output] {
	e.levels = append(e.levels, InfixLevel(level))
	return e
}
func (e expr[Output]) AddPostfixLevel(level ...PostfixOp[Output]) expr[Output] {
	e.levels = append(e.levels, PostfixLevel(level))
	return e
}
func (e expr[Output]) AddParentheses(open, close string, safeSpot bool) expr[Output] {
	e.parens = append(e.parens, parens{open: open, close: close, safeSpot: safeSpot})
	return e
}

// WithSpace sets the parser for handling spaces between tokens in the expression and
// returns the updated expression object.
// If no parser is explicitly set, Whitespace0 is the default.
func (e expr[Output]) WithSpace(spaceParser comb.Parser[string]) expr[Output] {
	e.space = spaceParser
	return e
}

// WithExpected sets what kind of expression is expected and
// returns the updated expression object.
// This is used by other parsers embedding this one, like the `Not` parser.
// If nothing is explicitly set, 'expression' is the default.
func (e expr[Output]) WithExpected(expected string) expr[Output] {
	e.expected = expected
	return e
}

// Parser performs the last checks and returns the functional expression parser.
// It will panic in the following cases:
//   - double opening parentheses
//   - double operators of the same type (prefix, infix or postfix)
func (e expr[Output]) Parser() comb.Parser[Output] {
	var p comb.Parser[Output]

	ee := e.prepareParens()
	ee = ee.checkOperators()
	if ee.space == nil {
		ee.space = Whitespace0()
	}
	if ee.expected == "" {
		ee.expected = "expression"
	}
	ee.levels = append([]PrecedenceLevel[Output]{{}}, ee.levels...) // add level for values and parentheses
	ee.id = func() int32 { return p.ID() }
	ee.recoverCache = omap.New[int, safeSpot](len(ee.safeSpots))
	p = comb.NewParserWithData(ee.expected, ee.parseWithData, ee.recover)
	if len(ee.safeSpots) > 0 {
		return comb.SafeSpot(p)
	}
	return p
}
func (e expr[Output]) checkOperators() expr[Output] {
	prefixCheck := make(map[string]struct{})
	infixCheck := make(map[string]struct{})
	postfixCheck := make(map[string]struct{})
	safeSpots := make([]safeSpot, 0, 64)

	safeOpenParens := make([]string, 0, len(e.parens))
	safeCloseParens := make([]string, 0, len(e.parens))
	for _, paren := range e.parens {
		if paren.safeSpot {
			safeOpenParens = append(safeOpenParens, paren.open)
			safeCloseParens = append(safeCloseParens, paren.close)
		}
	}
	if len(safeOpenParens) > 0 {
		safeSpots = append(safeSpots, safeSpot{op: "(", l: 0, rec: OneOf(safeOpenParens...)})
	}
	if e.value.IsSafeSpot() {
		safeSpots = append(safeSpots, safeSpot{op: "value", l: 0, rec: e.value})
	}
	if len(safeCloseParens) > 0 {
		safeSpots = append(safeSpots, safeSpot{op: ")", l: 0, rec: OneOf(safeCloseParens...)})
	}
	for l, level := range e.levels {
		sops := make([]string, len(level.prefixLevel)+len(level.infixLevel)+len(level.postfixLevel))
		switch {
		case level.prefixLevel != nil:
			for i, op := range level.prefixLevel {
				if _, ok := prefixCheck[op.Op]; ok {
					panic(fmt.Sprintf("prefix operation %q is a duplicate", op.Op))
				}
				prefixCheck[op.Op] = struct{}{}
				if op.SafeSpot {
					safeSpots = append(safeSpots, safeSpot{op: op.Op, l: l + 1, rec: e.oneOfOperator(op.Op)})
				}
				sops[i] = op.Op
			}
		case level.infixLevel != nil:
			for i, op := range level.infixLevel {
				if _, ok := infixCheck[op.Op]; ok {
					panic(fmt.Sprintf("infix operation %q is a duplicate", op.Op))
				}
				infixCheck[op.Op] = struct{}{}
				if op.SafeSpot {
					safeSpots = append(safeSpots, safeSpot{op: op.Op, l: l + 1, rec: e.oneOfOperator(op.Op)})
				}
				sops[i] = op.Op
			}
		default:
			for i, op := range level.postfixLevel {
				if _, ok := postfixCheck[op.Op]; ok {
					panic(fmt.Sprintf("postfix operation %q is a duplicate", op.Op))
				}
				postfixCheck[op.Op] = struct{}{}
				if op.SafeSpot {
					safeSpots = append(safeSpots, safeSpot{op: op.Op, l: l + 1, rec: e.oneOfOperator(op.Op)})
				}
				sops[i] = op.Op
			}
		}
		e.levels[l].opParser = e.oneOfOperator(sops...)
	}
	e.safeSpots = safeSpots
	return e
}
func (e expr[Output]) prepareParens() expr[Output] {
	if len(e.parens) == 0 {
		return e
	}
	opens := make([]string, len(e.parens))
	closes := make([]string, len(e.parens))
	parsers := make(map[string]comb.Parser[string], len(e.parens))
	check := make(map[string]struct{}, len(e.parens))

	for i, paren := range e.parens {
		if _, ok := check[paren.open]; ok {
			panic(fmt.Sprintf("opening parentheses %q (index %d) is already defined", paren.open, i))
		}
		check[paren.open] = struct{}{}
		opens[i] = paren.open
		closes[i] = paren.close
		parsers[paren.open] = String(paren.close)
	}
	e.openParenParser = OneOf(opens...)
	e.closeParenParser = OneOf(closes...)
	e.closeParenParsers = parsers
	return e
}
func (e expr[Output]) oneOfOperator(collection ...string) comb.Parser[string] {
	n := len(collection)
	if n == 0 {
		panic("oneOfOperator has no strings to match")
	}
	expected := fmt.Sprintf("one operator of %q", collection)

	parse := func(state comb.State) (comb.State, string, *comb.ParserError) {
		input := state.CurrentString()
		for _, token := range collection {
			if strings.HasPrefix(input, token) {
				nState := state.MoveBy(len(token))
				if ok, _ := isEndOfOp(nState, e.openParenParser, e.closeParenParser); ok {
					return nState, token, nil
				}
			}
		}
		return state, "", state.NewSyntaxError(expected)
	}

	return comb.NewParser[string](expected, parse, e.indexOfAnyOperator(collection...))
}
func (e expr[Output]) indexOfAnyOperator(stops ...string) comb.Recoverer {
	n := len(stops)

	if n == 0 {
		panic("no operators provided")
	}

	return func(state comb.State, _ interface{}) (int, interface{}) {
		orgInput := state.CurrentString()
		pos := comb.RecoverWasteTooMuch
		for i := 0; i < n; i++ {
			input := orgInput
			start := 0
			stopLen := len(stops[i])
			found := false // we might have to try multiple times because sometimes it just looks like the op but isn't
			for !found {   // e.g.: "++" instead of "+"
				switch j := strings.Index(input, stops[i]); j {
				case -1: // ignore
					found = true
				case 0: // it won't get better than this
					nState := state.MoveBy(stopLen)
					opLen := endOfOp(nState, e.openParenParser, e.closeParenParser)
					if opLen == 0 {
						if pos < 0 || start < pos {
							if start == 0 {
								return 0, nil
							}
							pos = start
							found = true
						}
					} else {
						start += stopLen + opLen
						input = input[stopLen+opLen:]
					}
				default:
					if pos < 0 || start+j < pos {
						nState := state.MoveBy(start + j + stopLen)
						opLen := endOfOp(nState, e.openParenParser, e.closeParenParser)
						if opLen == 0 {
							pos = start + j
							found = true
						} else {
							start += j + stopLen + opLen
							input = input[j+stopLen+opLen:]
						}
					}
				}
			}
		}
		return pos, nil
	}
}
func endOfOp(state comb.State, openParenParser, closeParenParser comb.Parser[string]) int {
	end := 0
	for {
		found, rsize := isEndOfOp(state, openParenParser, closeParenParser)
		if found {
			return end
		}
		end += rsize
		state = state.MoveBy(rsize)
	}
}
func isEndOfOp(state comb.State, openParenParser, closeParenParser comb.Parser[string]) (bool, int) {
	if state.AtEnd() {
		return true, 0
	}
	r, rsize := utf8.DecodeRuneInString(state.CurrentString())
	if r != utf8.RuneError {
		if IsAlphanumeric(r) || unicode.IsSpace(r) {
			return true, 0
		}
	}
	if openParenParser != nil {
		if _, _, err := openParenParser.Parse(state); err == nil {
			return true, 0
		}
	}
	if closeParenParser != nil {
		if _, _, err := closeParenParser.Parse(state); err == nil {
			return true, 0
		}
	}
	return false, rsize
}

// recover finds the operator with minimal waste that has the highest priority.
func (e expr[Output]) recover(state comb.State, data interface{}) (int, interface{}) {
	rData, _ := data.(*recoverData[Output])
	if rData == nil {
		rData = &recoverData[Output]{lData: make([]levelData[Output], len(e.levels))}
	}

	pID := e.id()
	_, _ = fmt.Fprintf(os.Stderr, "\n\nERROR?: pID: %d\n", pID)
	pos := state.CurrentPos()
	icache := state.GetFromCache(pID)
	_, _ = fmt.Fprintf(os.Stderr, "ERROR?: pos: %d\n", pos)
	_, _ = fmt.Fprintf(os.Stderr, "ERROR?: icache: %v\n", icache)
	var cache *omap.OrderedMap[int, safeSpot]
	var cache2 *omap.OrderedMap[int, safeSpot]
	if icache == nil {
		//if e.recoverCache.Len() == 0 {
		cache2 = e.recoverCache
		cache = omap.New[int, safeSpot](len(e.safeSpots))
		for i, ss := range e.safeSpots {
			waste, _ := ss.rec.Recover(state, nil)
			if waste < 0 {
				cache.Add(math.MaxInt-i, ss)  // don't add them all to the same spot
				cache2.Add(math.MaxInt-i, ss) // don't add them all to the same spot
			} else {
				cache.Add(pos+waste, ss)
				cache2.Add(pos+waste, ss)
			}
		}
		if pID >= 0 { // don't cache parsers created by SafeSpot to find Forbidden recoverers
			state.PutIntoCache(pID, cache)
		}
		e.recoverCache = cache2
	} else {
		cache = icache.(*omap.OrderedMap[int, safeSpot])
		cache2 = e.recoverCache
	}

	n := state.CurrentPos() + state.BytesRemaining()
	//cache = cache2
	for {
		npos, ss := cache.GetFirst()
		if npos >= pos {
			if npos >= n {
				return comb.RecoverWasteTooMuch, rData
			}
			rData.safeSpotOp = ss.op
			rData.safeSpotLevel = ss.l
			return npos - pos, rData
		} else {
			waste, _ := ss.rec.Recover(state, nil)
			if waste < 0 {
				cache.ReplaceFirst(math.MaxInt, ss)
			} else {
				cache.ReplaceFirst(pos+waste, ss)
			}
		}
	}
}

func (e expr[Output]) parseWithData(state comb.State, data interface{}) (comb.State, Output, *comb.ParserError, interface{}) {
	rData, _ := data.(*recoverData[Output])
	return e.parseLevelWithData(len(e.levels)-1, state, rData)
}
func (e expr[Output]) parseLevelWithData(
	l int, state comb.State, data *recoverData[Output],
) (comb.State, Output, *comb.ParserError, *recoverData[Output]) {
	if l == 0 { // parse value or parentheses
		return e.parseValueLevelWithData(state, data)
	}
	switch {
	case e.levels[l].prefixLevel != nil:
		return e.parsePrefixLevelWithData(l, e.levels[l], state, data)
	case e.levels[l].infixLevel != nil:
		return e.parseInfixLevelWithData(l, e.levels[l], state, data)
	default:
		return e.parsePostfixLevelWithData(l, e.levels[l], state, data)
	}
}
func (e expr[Output]) parseValueLevelWithData(
	startState comb.State,
	data *recoverData[Output],
) (comb.State, Output, *comb.ParserError, *recoverData[Output]) {
	var out Output
	var err *comb.ParserError
	var rData *recoverData[Output]

	if data == nil {
		rData = &recoverData[Output]{lData: make([]levelData[Output], len(e.levels))}
	} else {
		rData = data
	}
	state := startState
	nState := state

	if data == nil {
		nState, err = e.parseSpace(state)
		if err != nil {
			rData.lData[0] = levelData[Output]{exit: 1}
			return nState, out, err, rData // exit 1
		}
		state = nState
	}

	openParen := ""
	if e.openParenParser != nil && (data == nil || data.safeSpotOp == "(") {
		nState, openParen, err = e.openParenParser.Parse(state)
	}
	if err != nil || e.openParenParser == nil || (data != nil && data.safeSpotOp == "value") {
		nState, out, err = e.value.Parse(state)
		if err != nil {
			rData.lData[0] = levelData[Output]{exit: 2, out: out}
			return state, out, comb.ClaimError(err), rData // exit 2
		}
		return nState, out, nil, nil
	}
	state = nState

	if data == nil || data.safeSpotOp == "(" {
		nState, out, err, data = e.parseLevelWithData(len(e.levels)-1, state, nil)
		if err != nil {
			rData.lData[0] = levelData[Output]{exit: 3, out: out, op: openParen}
			return nState, out, err, rData // exit 3
		}
		state = nState

		nState, err = e.parseSpace(state)
		if err != nil {
			rData.lData[0] = levelData[Output]{exit: 4, out: out, op: openParen}
			return nState, out, err, rData // exit 4
		}
		state = nState
	} else {
		out = rData.lData[0].out
	}

	// special case: the closing parenthesis is the safe spot
	if e.closeParenParser != nil && data != nil && data.safeSpotOp == ")" {
		nState, _, err = e.closeParenParser.Parse(state)
		if err != nil {
			rData.lData[0].exit = 5
			return nState, out, comb.ClaimError(err), rData // exit 5
		}
		return nState, rData.lData[0].out, nil, nil
	}

	nState, _, err = e.closeParenParsers[openParen].Parse(state)
	if err != nil {
		rData.lData[0] = levelData[Output]{exit: 6, out: out, op: openParen}
		return state, out, comb.ClaimError(err), rData // exit 6
	}
	return nState, out, nil, nil
}
func (e expr[Output]) parsePrefixLevelWithData(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	data *recoverData[Output],
) (comb.State, Output, *comb.ParserError, *recoverData[Output]) {
	var out Output
	var err *comb.ParserError
	var rData *recoverData[Output]

	parseSpace, parseOp, parseVal2 := prefixParseCase(l, data)

	if data == nil {
		rData = &recoverData[Output]{lData: make([]levelData[Output], len(e.levels))}
	} else {
		rData = data
	}
	state := startState
	nState := state
	op := ""

	if parseSpace {
		nState, err = e.parseSpace(state)
		if err != nil {
			nState, out, err, rData = e.parseLevelWithData(l-1, startState, data) // we can't parse, maybe the next level can
			if err != nil {
				err.PatchMessage("prefix operator " + level.opsText + " or ")
				rData.lData[l] = levelData[Output]{exit: 1, out: out}
				return nState, out, err, rData
			}
			return nState, out, nil, nil
		}
		state = nState
	}
	if parseOp {
		nState, op, err = level.opParser.Parse(state)
		if err != nil {
			nState, out, err, rData = e.parseLevelWithData(l-1, startState, data) // we can't parse, maybe the next level can
			if err != nil {
				err.PatchMessage("prefix operator " + level.opsText + " or ")
				if len(rData.lData[l].preOps) == 0 {
					rData.lData[l] = levelData[Output]{exit: 2, out: out}
				}
				return nState, out, err, rData
			}
			return nState, out, nil, nil
		}
		state = nState
	} else {
		if len(rData.lData[l].preOps) > 0 {
			op = rData.lData[l].preOps[len(rData.lData[l].preOps)-1]
			rData.lData[l].preOps = rData.lData[l].preOps[:len(rData.lData[l].preOps)-1]
		}
	}
	safeOps := rData.lData[l].preOps
	if parseVal2 {
		// go recursive to support: '-- ++ a'
		if parseOp {
			nState, out, err, rData = e.parseLevelWithData(l, state, nil)
		} else {
			nState, out, err, rData = e.parseLevelWithData(l-1, startState, data) // we didn't parse, maybe the next level will
			if err != nil {
				err.PatchMessage("prefix operator " + level.opsText + " or ")
			}
		}
		if err != nil {
			if op != "" {
				safeOps = append(safeOps, op)
			}
			rData.lData[l] = levelData[Output]{exit: 3, out: out, preOps: safeOps}
			return nState, out, err, rData
		}
	}

	if op != "" {
		out = level.opFn1s[op](out)
	}
	for i := len(safeOps) - 1; i >= 0; i-- {
		out = level.opFn1s[safeOps[i]](out)
	}
	if level.opSafeSpots[op] {
		nState = nState.MoveSafeSpot()
	}
	return nState, out, nil, nil
}
func (e expr[Output]) parseInfixLevelWithData(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	data *recoverData[Output],
) (comb.State, Output, *comb.ParserError, *recoverData[Output]) {
	var out Output
	var err *comb.ParserError
	var rData *recoverData[Output]

	parseVal1, parseSpace, parseOp, parseVal2 := infixParseCase(l, data)

	if data == nil {
		rData = &recoverData[Output]{lData: make([]levelData[Output], len(e.levels))}
	} else {
		rData = data
	}
	state := startState
	nState := state
	data2 := data
	op := ""

	if parseVal1 {
		nState, out, err, data2 = e.parseLevelWithData(l-1, state, data)
		if err != nil {
			err.PatchMessage("infix operator " + level.opsText + " or ")
			rData = data2
			rData.lData[l] = levelData[Output]{exit: 1, out: out}
			return nState, out, err, rData // exit 1
		}
		state = nState
		if rData.lData[l].op != "" {
			out = level.opFn2s[rData.lData[l].op](rData.lData[l].out, out)
		}
	} else {
		out = rData.lData[l].out
	}
	for {
		startState = state
		if parseSpace {
			nState, err = e.parseSpace(state)
			if err != nil {
				return state, out, nil, nil // good case
			}
			state = nState
		}
		parseSpace = true
		if parseOp {
			nState, op, err = level.opParser.Parse(state)
			if err != nil {
				return startState, out, nil, nil // good case
			}
			state = nState
		} else {
			op = rData.lData[l].op
		}
		parseOp = true
		val1 := out
		if parseVal2 {
			nState, out, err, data2 = e.parseLevelWithData(l-1, state, nil)
			if err != nil {
				err.PatchMessage("infix operator " + level.opsText + " or ")
				rData = data2
				rData.lData[l] = levelData[Output]{exit: 2, out: val1, op: op}
				return nState, level.opFn2s[op](val1, out), err, rData // exit 2
			}
			state = nState
		}
		parseVal2 = true

		if op != "" {
			out = level.opFn2s[op](val1, out)
		}
		if level.opSafeSpots[op] {
			state = nState.MoveSafeSpot()
		}
	}
}
func (e expr[Output]) parsePostfixLevelWithData(
	l int,
	level PrecedenceLevel[Output],
	startState comb.State,
	data *recoverData[Output],
) (comb.State, Output, *comb.ParserError, *recoverData[Output]) {
	var out Output
	var err *comb.ParserError
	var rData *recoverData[Output]

	parseVal1, parseSpace, parseOp := postfixParseCase(l, data)

	if data == nil {
		rData = &recoverData[Output]{lData: make([]levelData[Output], len(e.levels))}
	} else {
		rData = data
	}
	state := startState
	nState := state
	data2 := data
	op := ""

	if parseVal1 {
		nState, out, err, data2 = e.parseLevelWithData(l-1, state, data)
		if err != nil {
			err.PatchMessage("postfix operator " + level.opsText + " or ")
			rData = data2
			rData.lData[l] = levelData[Output]{exit: 1, out: out}
			return nState, out, err, rData // exit 1
		}
		state = nState
		startState = nState
	} else {
		out = rData.lData[l].out
	}
	for {
		if parseSpace {
			nState, err = e.parseSpace(state)
			if err != nil {
				return nState, out, nil, nil // not a real error
			}
			state = nState
		}
		parseSpace = true
		if parseOp {
			nState, op, err = level.opParser.Parse(state)
			if err != nil {
				return startState, out, nil, nil // not a real error
			}
			state = nState
		}
		parseOp = true

		if op != "" {
			out = level.opFn1s[op](out)
		}
		if level.opSafeSpots[op] {
			nState = nState.MoveSafeSpot()
		}
		state = nState
		startState = nState
	}
}

func prefixParseCase[Output any](l int, data *recoverData[Output]) (parseSpace, parseOp, parseVal2 bool) {
	if data == nil { // CASE1: no error => parse normally from the beginning
		return true, true, true
	}
	if data.safeSpotOp != "" && data.safeSpotLevel == l { // CASE2: we are the safe spot parser; don't call lower level
		return false, true, true // clean up own lData
	}
	// CASE3: data.safeSpotOp != "" && data.safeSpotLevel < l => we should call the next lower level and use its out as value
	return false, false, true // we will get safeSpot value; no parsing of op before value
}
func infixParseCase[Output any](l int, data *recoverData[Output]) (parseVal1, parseSpace, parseOp, parseVal2 bool) {
	if data == nil { // CASE1: no error => parse normally from the beginning
		return true, true, true, true
	}
	if data.safeSpotOp != "" && data.safeSpotLevel == l { // CASE2: we are the safe spot parser; call lower level for 2. value
		return false, false, true, true // clean up own lData
	}
	// CASE3: data.safeSpotOp != "" && data.safeSpotLevel < l => we should call the next lower level and use its out as value
	return true, true, true, true // we will get safeSpot value
}
func postfixParseCase[Output any](l int, data *recoverData[Output]) (parseVal1, parseSpace, parseOp bool) {
	if data == nil { // CASE1: no error => parse normally from the beginning
		return true, true, true
	}
	if data.safeSpotOp != "" && data.safeSpotLevel == l { // CASE2: we are the safe spot parser; parse only op
		return false, false, true // clean up own lData
	}
	// CASE3: data.safeSpotOp != "" && data.safeSpotLevel < l => we should call the next lower level and use its out as value
	return true, true, true
}

func (e expr[Output]) parseSpace(state comb.State) (comb.State, *comb.ParserError) {
	nState, _, err := e.space.Parse(state)
	if err != nil {
		return state, comb.ClaimError(err)
	}
	return nState, nil
}
