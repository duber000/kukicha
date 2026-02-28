# Kukicha Language Limitations

No known limitations at this time.

Previously documented gaps (all resolved):

1. ~~`empty` and `error` as variable names~~ — Fixed: context-sensitive
   parsing now allows both keywords to be used as identifiers when
   followed by assignment, operators, or delimiters.

2. ~~Non-generic slice functions~~ — Fixed: all slice functions are now
   generic. `Get`, `FirstOne`, `LastOne`, `Find`, `FindLast`, `Pop`,
   `Shift` use `[T any]`; `Unique`, `Contains`, `IndexOf` use
   `[K comparable]`.

3. ~~Float literal precision~~ — Fixed: codegen now preserves the original
   source text for float literals instead of formatting through `%f`.
