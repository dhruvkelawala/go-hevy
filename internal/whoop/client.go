package whoop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	credentialsPath string
	tokenPath       string
	apiBaseURL      *url.URL
	tokenURL        string
	httpClient      *http.Client
	now             func() time.Time
	refreshSkew     time.Duration
	credentials     *Credentials
	token           *Token
}

type Config struct {
	CredentialsPath string
	TokenPath       string
	APIBaseURL      string
	TokenURL        string
	HTTPClient      *http.Client
	Now             func() time.Time
	RefreshSkew     time.Duration
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func NewClient() (*Client, error) {
	credentialsPath, err := DefaultCredentialsPath()
	if err != nil {
		return nil, err
	}
	tokenPath, err := DefaultTokenPath()
	if err != nil {
		return nil, err
	}
	return NewClientWithConfig(Config{CredentialsPath: credentialsPath, TokenPath: tokenPath})
}

func NewClientWithConfig(cfg Config) (*Client, error) {
	apiBaseURL := strings.TrimSpace(cfg.APIBaseURL)
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
	}
	parsedBaseURL, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse WHOOP API base URL: %w", err)
	}
	tokenURL := strings.TrimSpace(cfg.TokenURL)
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	refreshSkew := cfg.RefreshSkew
	if refreshSkew == 0 {
		refreshSkew = time.Minute
	}
	return &Client{
		credentialsPath: cfg.CredentialsPath,
		tokenPath:       cfg.TokenPath,
		apiBaseURL:      parsedBaseURL,
		tokenURL:        tokenURL,
		httpClient:      httpClient,
		now:             now,
		refreshSkew:     refreshSkew,
	}, nil
}

func (c *Client) ListRecoveries(ctx context.Context, days int) (*RecoveryCollection, error) {
	if days <= 0 {
		return nil, fmt.Errorf("WHOOP recovery days must be greater than zero")
	}
	combined := &RecoveryCollection{Records: make([]RecoveryRecord, 0, days)}
	nextToken := ""
	remaining := days
	for remaining > 0 {
		limit := remaining
		if limit > 25 {
			limit = 25
		}
		query := map[string]string{"limit": strconv.Itoa(limit)}
		if nextToken != "" {
			query["nextToken"] = nextToken
		}
		var page RecoveryCollection
		if err := c.get(ctx, "/v2/recovery", query, &page); err != nil {
			return nil, err
		}
		combined.Records = append(combined.Records, page.Records...)
		combined.NextToken = page.NextToken
		remaining = days - len(combined.Records)
		if page.NextToken == "" || len(page.Records) == 0 {
			break
		}
		nextToken = page.NextToken
	}
	if len(combined.Records) > days {
		combined.Records = combined.Records[:days]
	}
	return combined, nil
}

func (c *Client) get(ctx context.Context, endpoint string, query map[string]string, out any) error {
	token, err := c.ensureValidToken(ctx)
	if err != nil {
		return err
	}
	statusCode, body, err := c.doRequest(ctx, http.MethodGet, endpoint, query, nil, token.AccessToken)
	if err != nil {
		return err
	}
	if statusCode == http.StatusUnauthorized {
		refreshedToken, refreshErr := c.refreshAccessToken(ctx)
		if refreshErr != nil {
			return refreshErr
		}
		statusCode, body, err = c.doRequest(ctx, http.MethodGet, endpoint, query, nil, refreshedToken.AccessToken)
		if err != nil {
			return err
		}
	}
	if statusCode < 200 || statusCode >= 300 {
		return formatAPIError(statusCode, body)
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode WHOOP response: %w", err)
	}
	return nil
}

func (c *Client) ensureValidToken(ctx context.Context) (*Token, error) {
	if err := c.loadCredentials(); err != nil {
		return nil, err
	}
	if c.token == nil {
		token, err := LoadToken(c.tokenPath)
		if err != nil {
			return nil, err
		}
		c.token = &token
	}
	if c.token.Expired(c.now().UTC(), c.refreshSkew) {
		return c.refreshAccessToken(ctx)
	}
	return c.token, nil
}

func (c *Client) refreshAccessToken(ctx context.Context) (*Token, error) {
	if err := c.loadCredentials(); err != nil {
		return nil, err
	}
	if c.token == nil {
		token, err := LoadToken(c.tokenPath)
		if err != nil {
			return nil, err
		}
		c.token = &token
	}
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", c.token.RefreshToken)
	values.Set("client_id", c.credentials.ClientID)
	values.Set("client_secret", c.credentials.ClientSecret)
	statusCode, body, err := c.doTokenRequest(ctx, values)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("refresh WHOOP token: %w", formatAPIError(statusCode, body))
	}
	var refreshed refreshResponse
	if err := json.Unmarshal(body, &refreshed); err != nil {
		return nil, fmt.Errorf("decode WHOOP token refresh response: %w", err)
	}
	if strings.TrimSpace(refreshed.AccessToken) == "" {
		return nil, fmt.Errorf("refresh WHOOP token: response missing access_token")
	}
	if strings.TrimSpace(refreshed.RefreshToken) == "" {
		refreshed.RefreshToken = c.token.RefreshToken
	}
	now := c.now().UTC()
	updated := Token{
		AccessToken:  refreshed.AccessToken,
		RefreshToken: refreshed.RefreshToken,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(time.Duration(refreshed.ExpiresIn) * time.Second),
	}
	if refreshed.ExpiresIn <= 0 {
		return nil, fmt.Errorf("refresh WHOOP token: response missing expires_in")
	}
	if err := SaveToken(c.tokenPath, updated); err != nil {
		return nil, err
	}
	c.token = &updated
	return c.token, nil
}

func (c *Client) loadCredentials() error {
	if c.credentials != nil {
		return nil
	}
	credentials, err := LoadCredentials(c.credentialsPath)
	if err != nil {
		return err
	}
	c.credentials = &credentials
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, query map[string]string, body io.Reader, accessToken string) (int, []byte, error) {
	relative := *c.apiBaseURL
	relative.Path = path.Join(relative.Path, endpoint)
	values := relative.Query()
	for key, value := range query {
		if strings.TrimSpace(value) != "" {
			values.Set(key, value)
		}
	}
	relative.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, method, relative.String(), body)
	if err != nil {
		return 0, nil, fmt.Errorf("create WHOOP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("send WHOOP request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("read WHOOP response: %w", err)
	}
	return resp.StatusCode, data, nil
}

func (c *Client) doTokenRequest(ctx context.Context, values url.Values) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return 0, nil, fmt.Errorf("create WHOOP token refresh request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("send WHOOP token refresh request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("read WHOOP token refresh response: %w", err)
	}
	return resp.StatusCode, data, nil
}

func formatAPIError(statusCode int, body []byte) error {
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return fmt.Errorf("WHOOP API error (%d): %s", statusCode, message)
}
