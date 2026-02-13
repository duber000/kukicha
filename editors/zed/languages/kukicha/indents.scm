; Kukicha auto-indentation rules

; Increase indent after these
(function_declaration) @indent
(method_declaration) @indent
(type_declaration) @indent
(interface_declaration) @indent
(if_statement) @indent
(else_clause) @indent
(for_range_statement) @indent
(for_numeric_statement) @indent
(for_condition_statement) @indent
(for_infinite_statement) @indent
(switch_statement) @indent
(switch_case) @indent
(function_literal) @indent
(arrow_lambda) @indent
(go_statement) @indent

; Block ends
(block) @indent @end
