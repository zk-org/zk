# The borrow checker

The *borrow checker* of Rust's compiler is comparing the scope of borrowed references to the scope of the owned data to prevent [dangling pointers](3cut). It also makes sure that the relationship between *lifetimes* of several reference match.

In some deterministic patterns, the *borrow checker* automatically infer the lifetimes following [lifetime elision rules](t9i4). But when the *borrow checker* can't automatically infer the lifetimes, we need to help it by annotating our references with [generic lifetime annotations](554k).

:programming:rust:
