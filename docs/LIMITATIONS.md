# Kukicha Language Limitations

Known gaps between Kukicha syntax and Go patterns. 

1. empty is a reserved keyword — can't be used as a
  variable name - same as error
2. Non-generic slice functions (Unique, Contains,
  IndexOf, Get, FirstOne, LastOne, Find, Pop, Shift) take
   []any literally, not []T — must pass list of any{...}
  at call sites
3. Float literals lose precision in codegen — avoid
  small tolerances like 0.000000001; use simple range
  checks (x < 3.14 or x > 3.15) instead
