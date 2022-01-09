# The Stack and the Heap

Both the Stack and the Heap are parts of the memory that is accessible to an app during runtime.

## Stack

* The Stack stores values in a last in, first out fashion.
* All values stored must have a known fixed size.
    * Data with unknown or changeable size must be stored on the Heap instead.
* Are stored on the Stack:
    * Arguments and local variables when calling a function.

## Heap

* The Heap is less organized than the Stack
* When *allocating on the Heap*, the memory allocator:
    1. Looks for an empty spot for the requested size.
    2. Marks the spot as reserved.
    3. Returns a pointer to the spot.
    * The pointer is then usually stored on the Stack. When we want to access the data, we must follow the pointer to the Heap.

## Stack vs Heap

* *Pushing to the Stack* is faster than *allocating on the Heap* because the Stack doesn't need to find an empty spot.
* Accessing data from the Stack is also faster because we don't have to jump around in the memory and to follow pointers.

:programming:
