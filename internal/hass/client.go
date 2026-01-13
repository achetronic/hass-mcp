package hass

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Home Assistant API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Home Assistant client
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to the Home Assistant API
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetVersion returns the Home Assistant version
func (c *Client) GetVersion() (string, error) {
	body, err := c.doRequest("GET", "/api/config", nil)
	if err != nil {
		return "", err
	}

	var config struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &config); err != nil {
		return "", fmt.Errorf("failed to parse config response: %w", err)
	}

	return config.Version, nil
}

// EntityState represents the state of a Home Assistant entity
type EntityState struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed,omitempty"`
	LastUpdated string                 `json:"last_updated,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// GetEntityState gets the state of a specific entity
func (c *Client) GetEntityState(entityID string) (*EntityState, error) {
	body, err := c.doRequest("GET", "/api/states/"+entityID, nil)
	if err != nil {
		return nil, err
	}

	var state EntityState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse entity state: %w", err)
	}

	return &state, nil
}

// GetAllStates gets all entity states
func (c *Client) GetAllStates() ([]EntityState, error) {
	body, err := c.doRequest("GET", "/api/states", nil)
	if err != nil {
		return nil, err
	}

	var states []EntityState
	if err := json.Unmarshal(body, &states); err != nil {
		return nil, fmt.Errorf("failed to parse states response: %w", err)
	}

	return states, nil
}

// GetEntitiesByDomain gets all entities for a specific domain
func (c *Client) GetEntitiesByDomain(domain string) ([]EntityState, error) {
	allStates, err := c.GetAllStates()
	if err != nil {
		return nil, err
	}

	var filtered []EntityState
	prefix := domain + "."
	for _, state := range allStates {
		if strings.HasPrefix(state.EntityID, prefix) {
			filtered = append(filtered, state)
		}
	}

	return filtered, nil
}

// SearchEntities searches entities by query string
func (c *Client) SearchEntities(query string, limit int) ([]EntityState, error) {
	allStates, err := c.GetAllStates()
	if err != nil {
		return nil, err
	}

	if query == "" || query == "*" {
		if limit > 0 && len(allStates) > limit {
			return allStates[:limit], nil
		}
		return allStates, nil
	}

	queryLower := strings.ToLower(query)
	var filtered []EntityState

	for _, state := range allStates {
		if len(filtered) >= limit && limit > 0 {
			break
		}

		// Search in entity_id
		if strings.Contains(strings.ToLower(state.EntityID), queryLower) {
			filtered = append(filtered, state)
			continue
		}

		// Search in friendly_name
		if friendlyName, ok := state.Attributes["friendly_name"].(string); ok {
			if strings.Contains(strings.ToLower(friendlyName), queryLower) {
				filtered = append(filtered, state)
				continue
			}
		}

		// Search in state
		if strings.Contains(strings.ToLower(state.State), queryLower) {
			filtered = append(filtered, state)
			continue
		}
	}

	return filtered, nil
}

// CallService calls a Home Assistant service
func (c *Client) CallService(domain, service string, data map[string]interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("/api/services/%s/%s", domain, service)
	return c.doRequest("POST", endpoint, data)
}

// GetAutomations gets all automation entities
func (c *Client) GetAutomations() ([]EntityState, error) {
	return c.GetEntitiesByDomain("automation")
}

// RestartHomeAssistant restarts Home Assistant
func (c *Client) RestartHomeAssistant() error {
	_, err := c.CallService("homeassistant", "restart", nil)
	return err
}

// HistoryState represents a historical state entry
type HistoryState struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
}

// GetEntityHistory gets the history of an entity
func (c *Client) GetEntityHistory(entityID string, hours int) ([]HistoryState, error) {
	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	startTimeISO := startTime.Format("2006-01-02T15:04:05Z")
	endTimeISO := endTime.Format("2006-01-02T15:04:05Z")

	params := url.Values{}
	params.Set("filter_entity_id", entityID)
	params.Set("minimal_response", "true")
	params.Set("end_time", endTimeISO)

	endpoint := fmt.Sprintf("/api/history/period/%s?%s", startTimeISO, params.Encode())

	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// The API returns a list of lists
	var historyLists [][]HistoryState
	if err := json.Unmarshal(body, &historyLists); err != nil {
		return nil, fmt.Errorf("failed to parse history response: %w", err)
	}

	// Flatten the lists
	var history []HistoryState
	for _, list := range historyLists {
		history = append(history, list...)
	}

	return history, nil
}

// GetErrorLog gets the Home Assistant error log
func (c *Client) GetErrorLog() (string, error) {
	body, err := c.doRequest("GET", "/api/error_log", nil)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
