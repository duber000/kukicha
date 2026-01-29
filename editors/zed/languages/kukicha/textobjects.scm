; Kukicha text objects for Vim-style motions

; Functions
(function_declaration) @function.around
(function_declaration
  body: (block) @function.inside)

; Methods
(method_declaration) @function.around
(method_declaration
  body: (block) @function.inside)

; Types (classes in vim terms)
(type_declaration) @class.around
(type_declaration
  body: (type_body) @class.inside)

; Interfaces
(interface_declaration) @class.around
(interface_declaration
  body: (interface_body) @class.inside)

; Parameters
(parameter) @parameter.around
(parameter
  name: (identifier) @parameter.inside)

; Comments
(comment) @comment.around
(comment) @comment.inside

; Conditionals
(if_statement) @conditional.around
(if_statement
  consequence: (block) @conditional.inside)

; Loops
(for_range_statement) @loop.around
(for_range_statement
  body: (block) @loop.inside)

(for_numeric_statement) @loop.around
(for_numeric_statement
  body: (block) @loop.inside)

(for_condition_statement) @loop.around
(for_condition_statement
  body: (block) @loop.inside)
