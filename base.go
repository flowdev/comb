package gomme

import (
	"cmp"
	"slices"
	"strings"
)

// Separator is a generic type for separators (byte, rune, []byte or string)
type Separator interface {
	~rune | ~byte | ~string | ~[]byte
}

// Parser defines the type of a generic Parser function
type Parser[Output any] func(input State) (State, Output)

// Input is the input data for all the parsers.
// It can be either UTF-8 encoded text (a.k.a. string) or raw bytes.
// The parsers store and advance the position within the data but never change the data itself.
// This allows good error reporting including the full line of text containing the error.
type Input struct {
	// Go is fundamentally working with bytes and can interpret them as strings or as containing runes.
	// There are no standard library functions for handling []rune or the like.
	bytes []byte
	pos   uint // position in the sequence a.k.a. the *byte* index
}

// Message is an (error) message from the parser.
// It consists of the Text itself and the position in the input where it happened.
type Message struct {
	Text string
	Pos  uint
}

func ZeroOf[T any]() T {
	var t T
	return t
}

// State represents the current state of a parser.
// It consists of the Input, the pointOfNoReturn mark
// and a collection of error messages.
type State struct {
	input           Input
	pointOfNoReturn uint // mark set by the NoWayBack parser
	Messages        []Message
}

// NewFromString creates a new input data structure suitable for parsing.
func NewFromString(input string) State {
	return State{input: Input{bytes: []byte(input)}}
}

// NewFromBytes creates a new input data structure suitable for parsing.
func NewFromBytes(input []byte) State {
	return State{input: Input{bytes: input}}
}

func (st State) AtEnd() bool {
	return st.input.pos >= uint(len(st.input.bytes))
}

func (st State) BytesRemaining() uint {
	return uint(len(st.input.bytes)) - st.input.pos
}

func (st State) CurrentString() string {
	return string(st.input.bytes[st.input.pos:])
}

func (st State) CurrentBytes() []byte {
	return st.input.bytes[st.input.pos:]
}

func (st State) StringTo(remaining State) string {
	return string(st.BytesTo(remaining))
}

func (st State) BytesTo(remaining State) []byte {
	if remaining.input.pos < st.input.pos {
		return []byte{}
	}
	if remaining.input.pos > uint(len(st.input.bytes)) {
		return st.input.bytes[st.input.pos:]
	}
	return st.input.bytes[st.input.pos:remaining.input.pos]
}

func (st State) MoveBy(countBytes uint) State {
	st.input.pos += countBytes
	ulen := uint(len(st.input.bytes))
	if st.input.pos > ulen { // prevent overrun
		st.input.pos = ulen
	}
	return st
}

func (st State) Moved(other State) bool {
	return st.input.pos != other.input.pos
}

func (st State) ReachedPointOfNoReturn() State {
	st.pointOfNoReturn = st.input.pos
	return st
}

func (st State) NoWayBack() bool {
	return st.pointOfNoReturn > st.input.pos
}

//func (st State) PointOfNoReturn() uint {
//	return st.pointOfNoReturn
//}

func (st State) Clean() State {
	st.Messages = make([]Message, 0, 16)
	return st
}

// Success return the State advanced to the position of
// the subState.
func (st State) Success(subState State) State {
	st.input.pos = subState.input.pos
	return st
}

// Failure return the State with errors kept from
// the subState.
func (st State) Failure(subState State) State {
	st.pointOfNoReturn = max(st.pointOfNoReturn, subState.pointOfNoReturn)

	st.Messages = append(st.Messages, subState.Messages...)
	slices.SortFunc(st.Messages, func(a, b Message) int { // always keep them sorted
		return cmp.Compare(a.Pos, b.Pos)
	})

	return st
}

// Failed returns whether this state is in a failed state or not.
func (st State) Failed() bool {
	return len(st.Messages) > 0
}

// AddError adds the messages to this state at the current position.
func (st State) AddError(messages ...string) State {
	ms := make([]Message, len(messages))
	for i, msg := range messages {
		ms[i] = Message{Text: msg, Pos: st.input.pos}
	}

	return st.Failure(State{Messages: ms})
}

// Error returns a human readable error string.
func (st State) Error() string {
	fullMsg := strings.Builder{}
	for _, message := range st.Messages {
		fullMsg.WriteString(message.Text)
		fullMsg.WriteRune('\n')
	}

	return fullMsg.String()
}

// BetterOf returns the more advanced (in the input) state of the two.
// This should be used for parsers that are alternatives. So the best error is kept.
func BetterOf(state, other State) State {
	if state.input.pos < other.input.pos {
		return other
	}
	return state
}
