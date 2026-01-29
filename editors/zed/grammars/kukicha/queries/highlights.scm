; Kukicha syntax highlighting for Zed

; Boolean literals
(boolean_literal) @constant.builtin

; Numeric literals
(integer_literal) @number
(float_literal) @number

; String literals
(interpreted_string_literal) @string
(raw_string_literal) @string
(string_content) @string
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

; Identifiers (general fallback - should be last)
(identifier) @variable

; Operators
[":=" "=" "|>" "==" "!=" "<" ">" "<=" ">=" "+" "-" "*" "/" "%" "++" "--" "&&" "||" "!" "|"] @operator

; Punctuation - Brackets
["(" ")" "[" "]" "{" "}"] @punctuation.bracket

; Punctuation - Delimiters
["," "." ":"] @punctuation.delimiter
