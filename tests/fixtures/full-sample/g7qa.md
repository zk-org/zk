# Concurrency in Rust

*   Thanks to the [Ownership pattern](88el), Rust has a model of [Fearless concurrency](2cl7).
*   Rust aims to have a small runtime, so it doesn't support [green threads](inbox/my59).
    *   Crates exist to add support for green threads if needed.
    *   Instead, Rust relies on the OS threads, a model called 1-1.

*   Rust offers a number of constructs for sharing data between threads:
    *   [Channel](fwsj) for a safe [message passing](4oma) approach.
    *   [Mutex](inbox/er4k) for managing shared state.

:rust:programming:
