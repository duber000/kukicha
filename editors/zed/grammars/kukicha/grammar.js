/**
 * Tree-sitter grammar for Kukicha
 * A beginner-friendly language that compiles to Go
 */

module.exports = grammar({
  name: 'kukicha',

  extras: $ => [
    /\s/,
    $.comment,
  ],

  externals: $ => [
    $._indent,
    $._dedent,
    $._newline,
  ],

  word: $ => $.identifier,

  conflicts: $ => [
    [$.simple_type, $.primary_expression],
    [$.qualified_type, $.primary_expression],
    [$.list_literal, $.map_literal],
    [$.simple_type, $.qualified_type],
  ],

  rules: {
    source_file: $ => seq(
      optional($.petiole_declaration),
      repeat($.import_declaration),
      repeat($._declaration),
    ),

    // Package declaration
    petiole_declaration: $ => seq(
      'petiole',
      field('name', $.identifier),
      $._newline,
    ),

    // Import declarations
    import_declaration: $ => seq(
      'import',
      field('path', $.interpreted_string_literal),
      optional(seq('as', field('alias', $.identifier))),
      $._newline,
    ),

    // Top-level declarations
    _declaration: $ => choice(
      $.function_declaration,
      $.method_declaration,
      $.type_declaration,
      $.interface_declaration,
      $.var_declaration,
    ),

    // Function declaration
    function_declaration: $ => seq(
      choice('func', 'function'),
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      optional(field('return_type', $.return_type)),
      $._newline,
      field('body', $.block),
    ),

    // Method declaration: func Name on receiver Type ReturnType
    method_declaration: $ => seq(
      choice('func', 'function'),
      field('name', $.identifier),
      'on',
      field('receiver', $.identifier),
      field('receiver_type', $.type),
      optional(field('return_type', $.return_type)),
      $._newline,
      field('body', $.block),
    ),

    // Type declaration
    type_declaration: $ => seq(
      'type',
      field('name', $.identifier),
      $._newline,
      field('body', $.type_body),
    ),

    type_body: $ => seq(
      $._indent,
      repeat1($.field_declaration),
      $._dedent,
    ),

    field_declaration: $ => seq(
      field('name', $.identifier),
      field('type', $.type),
      optional(field('tag', $.struct_tag)),
      $._newline,
    ),

    struct_tag: $ => repeat1(seq(
      $.identifier,
      ':',
      $.interpreted_string_literal,
    )),

    // Interface declaration
    interface_declaration: $ => seq(
      'interface',
      field('name', $.identifier),
      $._newline,
      field('body', $.interface_body),
    ),

    interface_body: $ => seq(
      $._indent,
      repeat1($.method_signature),
      $._dedent,
    ),

    method_signature: $ => seq(
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      optional(field('return_type', $.return_type)),
      $._newline,
    ),

    // Parameters
    parameter_list: $ => seq(
      '(',
      optional(commaSep1($.parameter)),
      ')',
    ),

    parameter: $ => seq(
      optional('many'),
      field('name', $.identifier),
      field('type', $.type),
    ),

    return_type: $ => choice(
      $.type,
      $.multiple_return_type,
    ),

    multiple_return_type: $ => seq(
      $.type,
      ',',
      commaSep1($.type),
    ),

    // Types
    type: $ => choice(
      $.simple_type,
      $.list_type,
      $.map_type,
      $.channel_type,
      $.reference_type,
      $.function_type,
    ),

    simple_type: $ => choice(
      $.identifier,
      $.qualified_type,
      $.primitive_type,
    ),

    qualified_type: $ => seq(
      field('package', $.identifier),
      '.',
      field('name', $.identifier),
    ),

    primitive_type: $ => choice(
      'int', 'int8', 'int16', 'int32', 'int64',
      'uint', 'uint8', 'uint16', 'uint32', 'uint64',
      'float32', 'float64',
      'string', 'bool', 'byte', 'rune', 'any', 'error',
    ),

    list_type: $ => seq(
      'list',
      'of',
      field('element', $.type),
    ),

    map_type: $ => seq(
      'map',
      'of',
      field('key', $.type),
      'to',
      field('value', $.type),
    ),

    channel_type: $ => seq(
      'channel',
      'of',
      field('element', $.type),
    ),

    reference_type: $ => seq(
      'reference',
      field('type', $.type),
    ),

    function_type: $ => prec.right(seq(
      choice('func', 'function'),
      '(',
      optional(commaSep1($.type)),
      ')',
      optional($.type),
    )),

    // Block (indented)
    block: $ => seq(
      $._indent,
      repeat1($._statement),
      $._dedent,
    ),

    // Statements
    _statement: $ => choice(
      $.var_declaration,
      $.short_var_declaration,
      $.assignment_statement,
      $.expression_statement,
      $.return_statement,
      $.if_statement,
      $.for_range_statement,
      $.for_numeric_statement,
      $.for_condition_statement,
      $.for_infinite_statement,
      $.switch_statement,
      $.defer_statement,
      $.go_statement,
      $.send_statement,
      $.break_statement,
      $.continue_statement,
      $.inc_dec_statement,
    ),

    // Explicit variable declaration: var x int = 1
    var_declaration: $ => seq(
      choice('var', 'variable'),
      field('names', $.identifier_list),
      choice(
        seq(field('type', $.type), optional(seq('=', field('values', $.expression_list)))),
        seq('=', field('values', $.expression_list)),
      ),
      optional($.onerr_clause),
      $._newline,
    ),

    // Variable declaration with :=
    short_var_declaration: $ => seq(
      field('names', $.identifier_list),
      ':=',
      field('values', $.expression_list),
      optional($.onerr_clause),
      $._newline,
    ),

    identifier_list: $ => commaSep1($.identifier),
    expression_list: $ => commaSep1($._expression),

    // Assignment with =
    assignment_statement: $ => seq(
      field('left', $._lvalue),
      '=',
      field('right', $.expression_list),
      optional($.onerr_clause),
      $._newline,
    ),

    _lvalue: $ => choice(
      $.index_expression,
      $.selector_expression,
      $.identifier_list,
    ),

    // Expression statement
    expression_statement: $ => seq(
      $._expression,
      optional($.onerr_clause),
      $._newline,
    ),

    // Increment/Decrement
    inc_dec_statement: $ => seq(
      field('operand', $._expression),
      field('operator', choice('++', '--')),
      $._newline,
    ),

    // Return statement
    return_statement: $ => seq(
      'return',
      optional(field('value', $._expression)),
      $._newline,
    ),

    // If statement
    if_statement: $ => seq(
      'if',
      field('condition', $._expression),
      $._newline,
      field('consequence', $.block),
      optional($.else_clause),
    ),

    else_clause: $ => seq(
      'else',
      choice(
        seq($._newline, $.block),
        $.if_statement,
      ),
    ),

    // For range statement: for item in collection
    for_range_statement: $ => prec(2, seq(
      'for',
      field('index', optional(seq($.identifier, ','))),
      field('value', $.identifier),
      'in',
      field('iterable', $._expression),
      $._newline,
      field('body', $.block),
    )),

    // For numeric statement: for i from start to/through end
    for_numeric_statement: $ => prec(2, seq(
      'for',
      field('variable', $.identifier),
      'from',
      field('start', $._expression),
      field('bound_type', choice('to', 'through')),
      field('end', $._expression),
      $._newline,
      field('body', $.block),
    )),

    // For condition statement: for condition
    for_condition_statement: $ => seq(
      'for',
      field('condition', $._expression),
      $._newline,
      field('body', $.block),
    ),

    // Infinite loop: for
    for_infinite_statement: $ => seq(
      'for',
      $._newline,
      field('body', $.block),
    ),

    // Switch statement
    switch_statement: $ => seq(
      'switch',
      optional(field('value', $._expression)),
      $._newline,
      field('body', $.switch_body),
    ),

    switch_body: $ => seq(
      $._indent,
      repeat1($.switch_case),
      $._dedent,
    ),

    switch_case: $ => choice(
      seq(
        'when',
        field('value', commaSep1($._expression)),
        $._newline,
        field('body', $.block),
      ),
      seq(
        choice('otherwise', 'default'),
        $._newline,
        field('body', $.block),
      ),
    ),

    // Defer statement
    defer_statement: $ => seq(
      'defer',
      field('expression', $._expression),
      $._newline,
    ),

    // Go statement
    go_statement: $ => seq(
      'go',
      field('expression', $._expression),
      $._newline,
    ),

    // Send statement
    send_statement: $ => seq(
      'send',
      field('value', $._expression),
      'to',
      field('channel', $._expression),
      $._newline,
    ),

    // Break and continue
    break_statement: $ => seq('break', $._newline),
    continue_statement: $ => seq('continue', $._newline),

    // OnErr clause
    onerr_clause: $ => prec(5, seq(
      'onerr',
      choice(
        field('handler', choice(
          $.panic_expression,
          $.return_expression,
          'discard',
          $._expression,
        )),
        seq($._newline, field('body', $.block)),
      ),
    )),

    // Expressions
    _expression: $ => choice(
      $.primary_expression,
      $.binary_expression,
      $.unary_expression,
      $.pipe_expression,
    ),

    primary_expression: $ => choice(
      $.identifier,
      $.literal,
      $.parenthesized_expression,
      $.call_expression,
      $.index_expression,
      $.selector_expression,
      $.struct_literal,
      $.list_literal,
      $.map_literal,
      $.function_literal,
      $.make_expression,
      $.receive_expression,
      $.address_of_expression,
      $.dereference_expression,
      $.empty_expression,
      $.panic_expression,
      $.error_expression,
      $.recover_expression,
      $.close_expression,
      $.type_assertion,
      $.type_cast,
    ),

    parenthesized_expression: $ => seq('(', $._expression, ')'),

    // Binary expressions
    binary_expression: $ => choice(
      prec.left(1, seq($._expression, 'or', $._expression)),
      prec.left(1, seq($._expression, '||', $._expression)),
      prec.left(2, seq($._expression, 'and', $._expression)),
      prec.left(2, seq($._expression, '&&', $._expression)),
      prec.left(3, seq($._expression, '|', $._expression)),
      prec.left(4, seq($._expression, choice('==', '!=', 'equals'), $._expression)),
      prec.left(5, seq($._expression, choice('<', '>', '<=', '>='), $._expression)),
      prec.left(5, seq($._expression, 'in', $._expression)),
      prec.left(6, seq($._expression, choice('+', '-'), $._expression)),
      prec.left(7, seq($._expression, choice('*', '/', '%'), $._expression)),
    ),

    // Unary expressions
    unary_expression: $ => choice(
      prec.right(8, seq('not', $._expression)),
      prec.right(8, seq('!', $._expression)),
      prec.right(8, seq('-', $._expression)),
    ),

    // Pipe expression
    pipe_expression: $ => prec.left(1, seq(
      field('left', $._expression),
      '|>',
      field('right', $._expression),
    )),

    // Call expression
    call_expression: $ => prec(10, seq(
      field('function', $.primary_expression),
      field('arguments', $.argument_list),
    )),

    argument_list: $ => seq(
      '(',
      optional(commaSep1($.argument)),
      ')',
    ),

    argument: $ => choice(
      $._expression,
      '_', // Placeholder for pipes
    ),

    // Index expression
    index_expression: $ => prec(10, seq(
      field('operand', $.primary_expression),
      '[',
      field('index', $._expression),
      ']',
    )),

    // Selector expression (field access or method call)
    selector_expression: $ => prec(10, seq(
      field('operand', $.primary_expression),
      '.',
      field('field', $.identifier),
    )),

    // Struct literal
    struct_literal: $ => seq(
      field('type', $.simple_type),
      $._newline,
      $._indent,
      repeat1($.struct_field_init),
      $._dedent,
    ),

    struct_field_init: $ => seq(
      field('name', $.identifier),
      ':',
      field('value', $._expression),
      $._newline,
    ),

    // List literal
    list_literal: $ => choice(
      // Inline form
      seq(
        optional(seq('list', 'of', $.type)),
        '{',
        optional(commaSep1($._expression)),
        '}',
      ),
      // Type prefix form
      seq(
        'list',
        'of',
        $.type,
        '{',
        optional(commaSep1($._expression)),
        '}',
      ),
    ),

    // Map literal
    map_literal: $ => seq(
      optional(seq('map', 'of', $.type, 'to', $.type)),
      '{',
      optional(commaSep1($.map_entry)),
      '}',
    ),

    map_entry: $ => seq(
      field('key', $._expression),
      ':',
      field('value', $._expression),
    ),

    // Function literal
    function_literal: $ => seq(
      choice('func', 'function'),
      field('parameters', $.parameter_list),
      optional(field('return_type', $.return_type)),
      $._newline,
      field('body', $.block),
    ),

    // Make expression - make(Type) or make(Type, size) or make(Type, size, cap)
    make_expression: $ => seq(
      'make',
      '(',
      field('type', $.type),
      optional(seq(',', commaSep1($._expression))),
      ')',
    ),

    // Channel operations
    receive_expression: $ => seq(
      'receive',
      'from',
      field('channel', $._expression),
    ),

    close_expression: $ => seq(
      'close',
      '(',
      field('channel', $._expression),
      ')',
    ),

    // Pointer operations
    address_of_expression: $ => seq(
      'reference',
      'of',
      field('operand', $._expression),
    ),

    dereference_expression: $ => seq(
      'dereference',
      field('operand', $._expression),
    ),

    // Special expressions
    empty_expression: $ => seq(
      choice('empty', 'nil'),
      optional(field('type', $.type)),
    ),

    panic_expression: $ => seq(
      'panic',
      field('message', $._expression),
    ),

    error_expression: $ => seq(
      'error',
      field('message', $._expression),
    ),

    recover_expression: $ => 'recover',

    return_expression: $ => seq(
      'return',
      optional(field('value', $._expression)),
    ),

    // Type assertion
    type_assertion: $ => prec(10, seq(
      field('operand', $.primary_expression),
      '.',
      '(',
      field('type', $.type),
      ')',
    )),

    // Type cast
    type_cast: $ => prec(9, seq(
      field('operand', $._expression),
      'as',
      field('type', $.type),
    )),

    // Literals
    literal: $ => choice(
      $.integer_literal,
      $.float_literal,
      $.interpreted_string_literal,
      $.raw_string_literal,
      $.rune_literal,
      $.boolean_literal,
    ),

    integer_literal: $ => token(choice(
      /0[xX][0-9a-fA-F][0-9a-fA-F_]*/,
      /0[oO][0-7][0-7_]*/,
      /0[bB][01][01_]*/,
      /[0-9][0-9_]*/,
    )),

    float_literal: $ => token(choice(
      /[0-9][0-9_]*\.[0-9][0-9_]*/,
      /[0-9][0-9_]*\.[0-9][0-9_]*[eE][+-]?[0-9]+/,
      /[0-9][0-9_]*[eE][+-]?[0-9]+/,
    )),

    interpreted_string_literal: $ => seq(
      '"',
      repeat(choice(
        $.string_content,
        $.string_interpolation,
        $.escape_sequence,
      )),
      '"',
    ),

    string_content: $ => token.immediate(prec(1, /[^"\\{]+/)),

    string_interpolation: $ => seq(
      '{',
      $._expression,
      '}',
    ),

    escape_sequence: $ => token.immediate(seq(
      '\\',
      choice(
        /[nrtfvb\\'"]/,
        /x[0-9a-fA-F]{2}/,
        /u[0-9a-fA-F]{4}/,
        /U[0-9a-fA-F]{8}/,
        /[0-7]{1,3}/,
      ),
    )),

    raw_string_literal: $ => seq(
      '`',
      token.immediate(prec(1, /[^`]*/)),
      '`',
    ),

    rune_literal: $ => seq(
      "'",
      choice(
        /[^'\\]/,
        $.escape_sequence,
      ),
      "'",
    ),

    boolean_literal: $ => choice('true', 'false'),

    // Identifier
    identifier: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,

    // Comment
    comment: $ => token(seq('#', /.*/)),
  },
});

function commaSep1(rule) {
  return seq(rule, repeat(seq(',', rule)));
}

function commaSep(rule) {
  return optional(commaSep1(rule));
}
