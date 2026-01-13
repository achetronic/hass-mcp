package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolDomainSummary gets a summary of entities in a specific domain
func (tm *ToolsManager) HandleToolDomainSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Get example limit (default 3)
	exampleLimit := 3
	if limitVal, ok := args["example_limit"].(float64); ok {
		exampleLimit = int(limitVal)
	}

	entities, err := tm.dependencies.HassClient.GetEntitiesByDomain(domain)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting entities: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Build summary
	totalCount := len(entities)
	stateCounts := make(map[string]int)
	stateExamples := make(map[string][]map[string]interface{})
	attributesSummary := make(map[string]int)

	for _, entity := range entities {
		state := entity.State

		// Count states
		stateCounts[state]++

		// Add examples (up to the limit)
		if len(stateExamples[state]) < exampleLimit {
			example := map[string]interface{}{
				"entity_id":     entity.EntityID,
				"friendly_name": entity.EntityID,
			}
			if friendlyName, ok := entity.Attributes["friendly_name"].(string); ok {
				example["friendly_name"] = friendlyName
			}
			stateExamples[state] = append(stateExamples[state], example)
		}

		// Collect attribute keys
		for attrKey := range entity.Attributes {
			attributesSummary[attrKey]++
		}
	}

	// Get top 10 most common attributes
	type attrCount struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	var commonAttributes []attrCount
	for name, count := range attributesSummary {
		commonAttributes = append(commonAttributes, attrCount{name, count})
	}
	// Sort by count descending (simple bubble sort for small data)
	for i := 0; i < len(commonAttributes); i++ {
		for j := i + 1; j < len(commonAttributes); j++ {
			if commonAttributes[j].Count > commonAttributes[i].Count {
				commonAttributes[i], commonAttributes[j] = commonAttributes[j], commonAttributes[i]
			}
		}
	}
	if len(commonAttributes) > 10 {
		commonAttributes = commonAttributes[:10]
	}

	summary := map[string]interface{}{
		"domain":             domain,
		"total_count":        totalCount,
		"state_distribution": stateCounts,
		"examples":           stateExamples,
		"common_attributes":  commonAttributes,
	}

	jsonBytes, err := json.MarshalIndent(summary, "", "  ")
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
