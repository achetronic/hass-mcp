package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolGetHistory gets the history of an entity's state changes
func (tm *ToolsManager) HandleToolGetHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Get hours (default 24)
	hours := 24
	if hoursVal, ok := args["hours"].(float64); ok {
		hours = int(hoursVal)
	}

	history, err := tm.dependencies.HassClient.GetEntityHistory(entityID, hours)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting history: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	if len(history) == 0 {
		result := map[string]interface{}{
			"entity_id":     entityID,
			"states":        []interface{}{},
			"count":         0,
			"first_changed": nil,
			"last_changed":  nil,
			"note":          "No state changes found in the specified timeframe.",
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(jsonBytes),
				},
			},
		}, nil
	}

	// Sort by last_changed
	sort.Slice(history, func(i, j int) bool {
		return history[i].LastChanged < history[j].LastChanged
	})

	// Convert to generic maps for JSON output
	var states []map[string]interface{}
	for _, h := range history {
		state := map[string]interface{}{
			"state":        h.State,
			"last_changed": h.LastChanged,
			"last_updated": h.LastUpdated,
		}
		states = append(states, state)
	}

	result := map[string]interface{}{
		"entity_id":     entityID,
		"states":        states,
		"count":         len(history),
		"first_changed": history[0].LastChanged,
		"last_changed":  history[len(history)-1].LastChanged,
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
