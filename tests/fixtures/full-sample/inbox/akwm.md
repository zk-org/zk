# Errors should be handled differently in an application versus a library

Error handling should be approached differently depending on the context.

*   A *library* should focus on *producing* errors with meaningful context, by wrapping lower-level errors.
*   An *application* mainly consumes errors, by deciding how they are formatted and presented to the user.

## References

* [Rust: Structuring and handling errors in 2020 - nick.groenen.me](https://nick.groenen.me/posts/rust-error-handling/)

:programming:
