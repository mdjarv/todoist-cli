package todoist

import (
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

type ListTasksOptions struct {
	Cursor string
	Limit  int
}

// Project types
type Project struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CommentCount  int    `json:"comment_count"`
	Order         int    `json:"order"`
	Color         string `json:"color"`
	Shared        bool   `json:"shared"`
	SyncID        int    `json:"sync_id"`
	FavoriteDelim string `json:"favorite_delim"`
	Favorite      bool   `json:"favorite"`
	InboxProject  bool   `json:"inbox_project"`
	TeamInbox     bool   `json:"team_inbox"`
	ViewStyle     string `json:"view_style"`
	URL           string `json:"url"`
	ParentID      string `json:"parent_id"`
}

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
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

	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ProjectsResponse{Projects: projects}, nil
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
