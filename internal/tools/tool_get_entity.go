// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolGetEntity gets the state of a Home Assistant entity
func (tm *ToolsManager) HandleToolGetEntity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	state, err := tm.dependencies.HassClient.GetEntityState(entityID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting entity state: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Check if detailed view is requested
	detailed, _ := args["detailed"].(bool)

	var result map[string]interface{}
	if detailed {
		// Return all fields
		result = map[string]interface{}{
			"entity_id":    state.EntityID,
			"state":        state.State,
			"attributes":   state.Attributes,
			"last_changed": state.LastChanged,
			"last_updated": state.LastUpdated,
		}
	} else {
		// Return lean format with essential fields
		result = map[string]interface{}{
			"entity_id": state.EntityID,
			"state":     state.State,
		}
		if friendlyName, ok := state.Attributes["friendly_name"]; ok {
			result["friendly_name"] = friendlyName
		}

		// Add domain-specific important attributes
		domain := getDomain(entityID)
		importantAttrs := getDomainImportantAttributes(domain)
		for _, attr := range importantAttrs {
			if val, ok := state.Attributes[attr]; ok {
				if result["attributes"] == nil {
					result["attributes"] = make(map[string]interface{})
				}
				result["attributes"].(map[string]interface{})[attr] = val
			}
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling response: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonBytes),
			},
		},
	}, nil
}
