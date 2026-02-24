; Kukicha syntax highlighting for Zed
; Edit in languages/kukicha/, then run: editors/zed/scripts/sync-highlights.sh

; Boolean literals
(boolean_literal) @constant.builtin

; Numeric literals
(integer_literal) @number
(float_literal) @number

; String literals
(interpreted_string_literal) @string
(raw_string_literal) @string
(plain_string_literal) @string
(string_content) @string
(plain_string_content) @string
(escape_sequence) @string.escape

; String interpolation
(string_interpolation) @embedded

; Character literals
(rune_literal) @string

; Comments
(comment) @comment

; Primitive types
(primitive_type) @type.builtin

; Function declarations
(function_declaration
  name: (identifier) @function)

; Method declarations
(method_declaration
  name: (identifier) @function.method)

; Interface method signatures
(method_signature
  name: (identifier) @function)

; Parameters
(parameter
  name: (identifier) @variable.parameter)

; Field declarations
(field_declaration
  name: (identifier) @property)

; Type declarations
(type_declaration
  name: (identifier) @type)

; Interface declarations
(interface_declaration
  name: (identifier) @type)

; Function calls - match identifier inside primary_expression
(call_expression
  function: (primary_expression (identifier) @function.call))

; Method calls
(call_expression
  function: (primary_expression
    (selector_expression
      field: (identifier) @function.call)))

; Field/property access
(selector_expression
  field: (identifier) @property)

; Struct field initialization
(struct_field_init
  name: (identifier) @property)

; Package declaration
(petiole_declaration
  name: (identifier) @module)

; Import paths
(import_declaration
  path: (interpreted_string_literal) @string)

; Built-in expressions (recover, panic, error, empty, make, close, reference, dereference)
(recover_expression) @function.builtin
(panic_expression) @function.builtin
(error_expression) @function.builtin
(empty_expression) @function.builtin
(make_expression) @function.builtin
(close_expression) @function.builtin
(receive_expression) @function.builtin
(address_of_expression) @function.builtin
(dereference_expression) @function.builtin

; print/min/max builtins (parsed as regular calls, not dedicated grammar nodes)
(call_expression
  function: (primary_expression
    (identifier) @function.builtin
    (#any-of? @function.builtin "print" "min" "max")))

; Type assertions and casts
(type_assertion) @keyword.operator
(type_cast) @keyword.operator

; Identifiers (general fallback - should be last)
(identifier) @variable

; Arrow lambda parameters
(arrow_lambda
  parameters: (arrow_lambda_params
    (parameter
      name: (identifier) @variable.parameter)))

; Single untyped arrow lambda parameter
(arrow_lambda
  parameters: (arrow_lambda_params
    name: (identifier) @variable.parameter))

; Operators
[":=" "=" "|>" "=>" "==" "!=" "<" ">" "<=" ">=" "+" "-" "*" "/" "%" "++" "--" "&&" "||" "!" "|"] @operator

; Keywords
["petiole" "import" "as" "type" "interface" "var" "variable" "func" "function" "on" "return" "if" "else" "for" "in" "from" "to" "through" "switch" "select" "when" "otherwise" "default" "go" "defer" "make" "list" "of" "map" "channel" "send" "receive" "close" "panic" "error" "empty" "discard" "many" "true" "false" "equals" "and" "or" "onerr" "explain" "not" "reference" "dereference" "nil" "break" "continue"] @keyword

; Control flow keywords (optional more specific highlighting)
["if" "else" "for" "switch" "select" "when" "otherwise" "default" "return" "break" "continue" "go" "defer" "onerr"] @keyword.control

; Logical operators as keywords
["and" "or" "not" "in" "equals"] @keyword.operator

; Punctuation - Brackets
["(" ")" "[" "]" "{" "}"] @punctuation.bracket

; Punctuation - Delimiters
["," "." ":"] @punctuation.delimiter
