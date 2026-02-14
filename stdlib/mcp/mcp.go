package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolHandler is a function that handles an MCP tool call.
// It receives arguments as a map and returns either a string or a *mcp.CallToolResult.
type ToolHandler func(args map[string]any) (any, error)

// New creates a new MCP server with the given name and version.
func New(name, version string) *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)
}

// Serve runs the server on the stdio transport. This is a blocking call.
func Serve(s *mcp.Server) error {
	return s.Run(context.Background(), &mcp.StdioTransport{})
}

// Tool registers a tool with the server.
func Tool(s *mcp.Server, name, description string, schema any, handler ToolHandler) {
	s.AddTool(&mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: schema,
	}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args map[string]any
		if len(req.Params.Arguments) > 0 {
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
		}

		res, err := handler(args)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
				IsError: true,
			}, nil
		}

		if r, ok := res.(*mcp.CallToolResult); ok {
			return r, nil
		}

		if s, ok := res.(string); ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: s},
				},
			}, nil
		}

		// Fallback for other types: marshal to JSON string
		data, _ := json.Marshal(res)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil
	})
}

// SchemaProperty represents a property in a JSON schema.
type SchemaProperty struct {
	Name        string
	Type        string
	Description string
}

// Prop creates a new SchemaProperty.
func Prop(name, typ, description string) SchemaProperty {
	return SchemaProperty{Name: name, Type: typ, Description: description}
}

// Schema creates a JSON schema object from a list of properties.
func Schema(props ...SchemaProperty) map[string]any {
	properties := make(map[string]any)
	for _, p := range props {
		properties[p.Name] = map[string]any{
			"type":        p.Type,
			"description": p.Description,
		}
	}
	return map[string]any{
		"type":       "object",
		"properties": properties,
	}
}

// Required adds a list of required property names to a schema.
func Required(schema any, names ...string) any {
	s, ok := schema.(map[string]any)
	if !ok {
		return schema
	}
	s["required"] = names
	return s
}

// TextResult creates a successful tool result with text content.
func TextResult(text string) any {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// ErrorResult creates an error tool result.
func ErrorResult(msg string) any {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}
