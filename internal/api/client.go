package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.hevyapp.com"

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	apiKey     string
	rateLimit  time.Duration
}

func NewClient(apiKey string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("missing API key: run `hevy init` or set GO_HEVY_API_KEY")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		rateLimit:  100 * time.Millisecond,
	}, nil
}

func (c *Client) request(ctx context.Context, method, endpoint string, query map[string]string, body any, out any) error {
	if c.rateLimit > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.rateLimit):
		}
	}

	relative := *c.baseURL
	relative.Path = path.Join(relative.Path, endpoint)

	values := relative.Query()
	for key, value := range query {
		if strings.TrimSpace(value) != "" {
			values.Set(key, value)
		}
	}
	relative.RawQuery = values.Encode()

	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		payload = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, relative.String(), payload)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if len(data) > 0 && json.Unmarshal(data, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("api error (%d): %s", resp.StatusCode, apiErr.Error)
		}
		message := strings.TrimSpace(string(data))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("api error (%d): %s", resp.StatusCode, message)
	}

	if out == nil || len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
