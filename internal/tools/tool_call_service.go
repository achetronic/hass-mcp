package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolCallService calls any Home Assistant service
func (tm *ToolsManager) HandleToolCallService(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	domain, ok := args["domain"].(string)
	if !ok || domain == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "Error: domain parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	service, ok := args["service"].(string)
	if !ok || service == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "Error: service parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	// Get optional data
	var data map[string]interface{}
	if dataVal, ok := args["data"].(map[string]interface{}); ok {
		data = dataVal
	}

	result, err := tm.dependencies.HassClient.CallService(domain, service, data)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error calling service: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Try to parse the result as JSON
	var response interface{}
	if len(result) > 0 {
		if err := json.Unmarshal(result, &response); err != nil {
			response = string(result)
		}
	} else {
		response = map[string]interface{}{
			"success": true,
			"domain":  domain,
			"service": service,
		}
	}

	jsonBytes, _ := json.MarshalIndent(response, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonBytes),
			},
		},
	}, nil
}
