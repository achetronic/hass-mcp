package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolSearchEntities searches for entities matching a query string
func (tm *ToolsManager) HandleToolSearchEntities(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	query, _ := args["query"].(string)

	// Get limit (default 20)
	limit := 20
	if limitVal, ok := args["limit"].(float64); ok {
		limit = int(limitVal)
	}

	// Handle empty or wildcard query
	if query == "" || query == "*" {
		query = ""
	}

	states, err := tm.dependencies.HassClient.SearchEntities(query, limit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error searching entities: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Build structured response
	domainsCount := make(map[string]int)
	var simplifiedEntities []map[string]interface{}

	for _, state := range states {
		domain := getDomain(state.EntityID)
		domainsCount[domain]++

		simplified := map[string]interface{}{
			"entity_id":     state.EntityID,
			"state":         state.State,
			"domain":        domain,
			"friendly_name": state.EntityID,
		}

		if friendlyName, ok := state.Attributes["friendly_name"].(string); ok {
			simplified["friendly_name"] = friendlyName
		}

		// Add domain-specific important attributes
		switch domain {
		case "light":
			if brightness, ok := state.Attributes["brightness"]; ok {
				simplified["brightness"] = brightness
			}
		case "sensor":
			if unit, ok := state.Attributes["unit_of_measurement"]; ok {
				simplified["unit"] = unit
			}
		case "climate":
			if temp, ok := state.Attributes["temperature"]; ok {
				simplified["temperature"] = temp
			}
		case "media_player":
			if title, ok := state.Attributes["media_title"]; ok {
				simplified["media_title"] = title
			}
		}

		simplifiedEntities = append(simplifiedEntities, simplified)
	}

	queryDesc := query
	if query == "" {
		queryDesc = "all entities (no filtering)"
	}

	result := map[string]interface{}{
		"count":   len(simplifiedEntities),
		"results": simplifiedEntities,
		"domains": domainsCount,
		"query":   queryDesc,
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

// getDomain extracts the domain from an entity ID
func getDomain(entityID string) string {
	parts := strings.Split(entityID, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// getDomainImportantAttributes returns important attributes for a domain
func getDomainImportantAttributes(domain string) []string {
	attributes := map[string][]string{
		"light":         {"brightness", "color_temp", "rgb_color", "supported_color_modes"},
		"switch":        {"device_class"},
		"binary_sensor": {"device_class"},
		"sensor":        {"device_class", "unit_of_measurement", "state_class"},
		"climate":       {"hvac_mode", "current_temperature", "temperature", "hvac_action"},
		"media_player":  {"media_title", "media_artist", "source", "volume_level"},
		"cover":         {"current_position", "current_tilt_position"},
		"fan":           {"percentage", "preset_mode"},
		"camera":        {"entity_picture"},
		"automation":    {"last_triggered"},
		"script":        {"last_triggered"},
	}

	if attrs, ok := attributes[domain]; ok {
		return attrs
	}
	return nil
}
