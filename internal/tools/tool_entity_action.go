package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolEntityAction performs an action on a Home Assistant entity (on, off, toggle)
func (tm *ToolsManager) HandleToolEntityAction(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	entityID, ok := args["entity_id"].(string)
	if !ok || entityID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "Error: entity_id parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	action, ok := args["action"].(string)
	if !ok || action == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "Error: action parameter is required",
				},
			},
			IsError: true,
		}, nil
	}

	// Validate action
	if action != "on" && action != "off" && action != "toggle" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error: Invalid action '%s'. Valid actions are 'on', 'off', 'toggle'", action),
				},
			},
			IsError: true,
		}, nil
	}

	// Map action to service name
	service := action
	if action != "toggle" {
		service = "turn_" + action
	}

	// Extract the domain from the entity_id
	domain := getDomain(entityID)

	// Prepare service data
	data := map[string]interface{}{
		"entity_id": entityID,
	}

	// Add optional params if provided
	if params, ok := args["params"].(map[string]interface{}); ok {
		for k, v := range params {
			data[k] = v
		}
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

	// Try to parse the result
	var response interface{}
	if len(result) > 0 {
		if err := json.Unmarshal(result, &response); err != nil {
			response = string(result)
		}
	} else {
		response = map[string]interface{}{
			"success":   true,
			"entity_id": entityID,
			"action":    action,
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
