# Mutex

*   Abbreviation of *mutual exclusion*.
*   An approach to manage safely shared state by allowing only a single thread to access a protected value at one time.
*   A mutex *guards* a protected data with a *locking system*.
*   Managing mutexes is tricky, using [channels](../fwsj) is an easier alternative.
    *   The main risk is to create *deadlocks*.
    *   Thanks to its [Ownership](../88el) pattern, Rust makes sure we can't mess up when using locks.

:programming:
