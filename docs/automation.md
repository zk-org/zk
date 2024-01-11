# Automating frequent tasks

`zk` was designed with automation in mind and strive to be [a good Unix citizen](https://en.wikipedia.org/wiki/Unix_philosophy). As such, it offers a number of ways to interface with other programs:

* write [command aliases](config-alias.md) or [named filters](config-filter.md) for repeated complex commands
* [call `zk` from other programs](external-call.md)
* [send notes for processing by other programs](external-processing.md)
* [create a note with initial content](note-creation.md) from a standard input pipe

If you find out that `zk` does not behave as expected or could communicate better with other programs, [please post an issue](https://github.com/zk-org/zk/issues).
