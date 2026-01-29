/**
 * External scanner for Kukicha
 * Handles INDENT, DEDENT, and NEWLINE tokens for indentation-based syntax
 */

#include "tree_sitter/parser.h"
#include <stdlib.h>
#include <string.h>

#define MAX_INDENT_DEPTH 100

enum TokenType {
    INDENT,
    DEDENT,
    NEWLINE,
};

typedef struct {
    uint16_t indent_stack[MAX_INDENT_DEPTH];
    uint16_t indent_depth;
    uint16_t pending_dedents;
} Scanner;

static void advance(TSLexer *lexer) {
    lexer->advance(lexer, false);
}

static void skip(TSLexer *lexer) {
    lexer->advance(lexer, true);
}

void *tree_sitter_kukicha_external_scanner_create(void) {
    Scanner *scanner = malloc(sizeof(Scanner));
    if (scanner) {
        scanner->indent_depth = 1;
        scanner->indent_stack[0] = 0;
        scanner->pending_dedents = 0;
        for (int i = 1; i < MAX_INDENT_DEPTH; i++) {
            scanner->indent_stack[i] = 0;
        }
    }
    return scanner;
}

void tree_sitter_kukicha_external_scanner_destroy(void *payload) {
    free(payload);
}

unsigned tree_sitter_kukicha_external_scanner_serialize(
    void *payload,
    char *buffer
) {
    Scanner *scanner = (Scanner *)payload;

    size_t size = 0;

    buffer[size++] = (char)scanner->pending_dedents;
    buffer[size++] = (char)scanner->indent_depth;

    for (uint16_t i = 0; i < scanner->indent_depth && size < TREE_SITTER_SERIALIZATION_BUFFER_SIZE; i++) {
        buffer[size++] = (char)(scanner->indent_stack[i] & 0xFF);
        buffer[size++] = (char)((scanner->indent_stack[i] >> 8) & 0xFF);
    }

    return size;
}

void tree_sitter_kukicha_external_scanner_deserialize(
    void *payload,
    const char *buffer,
    unsigned length
) {
    Scanner *scanner = (Scanner *)payload;

    scanner->pending_dedents = 0;
    scanner->indent_depth = 1;
    scanner->indent_stack[0] = 0;

    if (length == 0) return;

    size_t pos = 0;

    scanner->pending_dedents = (uint16_t)(unsigned char)buffer[pos++];

    if (pos >= length) return;

    scanner->indent_depth = (uint16_t)(unsigned char)buffer[pos++];
    if (scanner->indent_depth > MAX_INDENT_DEPTH) {
        scanner->indent_depth = MAX_INDENT_DEPTH;
    }

    for (uint16_t i = 0; i < scanner->indent_depth && pos + 1 < length; i++) {
        scanner->indent_stack[i] = (uint16_t)((unsigned char)buffer[pos] | ((unsigned char)buffer[pos + 1] << 8));
        pos += 2;
    }
}

static bool is_eof(TSLexer *lexer) {
    return lexer->eof(lexer);
}

bool tree_sitter_kukicha_external_scanner_scan(
    void *payload,
    TSLexer *lexer,
    const bool *valid_symbols
) {
    Scanner *scanner = (Scanner *)payload;

    // Handle pending dedents first (highest priority)
    if (scanner->pending_dedents > 0 && valid_symbols[DEDENT]) {
        scanner->pending_dedents--;
        lexer->result_symbol = DEDENT;
        return true;
    }

    bool at_line_start = (lexer->get_column(lexer) == 0);

    // At the start of a line, count indentation and check for INDENT/DEDENT
    if (at_line_start && (valid_symbols[INDENT] || valid_symbols[DEDENT])) {
        // Skip blank lines and count indentation
        while (true) {
            uint16_t indent = 0;

            // Count leading whitespace
            while (!is_eof(lexer) && (lexer->lookahead == ' ' || lexer->lookahead == '\t')) {
                if (lexer->lookahead == ' ') {
                    indent++;
                } else {
                    indent += 4;
                }
                skip(lexer);
            }

            // Check for blank line
            if (lexer->lookahead == '\n') {
                skip(lexer);
                continue;
            }
            if (lexer->lookahead == '\r') {
                skip(lexer);
                if (lexer->lookahead == '\n') {
                    skip(lexer);
                }
                continue;
            }

            // Check for comment line
            if (lexer->lookahead == '#') {
                while (!is_eof(lexer) && lexer->lookahead != '\n' && lexer->lookahead != '\r') {
                    skip(lexer);
                }
                if (lexer->lookahead == '\n' || lexer->lookahead == '\r') {
                    skip(lexer);
                    if (lexer->lookahead == '\n') {
                        skip(lexer);
                    }
                }
                continue;
            }

            // EOF handling
            if (is_eof(lexer)) {
                if (scanner->indent_depth > 1 && valid_symbols[DEDENT]) {
                    scanner->indent_depth--;
                    lexer->result_symbol = DEDENT;
                    return true;
                }
                return false;
            }

            // Found content - mark the position (we don't consume content)
            lexer->mark_end(lexer);

            uint16_t current_indent = scanner->indent_stack[scanner->indent_depth - 1];

            if (indent > current_indent && valid_symbols[INDENT]) {
                if (scanner->indent_depth < MAX_INDENT_DEPTH) {
                    scanner->indent_stack[scanner->indent_depth++] = indent;
                }
                lexer->result_symbol = INDENT;
                return true;
            }

            if (indent < current_indent && valid_symbols[DEDENT]) {
                uint16_t dedent_count = 0;
                while (scanner->indent_depth > 1 &&
                       scanner->indent_stack[scanner->indent_depth - 1] > indent) {
                    scanner->indent_depth--;
                    dedent_count++;
                }

                if (dedent_count > 0) {
                    scanner->pending_dedents = dedent_count - 1;
                    lexer->result_symbol = DEDENT;
                    return true;
                }
            }

            // Same indent or no valid token - done
            break;
        }
    }

    // Check for NEWLINE
    if (valid_symbols[NEWLINE]) {
        // Skip trailing whitespace
        while (lexer->lookahead == ' ' || lexer->lookahead == '\t') {
            skip(lexer);
        }

        if (lexer->lookahead == '\n') {
            advance(lexer);
            if (lexer->lookahead == '\r') {
                advance(lexer);
            }
            lexer->result_symbol = NEWLINE;
            return true;
        }

        if (lexer->lookahead == '\r') {
            advance(lexer);
            if (lexer->lookahead == '\n') {
                advance(lexer);
            }
            lexer->result_symbol = NEWLINE;
            return true;
        }

        // At EOF, treat as newline
        if (is_eof(lexer)) {
            lexer->result_symbol = NEWLINE;
            return true;
        }
    }

    return false;
}
