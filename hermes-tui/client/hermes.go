package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL   string
	SessionID string
	HTTP      *http.Client
}

type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId"`
}

type ChatResponse struct {
	Reply     string   `json:"reply"`
	ToolsUsed []string `json:"tools_used"`
	Profile   *Profile `json:"profile"`
	Error     string   `json:"error,omitempty"`
}

type Profile struct {
	Name        string   `json:"name"`
	Interests   []string `json:"interests"`
	Goals       []string `json:"goals"`
	Preferences []string `json:"preferences"`
	Facts       []string `json:"facts"`
	Habits      []string `json:"habits"`
	UpdatedAt   string   `json:"updated_at"`
}

type Task struct {
	ID          int    `json:"id"`
	TaskDesc    string `json:"task"`
	Priority    string `json:"priority"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type ProfileResponse struct {
	Profile *Profile     `json:"profile"`
	Tasks   TasksSummary `json:"tasks"`
}

type TasksSummary struct {
	ActiveCount int    `json:"active_count"`
	Tasks       []Task `json:"tasks"`
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

func New(baseURL, sessionID string) *Client {
	return &Client{
		BaseURL:   baseURL,
		SessionID: sessionID,
		HTTP:      &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Chat(message string) (*ChatResponse, error) {
	payload, _ := json.Marshal(ChatRequest{Message: message, SessionID: c.SessionID})
	resp, err := c.HTTP.Post(c.BaseURL+"/chat", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var r ChatResponse
	json.Unmarshal(body, &r)
	if r.Error != "" && r.Reply == "" {
		return nil, fmt.Errorf("%s", r.Error)
	}
	return &r, nil
}

func (c *Client) GetProfile() (*ProfileResponse, error) {
	payload, _ := json.Marshal(map[string]string{"sessionId": c.SessionID})
	resp, err := c.HTTP.Post(c.BaseURL+"/profile", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var r ProfileResponse
	json.Unmarshal(body, &r)
	return &r, nil
}

func (c *Client) GetTasks() (*TasksResponse, error) {
	payload, _ := json.Marshal(map[string]string{"sessionId": c.SessionID})
	resp, err := c.HTTP.Post(c.BaseURL+"/tasks", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var r TasksResponse
	json.Unmarshal(body, &r)
	return &r, nil
}

func (c *Client) ClearHistory() error {
	payload, _ := json.Marshal(map[string]string{"sessionId": c.SessionID})
	resp, err := c.HTTP.Post(c.BaseURL+"/clear", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) HealthCheck() bool {
	resp, err := c.HTTP.Get(c.BaseURL + "/")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}