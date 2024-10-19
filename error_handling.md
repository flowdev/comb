# Error Handling

The handling of (syntax) errors is the by far hardest part of this project.
I had to refactor the project **three** times to get it right and
almost made a PhD in computer science understanding all those
scientific papers about error handling in parsers
with their extremely concise notation that is explained nowhere
because it is the well known standard in the field.
Thank you, Sérgio Medeiros and Fabio Mascarenhas, for your paper
[Syntax Error Recovery in Parsing Expression Grammars](https://dl.acm.org/doi/10.1145/3167132.3167261).
That brought me on the right track.
And thank you, Terence Parr and [ANTLR](https://www.antlr.org/),
for an OpenSource parser to compare against.
I would have switched to it if I had found the Go support of ANTLR early enough.

So please take some time to understand this before making or suggesting
any major changes.

The error handling consists of error reporting and recovering from errors.

## Error Reporting

Syntax errors are always reported in the form:
> expected "token" [line:column] source line incl. marker ▶ at error position

Programming errors (in one of Your parsers) are always reported in the form:
> programming error: message [line:column] source line incl. marker ▶ at error position

Semantic and miscellaneous errors are always reported in the form:
> message [line:column] source line incl. marker ▶ at error position

Calculating the correct line and column of the error and setting the marker
correctly are the hardest problems here.
And they bring the most benefit to the user.

## Recovering From Errors

In general, we distinguish between simple **leaf** parsers that don't use
any sub-parsers and **branch** parsers that do use one or more sub-parsers.

For recovering from errors the parser uses a minimal set of modes. \
Great care has to be taken by all branch parsers because we not only
have to find a safe point to recover to, but also have to have the correct
Go call stack to be able to parse correctly after recovering.
Furthermore, we need to use the parsers more often than other (e.g. LR-) parsers. \
This is a downside of all parser combinators.
We mitigate it with some helpers and by caching as much as reasonable.

The `NoWayBack` parser plays a key role in error recovery.
It is the one to conclude that an error has indeed to be handled
(if its position is before the error),
and it also marks the next safe state to which we want to recover to
(if its position is behind the error). \
A `NoWayBack` parser at the exact error position isn't of help
for that particular error. \
Finally, the `NoWayBack` parser is used to prevent the `FirstSuccessful` parser
from trying other sub-parsers even in case of an error.
This way we prevent unnecessary backtracking.

So please use the `NoWayBack` parser as much as reasonable for your grammar!
As it keeps the backtracking to a minimum, it also makes the parser perform better.

The `FirstSuccessful` and `NoWayBack` parsers are special **branch** parsers.

The following sections define the modes and their relationships in detail.

### Parser Modes

These are the modes:

##### happy:
Normal parsing discovering and reporting errors
(with `State.NewError` or `State.ErrorAgain` for cached results). \
The error will be witnessed by the immediate parent branch parser.

If we happen to handle an error and hit a `NoWayBack` parser then
we will be very happy to clean up. \
This means we were able to handle the error by modifying the input
and didn't have to use any `Resolverer`.

##### error:
An error was found but might be mitigated by backtracking and the
`FirstSuccessful` parser.
In this mode the parser goes back to find the last `NoWayBack` parser or
trying later alternatives in the `FirstSuccessful` parser.

The previous `NoWayBack` parser might be hidden deep in a sub-parser
that is earlier in sequence but not on the Go call stack anymore.

So in this mode all parsers that use sub-parsers in sequence have to use them
in reverse order to find the right `NoWayBack` parser. \
Funnily this also applies to parsers that use the *same* sub-parser
multiple times. So if the second time the sub-parser was used, failed
then it might very well be that the first (successful) time it applied
a `NoWayBack` parser. And that would be the right one to find.

Only the `FirstSuccessful` parser (not as parent parser but as sibling this time)
is different. It has to find the first successful sub-parser and
its `NoWayBack` parser again. \
As parent parser (if the **error** mode is switched to while trying alternatives)
it can just try another alternative (normal **happy** mode behaviour).

##### handle:
We now know that the error found has to be handled.
We find the exact position and witness parser again by simply parsing
one more time (forward) in the new mode (possibly omitting any semantics).

The witness parser should
1. modify the input (respecting `maxDel`),
2. switch to **happy** mode and
3. parse again (possibly omitting the failing parser).
4. switch to **escape** mode if everything else fails.

##### rewind:
We failed again and have to try again with more deletion or
without using the parser that failed originally.

So we have to go backward similar to the **error** mode.
But with the distinction that we aren't looking for a `NoWayBack` parser
before the error position, but instead for the immediate parent branch parser of
the failing leaf parser that witnessed the error.

##### escape:
All deletion of input and inserting of good input didn't help.
Now we are out of options and can just escape this using a `Resolverer`.

So we find the best (least waste) `Resolverer` and its `NoWayBack` parser
executes it and finally cleans up and switches back to **happy** mode. \
The best `Resolverer` to use can't be determined statically,
because it depends on the input.

### Parsing Directions Per Mode

The direction of parsing changes with the mode.
Normal parsing is forward of course but in some other modes we have to move backward.
Here is the full table:

|   Mode | Direction                                      |
|-------:|:-----------------------------------------------|
|  happy | forward (until a failure is witnessed)         |
|  error | **backward** (to the **previous** `NoWayBack`) |
| handle | forward (to the `witness parser (1)`)          |
| rewind | **backward** (to the `witness parser (1)`)     |
| escape | forward (to the **next** `NoWayBack`)          |

So the parsers move only in the **error** and **rewind** modes backward,
and forward in all other modes.

### Relationships Between Modes

The relationships between the modes are shown in the following
state diagram.
The diagram also shows where a mode change can happen and the condition
(next to the mode) that has to be fulfilled for the change.

The position of the error is shortened to `errPos`. \
The first parent branch parser to witness the error to be handled
is called 'witness parser (1)'. \
A possibly different parent branch parser to witness an error
during handling of the first is called 'witness parser (2)'.

```mermaid
---
title: Parser Modes And Their Changes
---
stateDiagram-v2
    [*] --> happy: start

    happy --> error: State.NewError + witness parser (1) (no error yet)
    error --> happy: FirstSuccessful (successful parser found)
    error --> handle: NoWayBack (pos < errPos)
    handle --> happy: witness parser (1)
    happy --> rewind: State.NewError + witness parser (2) (error exists)
    rewind --> happy: witness parser (1)
    happy --> happy: NoWayBack (pos > errPos) clean up
    rewind --> escape: witness parser (1)
    escape --> happy: NoWayBack (pos > errPos) clean up
```

Next we will look at the changes in modes that are possible within sub-parsers.

### Possible Mode Changes

The following table lists the mode changes that are possible in a leaf or
branch sub-parser. \
The `NoWayBack` parser and the `witness parser`s can only have leaf parsers
as sub-parser. Every other branch parser can also have branch parsers as
sub-parsers.

| Mode At Entry | Possible Modes After Leaf Parser | Possible Modes After Branch Parser |
|--------------:|:---------------------------------|:-----------------------------------|
|         happy | happy, error                     | happy, error, escape               |
|         error | error                            | error, handle                      |
|        handle | handle                           | handle, happy, escape              |
|        rewind | rewind                           | rewind, happy, escape              |
|        escape | escape                           | escape, happy, escape              |

The following sections detail some error recovery scenarios.

### Example Scenarios Or Error Recovery

Before we can dive into the scenarios themselves we have to define
a few abbreviations (or the diagrams would go beyond the screen).

- `Px`: any parser with no special role (`x` being a decimal number), e.g.: `P7`
- `NWB`: a `NoWayBack` parser wrapping any leaf parser
- `NWB3`: up to three `NoWayBack` parsers might be involved in a complex scenario
- `WP1`: `witness parser (1)` witnessing any sub-parser; it's the error handling parser
- `WP2`: `witness parser (2)` witnessing any sub-parser; it just witnesses a secondary error
- `FS`: a `FirstSuccessful` parser
- `FSx`: the `FirstSuccessful` parser number `x` (`x` being a decimal number), e.g.: `FS3`
- `parser(mode)`: the parser is in a certain mode, e.g.: `WP1(handle)`
- `parser(mode1, mode2)`: multiple possible modes are separated by a comma (','),
  e.g.: `WP2(error, rewind)`
- `NWB(parser)`: the `NoWayBack` parser wraps a parser, e.g: `NWB(other parser)`

A complex example is: `NWB(WP1(happy, handle))` meaning a `NoWayBack` parser
that also acts as `witness parser (1)` that is either in mode `happy` or
in mode `handle`.

We will use flow diagrams for the scenarios and the links between the nodes
show the order and potentially modes in parentheses.

#### Simple Sequence

The simple sequence scenario looks like this if nothing fails:

```mermaid
flowchart LR
  a([start])--->|"(happy)"|NWB1--->|"1 (happy)"|WP1--->|"2 (happy)"|NWB2--->|"3 (happy)"|b([end])
```

If `WP1` fails it will look like this:

```mermaid
flowchart LR
  st(["start"])
  p1["NWB1"]
  p2["WP1"]
  p3["NWB2"]
  ed(["end"])
  
  st--->|"(happy)"| p1
  p1--->|"1 (happy)"| p2
  p2--->|"2 (error)"| p1
  p1--->|"3 (handle)"| p2
  p2--->|"4 (rewind)"| p2
  p2--->|"5 (happy, escape)"| p3
  p3--->|"6 (happy)"| ed
```
The last step can be in mode `happy` if the error could be resolved by deletion or insertion.
It will be in mode `escape` if we have to use the `Resolverer` of `NWB2`.

#### Simple Sequence With Three `NoWayBack`s

A slight complication of the sequence above is the following (without failure):

```mermaid
flowchart LR
  st(["start"])
  p1["NWB1"]
  p2["NWB2(WP1)"]
  p3["WP2"]
  p4["NWB3"]
  ed(["end"])

  st--->|"(happy)"|p1--->|"1 (happy)"|p2--->|"2 (happy)"|p3--->|"3 (happy)"|p4
  p4--->|"4 (happy)"| ed
```

If `WP1` fails it will look like this:

```mermaid
flowchart LR
  st(["start"])
  p1["NWB1"]
  p2["NWB2(WP1)"]
  p3["WP2"]
  p4["NWB3"]
  ed(["end"])
  
  st--->|"(happy)"|p1--->|"1 (happy)"|p2--->|"6 (happy, escape)"|p3--->|"7 (happy, escape)"|p4
  p2--->|"2 (error)"|p1
  p1--->|"3 (handle)"|p2
  p2--->|"4 (happy)"|p3
  p3--->|"5 (rewind)"|p2
  p4--->|"4 (happy)"|ed
```
The last two steps can be in mode `happy` if the error could be resolved by deletion or insertion.
It will be in mode `escape` if we have to use the `Resolverer` of `NWB2`.

So this scenario is **really** the same as the most simple one above.
It mainly illustrates that the `NoWayBack` parser 2 (`NWB2`) isn't of any help
but only serves as `witness parser (1)` in this scenario.
And the error recovery is (sometimes) failing at the witness parser (2) (`WP2`).

#### Cascading Sequences

In this scenario all parts involved are distributed over different
sequence like parsers (parsers base on `Sequence`, `MapN` or `MultiMN`).

```mermaid
flowchart LR
  st(["start"])
  subgraph main["Main Sequence"]
    direction LR
    subgraph sub1["Subsequence 1"]
      direction TB
      p11["P4"]
      p12["NWB1"]
      p13["P5"]
      p11--->p12--->p13
    end
    subgraph sub2["Subsequence 2"]
      direction TB
      p21["P6"]
      p22["WP1"]
      p23["P7"]
      p21--->p22--->p23
    end
    sub1--->sub2
    subgraph sub3["Subsequence 3"]
      direction TB
      p31["P8"]
      p32["WP2"]
      p33["P9"]
      p31--->p32--->p33
    end
    sub2--->sub3
    subgraph sub4["Subsequence 4"]
      direction TB
      p41["P10"]
      p42["NWB2"]
      p43["P11"]
      p41--->p42--->p43
    end
    sub3--->sub4
  end
  ed(["end"])
  st--->main
  main--->ed
```
The important thing to note about this much more complex scenario is that all
the `Px` parsers play no active role in the game.
They don't change the mode or perform any kind of error handling.

They only pass on the state in the right direction (according to the parsing mode)
and possibly advance the position in the input.
The position in the input is **the** crucial thing here.
If that isn't handled perfectly, all caching will miss, and the parser
is **broken**.

#### Cascading Sequences With `FirstSuccessful` Parser

In this scenario all parts involved are inside different `FirstSuccessful` parsers.
The alternative sub-parsers of the `FirstSuccessful` parser are connected by
dotted lines and the alternatives are numbered in parentheses in the order
they are tried.

```mermaid
flowchart LR
  st(["start"])
  subgraph main["Main Sequence"]
    direction LR
    subgraph fs1["FirstSuccessful 1"]
      direction BT
      p1["(1) P1"]
      subgraph sub1["(2) Subsequence 1"]
        direction BT
        p11["P2"]
        p12["NWB1"]
        p13["P3"]
        p11--->p12--->p13
      end
      p2["(3) P4"]
      p1-.-sub1-.-p2
    end
    subgraph fs2["FirstSuccessful 2"]
      direction BT
      p3["(1) P5"]
      subgraph sub2["(2) Subsequence 2"]
        direction BT
        p21["P6"]
        p22["WP1"]
        p23["P7"]
        p21--->p22--->p23
      end
      p4["(3) P8"]
      p3-.-sub2-.-p4
    end
    fs1--->fs2
    subgraph fs3["FirstSuccessful 3"]
      direction BT
      p5["(1) P9"]
      subgraph sub3["(2) Subsequence 3"]
        direction BT
        p31["P10"]
        p32["WP2"]
        p33["P11"]
        p31--->p32--->p33
      end
      p6["(3) P12"]
      p5-.-sub3-.-p6
    end
    fs2--->fs3
    subgraph fs4["FirstSuccessful 4"]
      direction BT
      p7["(1) P13"]
      subgraph sub4["(2) Subsequence 4"]
        direction BT
        p41["P14"]
        p42["NWB2"]
        p43["P15"]
        p41--->p42--->p43
      end
      p8["(3) P16"]
      p7-.-sub4-.-p8
    end
    fs3--->fs4
  end
  ed(["end"])
  st--->main
  main--->ed
```

The important thing to note here is that the `FirstSuccessful` parser should
evaluate and choose between alternatives only in **happy** mode.

In modes **error**, **handle** and **rewind** it has to use the exact same
alternative as before. \
And in **escape** mode it has to choose between potentially multiple
`NoWayBack` parsers and their `Recoverer`s in a different way.

The following sections document the details what the parsers or
methods mentioned above should do in each mode.

### Method `State.NewError`

##### happy:
Create new error and switch to `mode=error`.

##### error:
Ignore. \
This can happen in an alternative branch that wasn't used because of
a better error further in the input in a later branch. \
With proper caching in place we can turn this into a programming error.

##### handle:
If `newError==error` then switch to `mode=record` \
else register programming error. \
We must have missed either the erroring parser in `mode==happy` or
the error to handle just now in `mode==handle`.

##### record:
Ignore call (should not happen).
This would just cost a bit of performance and it thus no
programming error to be fixed.

##### collect:
Like mode **_record_**.

##### choose:
Create new error.

##### play:
Like mode **_choose_**.

### Parser `NoWayBack`

##### happy:
Set the `noWayBackMark` in the State if the sub-parser has been successful.
Else just return the error.

##### error:
Switch to `mode=handle` and return.

##### handle:
Call sub-parser to find the error again.
If the mode hasn't changed to `record` after the call,
report a programming error because
this parser should have been the one switching to `mode=handle` or
we missed the error.

##### record:
If `mode==record` at the start then this is the safe place wanted to recover to.
Now recover. (Switch to `mode=play`.
Use `Deleter` to delete 1 to `maxDel` tokens and call all recorded parsers
in their order.
If no success just do the same without calling the very first
recorded parser (this simulates inserting correct input).
If still no success use the `Recoverer` to find the next safe spot in the input.
Move there, switch to `mode=happy` and resume normal parsing by calling the sub-parser.
This has to be successful (or we record a programming error).
Finally advance the `noWayBackMark` accordingly (like in `mode==happy`). )
Else just return (the sub-parser has already recorded itself).

##### collect:
Signal back to the `FirstSuccessful` parser that a `NoWayBack` parser was found
(including the `waste`?).

##### choose:
Signal the `waste` of the `Recoverer` back to the `FirstSuccessful` parser.

##### play:
If `mode==play` at the start then call sub-parser.
(If that is successful switch to `mode=happy` and clean up the error handling.
Else we have to return the error.)
Else register programming error since this parser wouldn't have recorded itself.

### Parser `FirstSuccessful`

##### happy:
Returns result of first successful parser or after first `NoWayBack`.

##### error:
Help finding the `NoWayBack` to switch over to `mode=handle`. \
It's the last one of the first successful sub-parser.
We know the right sub-parser from the cache and only call it if it failed.

##### handle:
Call sub-parsers until error is found again. \
We know the right sub-parser from the cache and only call it if it failed.


##### record:
If `mode==record` at start then switch to `mode=collect` to find guaranteed
`NoWayBack` and call **all** sub-parsers to 'collect' the ones with `NoWayBack`.
(If they all have a `NoWayBack` then first try to play nice with removed input
or else switch to `mode=choose`,
call all sub-parsers again, choose the one with the least waste,
switch to `mode=record` and finally call the chosen sub-parser again.
Else switch back to `mode=record`, record itself and return.)
Else return because the right sub-parser has already recorded itself.

##### collect:
If `mode==collect` at start then call **all** sub-parsers to 'collect'
the ones with `NoWayBack`.
(If they all have a `NoWayBack` then return signaling guaranteed `NoWayBack`
has been found.
Else return signaling **no** guaranteed `NoWayBack` has been found.)
Else register a programming error
(the `mode==collect` must not escape the initiating parser and its sub-parsers).

##### choose:
If `mode==choose` at start then call **all** sub-parsers to 'choose'
the first one with the least amount of waste, returning signaling
the minimal amount of waste found remembering the choice.
Else register a programming error
(the `mode==choose` must not escape the initiating parser and its sub-parsers).

##### play:
If `mode==play` at start then call the remembered sub-parser and expect the
mode changed to **_happy_** when the sub-parser returns.
Else register a programming error
(the **FirstSuccessful** parser doesn't record itself in this case).

### Branch Parsers (Sequential Combining Parsers)
These are all parsers that apply one or multiple sub-parsers in sequence.
They have got the obligation to witness and handle any error happening
in any of their sub-parsers.

So creating a **branch** parser is much more involved than creating a
**leaf** parser.

##### happy:
Normal parsing potentially witnessing an error and reporting oneself as
witness (using `State.IWitnessed`).

##### error (checked with `State.Failed`):
Use the parsers (that succeeded already) in reverse order to find the last one
with a `NoWayBack` parser that moved the mark. (Please cache this!) \
Return including the error if none could be found.

##### handle:
If the parser witnessed the current error at the current input position
(checked with `State.AmIWitness`) it has to switch to `mode=happy`.

##### record:
Records itself.

##### collect:
Call sub-parser (if any) or do nothing.

##### choose:
Like mode **collect**.

##### play:
Call sub-parser (if any) and expect the mode changed to **happy** when
the sub-parser returns, or do nothing.

### Leaf Parsers

All parsers that don't have any sub-parsers simply do their thing and
possibly call `State.NewError`.
They don't have to care about `mode`s at all. :)
But they can use `State.Semantic` to possibly skip costly semantics.

## Limits

There are some very academic cases that are intentionally not supported by this project:

- The same `FirstSuccessful` parser is used in a cascade more than 8 times
  without any `NoWayBack` parser in between. \
  It is absolutely possible to support this case. \
  But it would drive up the complexity of the `FirstSuccessful` parser.
  And that is one of the most complex parts of the code already.