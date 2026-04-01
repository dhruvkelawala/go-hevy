package whoop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func LoadToken(path string) (Token, error) {
	var token Token
	if strings.TrimSpace(path) == "" {
		return token, fmt.Errorf("resolve WHOOP token path: path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return token, fmt.Errorf("WHOOP token file not found: %s", path)
		}
		return token, fmt.Errorf("read WHOOP token file: %w", err)
	}
	if err := json.Unmarshal(data, &token); err != nil {
		return token, fmt.Errorf("parse WHOOP token file: %w", err)
	}
	if strings.TrimSpace(token.AccessToken) == "" || strings.TrimSpace(token.RefreshToken) == "" {
		return token, fmt.Errorf("WHOOP token file is missing access_token or refresh_token")
	}
	if token.ExpiresAt.IsZero() {
		return token, fmt.Errorf("WHOOP token file is missing expires_at")
	}
	return token, nil
}

func SaveToken(path string, token Token) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("resolve WHOOP token path: path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create WHOOP token directory: %w", err)
	}
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("encode WHOOP token file: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write WHOOP token file: %w", err)
	}
	return nil
}

func (t Token) Expired(now time.Time, skew time.Duration) bool {
	if t.ExpiresAt.IsZero() {
		return true
	}
	return !now.Add(skew).Before(t.ExpiresAt)
}
