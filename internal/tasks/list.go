package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	TasksAPI = "https://api.todoist.com/api/v1/tasks"
)

type Task struct {
	UserID         string                 `json:"user_id"`
	ID             string                 `json:"id"`
	ProjectID      string                 `json:"project_id"`
	SectionID      string                 `json:"section_id"`
	ParentID       string                 `json:"parent_id"`
	AddedByUID     string                 `json:"added_by_uid"`
	AssignedByUID  string                 `json:"assigned_by_uid"`
	ResponsibleUID string                 `json:"responsible_uid"`
	Labels         []string               `json:"labels"`
	Deadline       map[string]interface{} `json:"deadline"`
	Duration       map[string]interface{} `json:"duration"`
	Checked        bool                   `json:"checked"`
	IsDeleted      bool                   `json:"is_deleted"`
	AddedAt        string                 `json:"added_at"`
	CompletedAt    string                 `json:"completed_at"`
	UpdatedAt      string                 `json:"updated_at"`
	Due            map[string]interface{} `json:"due"`
	Priority       int                    `json:"priority"`
	ChildOrder     int                    `json:"child_order"`
	Content        string                 `json:"content"`
	Description    string                 `json:"description"`
	NoteCount      int                    `json:"note_count"`
	DayOrder       int                    `json:"day_order"`
	IsCollapsed    bool                   `json:"is_collapsed"`
}

type TasksResponse struct {
	Results    []Task `json:"results"`
	NextCursor string `json:"next_cursor"`
}

type ListOptions struct {
	Cursor string
	Limit  int
}

func ListTasks(accessToken string, options *ListOptions) (*TasksResponse, error) {
	// Build query parameters
	params := url.Values{}

	if options != nil {
		if options.Cursor != "" {
			params.Set("cursor", options.Cursor)
		}
		if options.Limit > 0 {
			params.Set("limit", strconv.Itoa(options.Limit))
		}
	}

	// Build URL
	apiURL := TasksAPI
	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Handle specific status codes
	switch resp.StatusCode {
	case 401:
		return nil, fmt.Errorf("unauthorized: please login again")
	case 200:
		// Success, continue
	default:
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var tasksResp TasksResponse
	if err := json.Unmarshal(body, &tasksResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tasksResp, nil
}
