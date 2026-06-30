// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolListAutomations lists all Home Assistant automations
func (tm *ToolsManager) HandleToolListAutomations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	automations, err := tm.dependencies.HassClient.GetAutomations()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting automations: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Process automation entities into a cleaner format
	var result []map[string]interface{}
	for _, entity := range automations {
		automation := map[string]interface{}{
			"id":        getDomain(entity.EntityID) + "." + getEntityName(entity.EntityID),
			"entity_id": entity.EntityID,
			"state":     entity.State,
			"alias":     entity.EntityID,
		}

		if friendlyName, ok := entity.Attributes["friendly_name"].(string); ok {
			automation["alias"] = friendlyName
		}

		if lastTriggered, ok := entity.Attributes["last_triggered"]; ok {
			automation["last_triggered"] = lastTriggered
		}

		result = append(result, automation)
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
