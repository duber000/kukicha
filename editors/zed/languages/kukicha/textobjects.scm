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

; Comments
(comment) @comment.around
(comment) @comment.inside
