package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolSystemOverview gets a comprehensive overview of the Home Assistant system
func (tm *ToolsManager) HandleToolSystemOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	allEntities, err := tm.dependencies.HassClient.GetAllStates()
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

	// Initialize overview structure
	overview := map[string]interface{}{
		"total_entities":    len(allEntities),
		"domains":           make(map[string]interface{}),
		"domain_samples":    make(map[string][]map[string]interface{}),
		"domain_attributes": make(map[string][]string),
		"area_distribution": make(map[string]map[string]int),
	}

	// Group entities by domain
	domainEntities := make(map[string][]map[string]interface{})
	for _, entity := range allEntities {
		domain := getDomain(entity.EntityID)
		entityMap := map[string]interface{}{
			"entity_id":  entity.EntityID,
			"state":      entity.State,
			"attributes": entity.Attributes,
		}
		domainEntities[domain] = append(domainEntities[domain], entityMap)
	}

	// Process each domain
	for domain, entities := range domainEntities {
		count := len(entities)

		// Collect state distribution
		stateDistribution := make(map[string]int)
		for _, entity := range entities {
			state := entity["state"].(string)
			stateDistribution[state]++
		}

		// Store domain information
		overview["domains"].(map[string]interface{})[domain] = map[string]interface{}{
			"count":  count,
			"states": stateDistribution,
		}

		// Select representative samples (up to 3 per domain)
		sampleLimit := 3
		if count < sampleLimit {
			sampleLimit = count
		}
		samples := make([]map[string]interface{}, 0, sampleLimit)
		for i := 0; i < sampleLimit; i++ {
			entity := entities[i]
			sample := map[string]interface{}{
				"entity_id":     entity["entity_id"],
				"state":         entity["state"],
				"friendly_name": entity["entity_id"],
			}
			if attrs, ok := entity["attributes"].(map[string]interface{}); ok {
				if friendlyName, ok := attrs["friendly_name"].(string); ok {
					sample["friendly_name"] = friendlyName
				}
			}
			samples = append(samples, sample)
		}
		overview["domain_samples"].(map[string][]map[string]interface{})[domain] = samples

		// Collect common attributes for this domain
		attributeCounts := make(map[string]int)
		for _, entity := range entities {
			if attrs, ok := entity["attributes"].(map[string]interface{}); ok {
				for attr := range attrs {
					attributeCounts[attr]++
				}
			}
		}

		// Get top 5 most common attributes
		type attrCount struct {
			name  string
			count int
		}
		var attrList []attrCount
		for name, count := range attributeCounts {
			attrList = append(attrList, attrCount{name, count})
		}
		// Sort by count descending
		for i := 0; i < len(attrList); i++ {
			for j := i + 1; j < len(attrList); j++ {
				if attrList[j].count > attrList[i].count {
					attrList[i], attrList[j] = attrList[j], attrList[i]
				}
			}
		}
		topAttrs := make([]string, 0, 5)
		for i := 0; i < len(attrList) && i < 5; i++ {
			topAttrs = append(topAttrs, attrList[i].name)
		}
		overview["domain_attributes"].(map[string][]string)[domain] = topAttrs

		// Group by area if available
		areaDistribution := overview["area_distribution"].(map[string]map[string]int)
		for _, entity := range entities {
			areaName := "Unknown"
			if attrs, ok := entity["attributes"].(map[string]interface{}); ok {
				if area, ok := attrs["area_name"].(string); ok && area != "" {
					areaName = area
				} else if areaID, ok := attrs["area_id"].(string); ok && areaID != "" {
					areaName = areaID
				}
			}

			if _, ok := areaDistribution[areaName]; !ok {
				areaDistribution[areaName] = make(map[string]int)
			}
			areaDistribution[areaName][domain]++
		}
	}

	// Add summary information
	overview["domain_count"] = len(domainEntities)

	// Get most common domains
	type domainCount struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	var mostCommon []domainCount
	for domain, entities := range domainEntities {
		mostCommon = append(mostCommon, domainCount{domain, len(entities)})
	}
	// Sort by count descending
	for i := 0; i < len(mostCommon); i++ {
		for j := i + 1; j < len(mostCommon); j++ {
			if mostCommon[j].Count > mostCommon[i].Count {
				mostCommon[i], mostCommon[j] = mostCommon[j], mostCommon[i]
			}
		}
	}
	if len(mostCommon) > 5 {
		mostCommon = mostCommon[:5]
	}
	overview["most_common_domains"] = mostCommon

	jsonBytes, err := json.MarshalIndent(overview, "", "  ")
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
