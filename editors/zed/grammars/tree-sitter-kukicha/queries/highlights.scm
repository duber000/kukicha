; Kukicha syntax highlighting for Zed

; Keywords - Package and imports
["petiole" "import" "as"] @keyword

; Keywords - Declarations
["func" "type" "interface"] @keyword

; Keywords - Control flow
["if" "else" "for" "in" "from" "to" "through" "switch" "case" "default"] @keyword.control

; Keywords - Flow control
["return" "break" "continue" "defer" "go"] @keyword.control.return

; Keywords - Error handling
["onerr" "panic" "recover" "error"] @keyword.exception

; Keywords - Logical operators (word form)
["and" "or" "not" "equals"] @keyword.operator

; Keywords - Type-related
["list" "map" "channel" "of" "to" "reference" "dereference" "on" "many"] @keyword.type

; Keywords - Special values
["empty" "discard"] @constant.builtin

; Keywords - Concurrency
["send" "receive" "close" "make"] @keyword

; Boolean literals
(boolean_literal) @constant.builtin.boolean

; Numeric literals
(integer_literal) @number
(float_literal) @number.float

; String literals
(interpreted_string_literal) @string
(raw_string_literal) @string
(string_content) @string
(escape_sequence) @string.escape

; String interpolation
(string_interpolation
  "{" @punctuation.special
  "}" @punctuation.special) @embedded

; Character literals
(rune_literal) @character

; Comments
(comment) @comment

; Identifiers
(identifier) @variable

; Type identifiers
(simple_type (identifier) @type)
(qualified_type
  package: (identifier) @namespace
  name: (identifier) @type)

; Primitive types
(primitive_type) @type.builtin

; Function declarations
(function_declaration
  name: (identifier) @function)

; Method declarations
(method_declaration
  name: (identifier) @function.method
  receiver: (identifier) @variable.parameter)

; Interface method signatures
(method_signature
  name: (identifier) @function.method)

; Parameters
(parameter
  name: (identifier) @variable.parameter)

; Field declarations
(field_declaration
  name: (identifier) @property)

; Struct tags
(struct_tag
  (identifier) @attribute
  (interpreted_string_literal) @string)

; Type declarations
(type_declaration
  name: (identifier) @type)

; Interface declarations
(interface_declaration
  name: (identifier) @type)

; Function calls
(call_expression
  function: (identifier) @function.call)

(call_expression
  function: (selector_expression
    field: (identifier) @function.method.call))

; Field/property access
(selector_expression
  field: (identifier) @property)

; Struct field initialization
(struct_field_init
  name: (identifier) @property)

; Variable declarations
(var_declaration
  names: (identifier_list (identifier) @variable))

; Operators
[":=" "="] @operator

; Pipe operator (special)
"|>" @operator

; Comparison operators
["==" "!=" "<" ">" "<=" ">="] @operator

; Arithmetic operators
["+" "-" "*" "/" "%"] @operator

; Increment/decrement
["++" "--"] @operator

; Logical operators (symbol form)
["&&" "||" "!" "|"] @operator

; Punctuation - Brackets
["(" ")" "[" "]" "{" "}"] @punctuation.bracket

; Punctuation - Delimiters
["," "." ":"] @punctuation.delimiter

; Package declaration
(petiole_declaration
  name: (identifier) @namespace)

; Import paths
(import_declaration
  path: (interpreted_string_literal) @string.special.path
  alias: (identifier)? @namespace)
