package granola

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const apiClientTimeout = 30 * time.Second

// Client wraps the Granola public API.
// This is an infrastructure concern — the domain has no knowledge of HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

func NewClient(baseURL string, httpClient *http.Client, token string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: apiClientTimeout}
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		token:      token,
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

// ListNotes calls GET /v1/notes with optional cursor pagination.
func (c *Client) ListNotes(ctx context.Context, createdAfter *time.Time, cursor string, pageSize int) (*NoteListResponse, error) {
	params := url.Values{}
	if createdAfter != nil {
		params.Set("created_after", createdAfter.Format(time.RFC3339))
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if pageSize > 0 {
		params.Set("page_size", strconv.Itoa(pageSize))
	}

	var resp NoteListResponse
	if err := c.get(ctx, "/v1/notes", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetNote calls GET /v1/notes/{id} with optional transcript inclusion.
func (c *Client) GetNote(ctx context.Context, id string, includeTranscript bool) (*NoteDetailResponse, error) {
	params := url.Values{}
	if includeTranscript {
		params.Set("include", "transcript")
	}

	var resp NoteDetailResponse
	if err := c.get(ctx, "/v1/notes/"+url.PathEscape(id), params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, target interface{}) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimited
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}
