package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	BaseURL  = "https://api.todoist.com/api/v1"
	TokenURL = "https://todoist.com/oauth/access_token"
)

type Client interface {
	ListTasks(ctx context.Context, options *ListTasksOptions) (*TasksResponse, error)
	ListProjects(ctx context.Context) (*ProjectsResponse, error)
	ExchangeCodeForToken(code, redirectURI, clientID, clientSecret string) (*TokenResponse, error)
	CloseTask(ctx context.Context, taskID string) error
	ReopenTask(ctx context.Context, taskID string) error
	CreateTask(ctx context.Context, options CreateTaskOptions) error
}

type client struct {
	httpClient  *http.Client
	accessToken string
}

func NewClient(accessToken string) Client {
	return &client{
		httpClient:  &http.Client{},
		accessToken: accessToken,
	}
}

// Task types
type Task struct {
	UserID         string         `json:"user_id"`
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	SectionID      string         `json:"section_id"`
	ParentID       string         `json:"parent_id"`
	AddedByUID     string         `json:"added_by_uid"`
	AssignedByUID  string         `json:"assigned_by_uid"`
	ResponsibleUID string         `json:"responsible_uid"`
	Labels         []string       `json:"labels"`
	Deadline       map[string]any `json:"deadline"`
	Duration       map[string]any `json:"duration"`
	Checked        bool           `json:"checked"`
	IsDeleted      bool           `json:"is_deleted"`
	AddedAt        string         `json:"added_at"`
	CompletedAt    string         `json:"completed_at"`
	UpdatedAt      string         `json:"updated_at"`
	Due            map[string]any `json:"due"`
	Priority       int            `json:"priority"`
	ChildOrder     int            `json:"child_order"`
	Content        string         `json:"content"`
	Description    string         `json:"description"`
	NoteCount      int            `json:"note_count"`
	DayOrder       int            `json:"day_order"`
	IsCollapsed    bool           `json:"is_collapsed"`
}

type TasksResponse struct {
	Results    []Task `json:"results"`
	NextCursor string `json:"next_cursor"`
}

type ListTasksOptions struct {
	Cursor string
	Limit  int
}

type CreateTaskOptions struct {
	Content      string   `json:"content"`
	Description  string   `json:"description,omitempty"`
	ProjectID    string   `json:"project_id,omitempty"`
	SectionID    string   `json:"section_id,omitempty"`
	ParentID     string   `json:"parent_id,omitempty"`
	Order        int      `json:"order,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Priority     int      `json:"priority,omitempty"`
	AssigneeID   int      `json:"assignee_id,omitempty"`
	DueDate      string   `json:"due_date,omitempty"`
	DueString    string   `json:"due_string,omitempty"`
	DueDatetime  string   `json:"due_datetime,omitempty"`
	DueLang      string   `json:"due_lang,omitempty"`
	Duration     int      `json:"duration,omitempty"`
	DurationUnit string   `json:"duration_unit,omitempty"`
	DeadlineDate string   `json:"deadline_date,omitempty"`
	DeadlineLang string   `json:"deadline_lang,omitempty"`
}

// Project types
type Project struct {
	ID             string         `json:"id"`
	CanAssignTasks bool           `json:"can_assign_tasks"`
	ChildOrder     int            `json:"child_order"`
	Color          string         `json:"color"`
	CreatorUID     string         `json:"creator_uid"`
	CreatedAt      string         `json:"created_at"`
	IsArchived     bool           `json:"is_archived"`
	IsDeleted      bool           `json:"is_deleted"`
	IsFavorite     bool           `json:"is_favorite"`
	IsFrozen       bool           `json:"is_frozen"`
	Name           string         `json:"name"`
	UpdatedAt      string         `json:"updated_at"`
	ViewStyle      string         `json:"view_style"`
	DefaultOrder   int            `json:"default_order"`
	Description    string         `json:"description"`
	PublicKey      string         `json:"public_key"`
	Access         map[string]any `json:"access"`
	Role           string         `json:"role"`
	ParentID       string         `json:"parent_id"`
	InboxProject   bool           `json:"inbox_project"`
	IsCollapsed    bool           `json:"is_collapsed"`
	IsShared       bool           `json:"is_shared"`
}

type ProjectsResponse struct {
	Results    []Project `json:"results"`
	NextCursor string    `json:"next_cursor"`
}

// OAuth types
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error,omitempty"`
}

// Task methods
func (c *client) ListTasks(ctx context.Context, options *ListTasksOptions) (*TasksResponse, error) {
	params := url.Values{}
	if options != nil {
		if options.Cursor != "" {
			params.Set("cursor", options.Cursor)
		}
		if options.Limit > 0 {
			params.Set("limit", strconv.Itoa(options.Limit))
		}
	}

	apiURL := BaseURL + "/tasks"
	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401:
		return nil, fmt.Errorf("unauthorized: please login again")
	case 200:
		// Success, continue
	default:
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tasksResp TasksResponse
	if err := json.Unmarshal(body, &tasksResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tasksResp, nil
}

func (c *client) CloseTask(ctx context.Context, taskID string) error {
	apiURL := BaseURL + "/tasks/" + taskID + "/close"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API error: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func (c *client) ReopenTask(ctx context.Context, taskID string) error {
	apiURL := BaseURL + "/tasks/" + taskID + "/reopen"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API error: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func (c *client) CreateTask(ctx context.Context, options CreateTaskOptions) error {
	apiURL := BaseURL + "/tasks"

	requestBody, err := json.Marshal(options)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d %s: %s", resp.StatusCode, resp.Status, body)
	}

	return nil
}

// Project methods
func (c *client) ListProjects(ctx context.Context) (*ProjectsResponse, error) {
	apiURL := BaseURL + "/projects"

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401:
		return nil, fmt.Errorf("unauthorized: please login again")
	case 200:
		// Success, continue
	default:
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var projResp ProjectsResponse
	if err := json.Unmarshal(body, &projResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &projResp, nil
}

// OAuth methods
func (c *client) ExchangeCodeForToken(code, redirectURI, clientID, clientSecret string) (*TokenResponse, error) {
	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token exchange error: %s", tokenResp.Error)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token received")
	}

	return &tokenResp, nil
}