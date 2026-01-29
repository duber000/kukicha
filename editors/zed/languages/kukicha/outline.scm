; Kukicha code outline for symbol navigation

; Functions
(function_declaration
  name: (identifier) @name) @item

; Methods
(method_declaration
  name: (identifier) @name
  receiver_type: (_) @context) @item

; Types (structs)
(type_declaration
  name: (identifier) @name) @item

; Interfaces
(interface_declaration
  name: (identifier) @name) @item

; Package declaration
(petiole_declaration
  name: (identifier) @name) @item
