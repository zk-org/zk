---
aliases: [dangling reference]
---

# Dangling pointers

A *dangling pointer* is a reference that is kept to freed data. With C, reading it causes a *segmentation fault*.

Rust protects against *dangling pointers* by making sure data is not freed until it goes out of scope ([Ownership in Rust](88el)).

:programming:
