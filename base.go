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
type Parser[Output any] func(input State) Result[Output]

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

// State represents the current state of a parser.
// It consists of the Input, the PointOfNoReturn mark,
// a Failed signal and a collection of error messages.
type State struct {
	input           Input
	PointOfNoReturn uint // mark set by the NoWayBack parser
	Failed          bool
	Messages        []Message
}

// NewInputFromString creates a new input data structure suitable for parsing.
func NewInputFromString(input string) State {
	return State{input: Input{bytes: []byte(input)}}
}

// NewInputFromBytes creates a new input data structure suitable for parsing.
func NewInputFromBytes(input []byte) State {
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
	st.PointOfNoReturn = st.input.pos
	return st
}

// KeepMessages returns a state with all messages from the other state appended to this state.
func (st State) KeepMessages(other State) State {
	st.Messages = append(st.Messages, other.Messages...)

	slices.SortFunc(st.Messages, func(a, b Message) int { // always keep them sorted
		return cmp.Compare(a.Pos, b.Pos)
	})

	return st
}

// Better returns the more advanced (in the input) state of the two.
// This should be used for parsers that are alternatives. So the best error is kept.
func (st State) Better(other State) State {
	if st.input.pos < other.input.pos {
		return other
	}
	return st
}

func (st State) Clean() State {
	st.Messages = make([]Message, 0, 16)
	return st
}

// Result is a generic parser result.
type Result[Output any] struct {
	Output    Output
	Err       *Error
	Remaining State
}

// Success creates a Result with an output set from
// the result of successful parsing.
func Success[Output any](output Output, r State) Result[Output] {
	return Result[Output]{Output: output, Remaining: r}
}

// Failure creates a Result with an error set from
// the result of failed parsing.
func Failure[Output any](err *Error, input State) Result[Output] {
	var output Output
	return Result[Output]{Output: output, Err: err, Remaining: input}
}

// Error represents a parsing error. It holds the input that was being parsed,
// the error that was produced, whether this is a fatal error or there is no way back
// plus what was expected to match.
// If the error is fatal, we have to stop parsing of the file completely.
// If there is no way back we might be able to continue parsing AFTER the error position.
type Error struct {
	Input State
}

// NewError produces a new Error from the provided input and names of
// parsers expected to succeed.
func NewError(st State, messages ...string) *Error {
	ms := make([]Message, len(messages))
	for i, msg := range messages {
		ms[i] = Message{Text: msg, Pos: st.input.pos}
	}
	st.Messages = append(st.Messages, ms...)
	return &Error{Input: st}
}

// Error returns a human readable error string.
func (e *Error) Error() string {
	fullMsg := strings.Builder{}
	for _, message := range e.Input.Messages {
		fullMsg.WriteString(message.Text)
		fullMsg.WriteByte('\n')
	}

	return fullMsg.String()
}
