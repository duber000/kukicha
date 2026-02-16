package formatter

import (
	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/lexer"
)

// Comment represents a comment in the source code
type Comment struct {
	Text       string // The comment text including the # prefix
	Line       int    // Line number where the comment appears
	Column     int    // Column number where the comment starts
	IsTrailing bool   // True if comment is on same line as code
}

// CommentAttachment holds comments attached to an AST node
type CommentAttachment struct {
	Leading  []Comment // Comments on lines immediately before the node
	Trailing *Comment  // Comment on same line after the node (optional)
}

// CommentMap maps AST nodes to their attached comments
type CommentMap map[ast.Node]*CommentAttachment

// ExtractComments extracts all comment tokens from a token stream
func ExtractComments(tokens []lexer.Token) []Comment {
	var comments []Comment

	for _, tok := range tokens {
		if tok.Type == lexer.TOKEN_COMMENT {
			comments = append(comments, Comment{
				Text:   tok.Lexeme,
				Line:   tok.Line,
				Column: tok.Column,
			})
		}
	}

	return comments
}

// AttachComments attaches comments to AST nodes
// Comments are attached based on line proximity:
// - Leading comments: on lines immediately before the node (no blank lines between)
// - Trailing comments: on the same line as the node
func AttachComments(comments []Comment, program *ast.Program) CommentMap {
	cm := make(CommentMap)

	if len(comments) == 0 {
		return cm
	}

	// Build a map of line -> last token on that line
	// This helps determine if a comment is trailing
	nodeLines := collectNodeLines(program)

	// Process each comment
	for i := range comments {
		comment := &comments[i]

		// Check if this is a trailing comment (same line as a node)
		if nodeLines[comment.Line] {
			comment.IsTrailing = true
		}
	}

	// Attach comments to nodes
	attachCommentsToProgram(comments, program, cm)

	return cm
}

// collectNodeLines returns a set of lines that have AST nodes starting on them
func collectNodeLines(program *ast.Program) map[int]bool {
	lines := make(map[int]bool)

	// Collect from petiole declaration
	if program.PetioleDecl != nil {
		lines[program.PetioleDecl.Pos().Line] = true
	}

	// Collect from imports
	for _, imp := range program.Imports {
		lines[imp.Pos().Line] = true
	}

	// Collect from declarations
	for _, decl := range program.Declarations {
		collectDeclLines(decl, lines)
	}

	return lines
}

func collectDeclLines(decl ast.Declaration, lines map[int]bool) {
	lines[decl.Pos().Line] = true

	switch d := decl.(type) {
	case *ast.FunctionDecl:
		if d.Body != nil {
			collectBlockLines(d.Body, lines)
		}
	case *ast.TypeDecl:
		if d.AliasType == nil {
			for _, field := range d.Fields {
				lines[field.Name.Token.Line] = true
			}
		}
	case *ast.InterfaceDecl:
		for _, method := range d.Methods {
			lines[method.Name.Token.Line] = true
		}
	}
}

func collectBlockLines(block *ast.BlockStmt, lines map[int]bool) {
	for _, stmt := range block.Statements {
		collectStmtLines(stmt, lines)
	}
}

func collectStmtLines(stmt ast.Statement, lines map[int]bool) {
	lines[stmt.Pos().Line] = true

	switch s := stmt.(type) {
	case *ast.IfStmt:
		collectBlockLines(s.Consequence, lines)
		if s.Alternative != nil {
			switch alt := s.Alternative.(type) {
			case *ast.ElseStmt:
				collectBlockLines(alt.Body, lines)
			case *ast.IfStmt:
				collectStmtLines(alt, lines)
			}
		}
	case *ast.ForRangeStmt:
		collectBlockLines(s.Body, lines)
	case *ast.ForNumericStmt:
		collectBlockLines(s.Body, lines)
	case *ast.ForConditionStmt:
		collectBlockLines(s.Body, lines)
	case *ast.SwitchStmt:
		for _, c := range s.Cases {
			collectBlockLines(c.Body, lines)
		}
		if s.Otherwise != nil {
			collectBlockLines(s.Otherwise.Body, lines)
		}
	}
}

func attachCommentsToProgram(comments []Comment, program *ast.Program, cm CommentMap) {
	commentIdx := 0

	// Helper to get next declaration line
	getNextDeclLine := func(idx int) int {
		if program.PetioleDecl != nil && idx == -2 {
			return program.PetioleDecl.Pos().Line
		}
		if idx == -1 {
			if len(program.Imports) > 0 {
				return program.Imports[0].Pos().Line
			}
			if len(program.Declarations) > 0 {
				return program.Declarations[0].Pos().Line
			}
			return -1
		}
		if idx < len(program.Declarations) {
			return program.Declarations[idx].Pos().Line
		}
		return -1
	}

	// Attach leading comments to petiole declaration
	if program.PetioleDecl != nil {
		petioleLine := program.PetioleDecl.Pos().Line
		commentIdx = attachLeadingComments(comments, commentIdx, petioleLine, program.PetioleDecl, cm)
		commentIdx = attachTrailingComment(comments, commentIdx, petioleLine, program.PetioleDecl, cm)
	}

	// Attach comments to imports
	for _, imp := range program.Imports {
		impLine := imp.Pos().Line
		commentIdx = attachLeadingComments(comments, commentIdx, impLine, imp, cm)
		commentIdx = attachTrailingComment(comments, commentIdx, impLine, imp, cm)
	}

	// Attach comments to declarations
	for i, decl := range program.Declarations {
		declLine := decl.Pos().Line
		commentIdx = attachLeadingComments(comments, commentIdx, declLine, decl, cm)

		// Recursively attach comments within the declaration
		attachCommentsToDecl(comments, &commentIdx, decl, cm)

		// Find next decl line for trailing detection
		nextLine := getNextDeclLine(i + 1)
		_ = nextLine // May be used for more sophisticated trailing detection

		commentIdx = attachTrailingComment(comments, commentIdx, declLine, decl, cm)
	}
}

func attachLeadingComments(comments []Comment, startIdx int, nodeLine int, node ast.Node, cm CommentMap) int {
	idx := startIdx

	// Find comments that are leading (immediately before this node)
	var leading []Comment
	for idx < len(comments) && comments[idx].Line < nodeLine {
		// Check if there's a blank line between this comment and the next
		if idx+1 < len(comments) && comments[idx+1].Line < nodeLine {
			// More comments before the node, include this one
			leading = append(leading, comments[idx])
		} else if comments[idx].Line == nodeLine-1 || (len(leading) > 0 && comments[idx].Line == leading[len(leading)-1].Line+1) {
			// Adjacent to node or to previous leading comment
			leading = append(leading, comments[idx])
		} else {
			// Gap between comment and node, this is a standalone comment
			// For now, we'll still attach it as leading
			leading = append(leading, comments[idx])
		}
		idx++
	}

	if len(leading) > 0 {
		if cm[node] == nil {
			cm[node] = &CommentAttachment{}
		}
		cm[node].Leading = leading
	}

	return idx
}

func attachTrailingComment(comments []Comment, startIdx int, nodeLine int, node ast.Node, cm CommentMap) int {
	if startIdx >= len(comments) {
		return startIdx
	}

	// Check if next comment is on same line (trailing)
	if comments[startIdx].Line == nodeLine {
		if cm[node] == nil {
			cm[node] = &CommentAttachment{}
		}
		trailing := comments[startIdx]
		trailing.IsTrailing = true
		cm[node].Trailing = &trailing
		return startIdx + 1
	}

	return startIdx
}

func attachCommentsToDecl(comments []Comment, idx *int, decl ast.Declaration, cm CommentMap) {
	switch d := decl.(type) {
	case *ast.FunctionDecl:
		if d.Body != nil {
			attachCommentsToBlock(comments, idx, d.Body, cm)
		}
	case *ast.TypeDecl:
		if d.AliasType == nil {
			for _, field := range d.Fields {
				fieldLine := field.Name.Token.Line
				*idx = attachLeadingComments(comments, *idx, fieldLine, field.Name, cm)
				*idx = attachTrailingComment(comments, *idx, fieldLine, field.Name, cm)
			}
		}
	case *ast.InterfaceDecl:
		for _, method := range d.Methods {
			methodLine := method.Name.Token.Line
			*idx = attachLeadingComments(comments, *idx, methodLine, method.Name, cm)
			*idx = attachTrailingComment(comments, *idx, methodLine, method.Name, cm)
		}
	}
}

func attachCommentsToBlock(comments []Comment, idx *int, block *ast.BlockStmt, cm CommentMap) {
	for _, stmt := range block.Statements {
		stmtLine := stmt.Pos().Line
		*idx = attachLeadingComments(comments, *idx, stmtLine, stmt, cm)
		attachCommentsToStmt(comments, idx, stmt, cm)
		*idx = attachTrailingComment(comments, *idx, stmtLine, stmt, cm)
	}
}

func attachCommentsToStmt(comments []Comment, idx *int, stmt ast.Statement, cm CommentMap) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		attachCommentsToBlock(comments, idx, s.Consequence, cm)
		if s.Alternative != nil {
			switch alt := s.Alternative.(type) {
			case *ast.ElseStmt:
				attachCommentsToBlock(comments, idx, alt.Body, cm)
			case *ast.IfStmt:
				attachCommentsToStmt(comments, idx, alt, cm)
			}
		}
	case *ast.ForRangeStmt:
		attachCommentsToBlock(comments, idx, s.Body, cm)
	case *ast.ForNumericStmt:
		attachCommentsToBlock(comments, idx, s.Body, cm)
	case *ast.ForConditionStmt:
		attachCommentsToBlock(comments, idx, s.Body, cm)
	case *ast.SwitchStmt:
		for _, c := range s.Cases {
			attachCommentsToBlock(comments, idx, c.Body, cm)
		}
		if s.Otherwise != nil {
			attachCommentsToBlock(comments, idx, s.Otherwise.Body, cm)
		}
	}
}
