# TODO

## Parser IDs

* Parsers don't know their ID (Problem???)
* AnyParser s do.
* Parsers could optionally implement `SetID(int32)` and `ID() int32` themselves.
* Similar with an optional `ParseAfterError` method.


## How to Store Data for Error Handling

| Criterium  | Error          | Result         | Parser        |
|------------|----------------| -------------- |---------------|
| clean up   | trivial        | almost trivial | hard          |
| store data | with parser ID | with parser ID | trivial       |
| get data   | with parser ID | with parser ID | with error ID |

### Conclusion

Storing temporary parse results and data for error handling in the error
is the cleanest solution.

* Errors don't need IDs.
* `ParseAfterError` is it.
* `NewParser` function for reduced interface?