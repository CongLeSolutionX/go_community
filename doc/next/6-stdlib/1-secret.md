### New secret package

The new [secret](/pkg/runtime/secret) package provides a facility for
securely erasing temporaries used in code that manipulates secret
information, typically cryptographic in nature.

The secret.Do function runs its function argument and then erases all
temporary storage (registers, stack, new heap allocations) used by
that function argument. Heap storage is not erased until that storage
is deemed unreachable by the garbage collector, which might take some
time after secret.Do completes.

This package is intended to make it easier to ensure [forward
secrecy](https://en.wikipedia.org/wiki/Forward_secrecy).
