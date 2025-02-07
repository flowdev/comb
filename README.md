# A Parser COMBinator Library For Go
![comb logo](logo.png)

Comb is a library that simplifies building parsers in Go.

For me, it has got the optimal feature set:
* Simple maintainability of a normal library thanks to being a parser combinator library.
* Report errors with exact (line and column) position.
* Report **multiple** errors.
* UNICODE support.
* Support for binary input (including byte position and hex dump for errors).
* Type safety (including filling arbitrary typed data) using generics.
* Idiomatic Go code (no generated code, ...).

It's based on [Gomme](https://github.com/oleiade/gomme) that showed how to get the
general developer experience and type safety right.

## Table of content

<!-- TOC -->
* [A Parser COMBinator Library For Go](#a-parser-combinator-library-for-go)
  * [Table of content](#table-of-content)
  * [Getting started](#getting-started)
  * [Examples](#examples)
  * [Documentation](#documentation)
  * [Installation](#installation)
  * [Guide](#guide)
    * [List of combinators](#list-of-combinators)
      * [Base combinators](#base-combinators)
      * [Bytes combinators](#bytes-combinators)
      * [Character combinators](#character-combinators)
      * [Combinators for Sequences](#combinators-for-sequences)
      * [Combinators for Applying Parsers Many Times](#combinators-for-applying-parsers-many-times)
      * [Combinators for Choices](#combinators-for-choices)
  * [Frequently asked questions](#frequently-asked-questions)
    * [Q: What's the name?](#q-whats-the-name)
    * [Q: What are parser combinators?](#q-what-are-parser-combinators)
    * [Q: Why would I use parser combinators instead of a specific parser?](#q-why-would-i-use-parser-combinators-instead-of-a-specific-parser)
    * [Q: Where can I learn more about parser combinators?](#q-where-can-i-learn-more-about-parser-combinators)
  * [Acknowledgements](#acknowledgements)
  * [Authors](#authors)
<!-- TOC -->


## Getting started

Here's how to quickly parse [hexadecimal color codes](https://developer.mozilla.org/en-US/docs/Web/CSS/color) using Gomme:

```golang
// RGBColor stores the three bytes describing a color in the RGB space.
type RGBColor struct {
    red   uint8
    green uint8
    blue  uint8
}

// ParseRGBColor creates a new RGBColor from a hexadecimal color string.
// The string must be a six-digit hexadecimal number, prefixed with a "#".
func ParseRGBColor(input string) (RGBColor, error) {
    parse := cmb.Map4(
        SaveSpot(C('#')),
        HexColorComponent("red hex color"),
        HexColorComponent("green hex color"),
        HexColorComponent("blue hex color"),
        func(_ rune, r, g, b string) (RGBColor, error) {
            return RGBColor{fromHex(r), fromHex(g), fromHex(b)}, nil
        },
    )

    return comb.RunOnString(input, parse)
}

// HexColorComponent produces a parser that parses a single hex color component,
// which is a two-digit hexadecimal number.
func HexColorComponent() gomme.Parser[string] {
    return SaveSpot(cmb.SatisfyMN(expected, 2, 2, cmb.IsHexDigit))
}

// fromHex converts a two digits hexadecimal number to its decimal value.
func fromHex(input string) uint8 {
    res, _ := strconv.ParseUint(input, 16, 8) // errors have been caught by the parser
    return uint8(res)
}
```

It's as simple as that! Feel free to explore more in the [examples](./examples) directory.

## Examples

See Comb in action with these handy examples:
- [Parsing hexadecimal color codes](./examples/hexcolor)
- [Parsing a simple CSV file](./examples/csv)
- [Parsing Redis' RESP protocol](./examples/redis)
- [Parsing JSON](./examples/json)

## Documentation

For more detailed information, refer to the official [documentation](https://pkg.go.dev/github.com/flowdev/comb).

## Installation

Like any other library:
```bash
go get github.com/flowdev/comb
```

## Guide

In this guide, we provide a detailed overview of the various combinators available in Comb.
Combinators are fundamental building blocks in parser construction,
each designed for a specific task.
By combining them, you can create complex parsers suited to your specific needs.
For each combinator, we've provided a brief description and a usage example.
Let's explore!

### List of combinators

#### Base combinators

| Combinator                                                           | Description                                                                                                                                                                                                             | Example                                                            |
| :------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------- |
| [`Map`](https://pkg.go.dev/github.com/oleiade/gomme#Map)             | Applies a function to the result of the provided parser, allowing you to transform the parser's result. | `Map(Digit1(), func(s string)int { return 123 })`                  |
| [`Optional`](https://pkg.go.dev/github.com/oleiade/gomme#Optional)   | Makes a parser optional. If unsuccessful, the parser returns a nil `Result.Output`.Output`.                                                                                                                         | `Optional(CRLF())`                                                 |
| [`Peek`](https://pkg.go.dev/github.com/oleiade/gomme#Peek)           | Applies the provided parser without consuming the input.                                                                                                                                                               |                                                                    |
| [`Recognize`](https://pkg.go.dev/github.com/oleiade/gomme#Recognize) | Returns the consumed input as the produced value when the provided parser is successful.                                                                                                                              | `Recognize(SeparatedPair(Token("key"), Char(':'), Token("value"))` |
| [`Assign`](https://pkg.go.dev/github.com/oleiade/gomme#Assign)       | Returns the assigned value when the provided parser is successful.                                                                                                                                                   | `Assign(true, Token("true"))`                                      |

#### Bytes combinators

| Combinator                                                               | Description                                                                                                                                                                                                        | Example                               |
| :----------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :------------------------------------ |
| [`Take`](https://pkg.go.dev/github.com/oleiade/gomme#Take)               | Parses the first N elements of the input.                                                                                                                                                                               | `Take(5)`                             |
| [`TakeUntil`](https://pkg.go.dev/github.com/oleiade/gomme#TakeUntil)     | Parses the input until the provided parser argument succeeds.                                                                                                                                                     | `TakeUntil(CRLF()))`                  |
| [`TakeWhileMN`](https://pkg.go.dev/github.com/oleiade/gomme#TakeWhileMN) | Parses the longest input slice fitting the length expectation (m <= input length <= n) and matching the predicate. The parser argument is a function taking a `rune` as input and returning a `bool`. | `TakeWhileMN(2, 6, gomme.isHexDigit)` |
| [`Token`](https://pkg.go.dev/github.com/oleiade/gomme#Token)             | Recognizes a specific pattern. Compares the input with the token's argument and returns the matching part.                                                                                                   | `Token("tolkien")`                    |

#### Character combinators

| Combinator | Description | Example |
| :--- | :--- | :--- |
| [`Char`](https://pkg.go.dev/github.com/oleiade/gomme#Char) | Parses a single instance of a provided character. | `Char('$')` |
| [`AnyChar`](https://pkg.go.dev/github.com/oleiade/gomme#AnyChar) | Parses a single instance of any character. | `AnyChar()` |
| [`Alpha0`](https://pkg.go.dev/github.com/oleiade/gomme#Alpha0) | Parses zero or more alphabetical ASCII characters (case insensitive). | `Alpha0()` |
| [`Alpha1`](https://pkg.go.dev/github.com/oleiade/gomme#Alpha1) | Parses one or more alphabetical ASCII characters (case insensitive). | `Alpha1()` |
| [`Alphanumeric0`](https://pkg.go.dev/github.com/oleiade/gomme#Alphanumeric0) | Parses zero or more alphabetical and numerical ASCII characters (case insensitive). | `Alphanumeric0()` |
| [`Alphanumeric1`](https://pkg.go.dev/github.com/oleiade/gomme#Alphanumeric1) | Parses one or more alphabetical and numerical ASCII characters (case insensitive). | `Alphanumeric1()` |
| [`Digit0`](https://pkg.go.dev/github.com/oleiade/gomme#Digit0) | Parses zero or more numerical ASCII characters: 0-9. | `Digit0()` |
| [`Digit1`](https://pkg.go.dev/github.com/oleiade/gomme#Digit1) | Parses one or more numerical ASCII characters: 0-9. | `Digit1()` |
| [`HexDigit0`](https://pkg.go.dev/github.com/oleiade/gomme#HexDigit0) | Parses zero or more hexadecimal ASCII characters (case insensitive). | `HexDigit0()` |
| [`HexDigit1`](https://pkg.go.dev/github.com/oleiade/gomme#HexDigit1) | Parses one or more hexadecimal ASCII characters (case insensitive). | `HexDigit1()` |
| [`Whitespace0`](https://pkg.go.dev/github.com/oleiade/gomme#Whitespace0) | Parses zero or more whitespace ASCII characters: space, tab, carriage return, line feed. | `Whitespace0()` |
| [`Whitespace1`](https://pkg.go.dev/github.com/oleiade/gomme#Whitespace1) | Parses one or more whitespace ASCII characters: space, tab, carriage return, line feed. | `Whitespace1()` |
| [`LF`](https://pkg.go.dev/github.com/oleiade/gomme#LF) | Parses a single new line character '\n'. | `LF()` |
| [`CRLF`](https://pkg.go.dev/github.com/oleiade/gomme#CRLF) | Parses a '\r\n' string. | `CRLF()` |
| [`OneOf`](https://pkg.go.dev/github.com/oleiade/gomme#OneOf) | Parses one of the provided characters. Equivalent to using `Alternative` over a series of `Char` parsers. | `OneOf('a', 'b' , 'c')` |
| [`Satisfy`](https://pkg.go.dev/github.com/oleiade/gomme#Satisfy) | Parses a single character, asserting that it matches the provided predicate. The predicate function takes a `rune` as input and returns a `bool`. `Satisfy` is useful for building custom character matchers. | `Satisfy(func(c rune)bool { return c == '{' || c == '[' })` |
| [`Space`](https://pkg.go.dev/github.com/oleiade/gomme#Space) | Parses a single space character ' '. | `Space()` |
| [`Tab`](https://pkg.go.dev/github.com/oleiade/gomme#Tab) | Parses a single tab character '\t'. | `Tab()` |
| [`Int64`](https://pkg.go.dev/github.com/oleiade/gomme#Int64) | Parses an `int64` from its textual representation. | `Int64()` |
| [`Int8`](https://pkg.go.dev/github.com/oleiade/gomme#Int8) | Parses an `int8` from its textual representation. | `Int8()` |
| [`UInt8`](https://pkg.go.dev/github.com/oleiade/gomme#UInt8) | Parses a `uint8` from its textual representation. | `UInt8()` |

#### Combinators for Sequences

| Combinator | Description | Example |
| :--- | :--- | :--- |
| [`Preceded`](https://pkg.go.dev/github.com/oleiade/gomme#Preceded) | Applies the prefix parser and discards its result. It then applies the main parser and returns its result. It discards the prefix value. It proves useful when looking for data prefixed with a pattern. For instance, when parsing a value, prefixed with its name. | `Preceded(Token("name:"), Alpha1())` |
| [`Terminated`](https://pkg.go.dev/github.com/oleiade/gomme#Terminated) | Applies the main parser, followed by the suffix parser whom it discards the result of, and returns the result of the main parser. Note that if the suffix parser fails, the whole operation fails, regardless of the result of the main parser. It proves useful when looking for suffixed data while not interested in retaining the suffix value itself. For instance, when parsing a value followed by a control character. | `Terminated(Digit1(), LF())` |
| [`Delimited`](https://pkg.go.dev/github.com/oleiade/gomme#Delimited) | Applies the prefix parser, the main parser, followed by the suffix parser, discards the result of both the prefix and suffix parsers, and returns the result of the main parser. Note that if any of the prefix or suffix parsers fail, the whole operation fails, regardless of the result of the main parser. It proves useful when looking for data surrounded by patterns helping them identify it without retaining its value. For instance, when parsing a value, prefixed by its name and followed by a control character. | `Delimited(Tag("name:"), Digit1(), LF())` |
| [`Pair`](https://pkg.go.dev/github.com/oleiade/gomme#Pair) | Applies two parsers in a row and returns a pair container holding both their result values. | `Pair(Alpha1(), Tag("cm"))` |
| [`SeparatedPair`](https://pkg.go.dev/github.com/oleiade/gomme#SeparatedPair) | Applies a left parser, a separator parser, and a right parser discards the result of the separator parser, and returns the result of the left and right parsers as a pair container holding the result values. | `SeparatedPair(Alpha1(), Tag(":"), Alpha1())` |
| [`Sequence`](https://pkg.go.dev/github.com/oleiade/gomme#Sequence) | Applies a sequence of parsers sharing the same signature. If any of the provided parsers fail, the whole operation fails. | `Sequence(SeparatedPair(Tag("name"), Char(':'), Alpha1()), SeparatedPair(Tag("height"), Char(':'), Digit1()))` |

#### Combinators for Applying Parsers Many Times

| Combinator | Description | Example |
| :--- | :--- | :--- |
| [`Count`](https://pkg.go.dev/github.com/oleiade/gomme#Count) | Applies the provided parser `count` times. If the parser fails before it can be applied `count` times, the operation fails. It proves useful whenever one needs to parse the same pattern many times in a row. | `Count(3, OneOf('a', 'b', 'c'))` |
| [`Many0`](https://pkg.go.dev/github.com/oleiade/gomme#Many0) | Keeps applying the provided parser until it fails and returns a slice of all the results. Specifically, if the parser fails to match, `Many0` still succeeds, returning an empty slice of results. It proves useful when trying to consume a repeated pattern, regardless of whether there's any match, like when trying to parse any number of whitespaces in a row. | `Many0(Char(' '))` |
| [`Many1`](https://pkg.go.dev/github.com/oleiade/gomme#Many1) | Keeps applying the provided parser until it fails and returns a slice of all the results. If the parser fails to match at least once, `Many1` fails. It proves useful when trying to consume a repeated pattern, like any number of whitespaces in a row, ensuring that it appears at least once. | `Many1(LF())` |
| [`SeparatedList0`](https://pkg.go.dev/github.com/oleiade/gomme#SeparatedList0) |  |  |
| [`SeparatedList1`](https://pkg.go.dev/github.com/oleiade/gomme#SeparatedList1) |  |  |

#### Combinators for Choices

| Combinator | Description | Example |
| :--- | :--- | :--- |
| [`Alternative`](https://pkg.go.dev/github.com/oleiade/gomme#Alternative) | Tests a list of parsers, one by one, until one succeeds. Note that all parsers must share the same signature (`Parser[I, O]`). | `Alternative(Token("abc"), Token("123"))` |


## Frequently asked questions

### Q: What's the name?

**A**: **Comb** first of all got its name from being a parser COMBinatior library.
But it is also very good at "combing" through input, finding errors.
Since the error handling system is the by far hardest part of the project the name feels right.

### Q: What are parser combinators?

**A**: Parser combinators offer a new way of building parsers.
Instead of writing a complex parser that analyzes an entire format,
you create small, simple parsers that handle the smallest units of the format.
These small parsers can then be combined to build more complex parsers.
It's a bit like using building blocks to construct whatever structure you want.

### Q: Why would I use parser combinators instead of a specific parser?

**A**: Parser combinators are incredibly flexible and intuitive.
Once you're familiar with them, they enable you to quickly create, maintain, and modify parsers.
They offer you a high degree of freedom in designing your parser and how it's used.

### Q: Where can I learn more about parser combinators?

A: Here are some resources we recommend:
- [You could have invented parser combinators](https://theorangeduck.com/page/you-could-have-invented-parser-combinators)
- [Functional Parsing](https://www.youtube.com/watch?v=dDtZLm7HIJs)
- [Building a Mapping Language in Go with Parser Combinators](https://www.youtube.com/watch?v=JiViND-bpmw)

## Acknowledgements

We've stood on the shoulders of giants to create Comb.
The library draws heavily on the extensive theoretical work done in the parser combinators space,
and we owe a huge thanks to [Gomme](https://github.com/oleiade/gomme) that
solved two important problems for us and gave hints for others.

Getting the error recovery mechanism right took at least 90% of the time
put into the project until now.

## Authors

- [@ole108](https://github.com/ole108) (main developer behind the flowdev organization)
- [@oleiade](https://github.com/oleiade) (for Gomme)
