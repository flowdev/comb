// Package gomme implements a parser combinator library.
// It provides a toolkit for developers to build reliable, fast, flexible, and easy-to-develop and maintain parsers
// for both textual and binary formats. It extensively uses the recent introduction of Generics in the Go programming
// language to offer flexibility in how combinators can be mixed and matched to produce the desired output while
// providing as much compile-time type safety as possible.
package gomme

// NoWayBack applies a child parser and marks a successful result with NoWayBack.
func NoWayBack[Output any](parse Parser[Output]) Parser[Output] {
	return func(state State) (State, Output) {
		newState, output := parse(state)
		if newState.Failed() {
			return newState, output
		}

		return newState.ReachedPointOfNoReturn(), output
	}
}

// Optional applies an optional child parser. Will return nil
// if not successful.
//
// N.B: Optional will ignore any parsing failures and errors.
func Optional[Output any](parse Parser[Output]) Parser[Output] {
	return func(input State) (State, Output) {
		newState, output := parse(input)
		if newState.NoWayBack() {
			return newState, output
		}
		if newState.Failed() {
			return input, ZeroOf[Output]()
		}
		return newState, output
	}
}

// Peek tries to apply the provided parser without consuming any input.
// It effectively allows to look ahead in the input.
func Peek[Output any](parse Parser[Output]) Parser[Output] {
	return func(state State) (State, Output) {
		newState, output := parse(state)
		if newState.Failed() {
			return state.Failure(newState), output
		}

		return state, output
	}
}

// Recognize returns the consumed input (instead of the original parsers output)
// as the produced value when the provided parser succeeds.
func Recognize[Output any](parse Parser[Output]) Parser[[]byte] {
	return func(state State) (State, []byte) {
		newState, _ := parse(state)
		if newState.Failed() {
			return newState, []byte{}
		}

		return newState, state.BytesTo(newState)
	}
}

// Assign returns the provided value if the parser succeeds, otherwise
// it returns an error result.
func Assign[Output1, Output2 any](value Output1, parse Parser[Output2]) Parser[Output1] {
	return func(input State) (State, Output1) {
		newState, _ := parse(input)
		if newState.Failed() {
			return newState, ZeroOf[Output1]()
		}

		return newState, value
	}
}

// Map1 applies a function to the successful result of 1 parser.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map1[PO1 any, MO any](parse Parser[PO1], fn func(PO1) (MO, error)) Parser[MO] {
	return func(state State) (State, MO) {
		newState, output := parse(state)
		if newState.Failed() {
			return state.Failure(newState), ZeroOf[MO]()
		}

		mapped, err := fn(output)
		if err != nil {
			return state.AddError(err.Error()), ZeroOf[MO]()
		}

		return newState, mapped
	}
}

// Map2 applies a function to the successful result of 2 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map2[PO1, PO2 any, MO any](parse1 Parser[PO1], parse2 Parser[PO2], fn func(PO1, PO2) (MO, error)) Parser[MO] {
	return func(state State) (State, MO) {
		newState1, output1 := parse1(state)
		if newState1.Failed() {
			return state.Failure(newState1), ZeroOf[MO]()
		}

		newState2, output2 := parse2(newState1)
		if newState2.Failed() {
			return state.Failure(newState2), ZeroOf[MO]()
		}

		mapped, err := fn(output1, output2)
		if err != nil {
			return state.AddError(err.Error()), ZeroOf[MO]()
		}

		return newState2, mapped
	}
}

// Map3 applies a function to the successful result of 3 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map3[PO1, PO2, PO3 any, MO any](parse1 Parser[PO1], parse2 Parser[PO2], parse3 Parser[PO3],
	fn func(PO1, PO2, PO3) (MO, error),
) Parser[MO] {
	return func(state State) (State, MO) {
		newState1, output1 := parse1(state)
		if newState1.Failed() {
			return state.Failure(newState1), ZeroOf[MO]()
		}

		newState2, output2 := parse2(newState1)
		if newState2.Failed() {
			return state.Failure(newState2), ZeroOf[MO]()
		}

		newState3, output3 := parse3(newState2)
		if newState3.Failed() {
			return state.Failure(newState3), ZeroOf[MO]()
		}

		mapped, err := fn(output1, output2, output3)
		if err != nil {
			return state.AddError(err.Error()), ZeroOf[MO]()
		}

		return newState3, mapped
	}
}

// Map4 applies a function to the successful result of 4 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map4[PO1, PO2, PO3, PO4 any, MO any](parse1 Parser[PO1], parse2 Parser[PO2], parse3 Parser[PO3], parse4 Parser[PO4],
	fn func(PO1, PO2, PO3, PO4) (MO, error),
) Parser[MO] {
	return func(state State) (State, MO) {
		newState1, output1 := parse1(state)
		if newState1.Failed() {
			return state.Failure(newState1), ZeroOf[MO]()
		}

		newState2, output2 := parse2(newState1)
		if newState2.Failed() {
			return state.Failure(newState2), ZeroOf[MO]()
		}

		newState3, output3 := parse3(newState2)
		if newState3.Failed() {
			return state.Failure(newState3), ZeroOf[MO]()
		}

		newState4, output4 := parse4(newState3)
		if newState4.Failed() {
			return state.Failure(newState4), ZeroOf[MO]()
		}

		mapped, err := fn(output1, output2, output3, output4)
		if err != nil {
			return state.AddError(err.Error()), ZeroOf[MO]()
		}

		return newState4, mapped
	}
}

// Map5 applies a function to the successful result of 5 parsers.
// Arbitrary complex data structures can be built with Map1 and Map2 alone.
// The other Map* parsers are provided for convenience.
func Map5[PO1, PO2, PO3, PO4, PO5 any, MO any](
	parse1 Parser[PO1], parse2 Parser[PO2], parse3 Parser[PO3], parse4 Parser[PO4], parse5 Parser[PO5],
	fn func(PO1, PO2, PO3, PO4, PO5) (MO, error),
) Parser[MO] {
	return func(state State) (State, MO) {
		newState1, output1 := parse1(state)
		if newState1.Failed() {
			return state.Failure(newState1), ZeroOf[MO]()
		}

		newState2, output2 := parse2(newState1)
		if newState2.Failed() {
			return state.Failure(newState2), ZeroOf[MO]()
		}

		newState3, output3 := parse3(newState2)
		if newState3.Failed() {
			return state.Failure(newState3), ZeroOf[MO]()
		}

		newState4, output4 := parse4(newState3)
		if newState4.Failed() {
			return state.Failure(newState4), ZeroOf[MO]()
		}

		newState5, output5 := parse5(newState4)
		if newState5.Failed() {
			return state.Failure(newState5), ZeroOf[MO]()
		}

		mapped, err := fn(output1, output2, output3, output4, output5)
		if err != nil {
			return state.AddError(err.Error()), ZeroOf[MO]()
		}

		return newState5, mapped
	}
}
