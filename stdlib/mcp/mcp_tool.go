package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Tool(server *mcp.Server, name, description string, schema any, handler ToolHandler) {
	server.AddTool(&mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: schema,
	}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := make(map[string]any)
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

		data, _ := json.Marshal(res)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, nil
	})
}

