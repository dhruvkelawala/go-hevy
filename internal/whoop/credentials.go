package whoop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultAPIBaseURL = "https://api.prod.whoop.com/developer"
	defaultTokenURL   = "https://api.prod.whoop.com/oauth/oauth2/token"
)

func DefaultCredentialsPath() (string, error) {
	dir, err := defaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

func DefaultTokenPath() (string, error) {
	dir, err := defaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

func LoadCredentials(path string) (Credentials, error) {
	var credentials Credentials
	if strings.TrimSpace(path) == "" {
		return credentials, fmt.Errorf("resolve WHOOP credentials path: path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return credentials, fmt.Errorf("WHOOP credentials file not found: %s", path)
		}
		return credentials, fmt.Errorf("read WHOOP credentials file: %w", err)
	}
	if err := json.Unmarshal(data, &credentials); err != nil {
		return credentials, fmt.Errorf("parse WHOOP credentials file: %w", err)
	}
	if strings.TrimSpace(credentials.ClientID) == "" || strings.TrimSpace(credentials.ClientSecret) == "" {
		return credentials, fmt.Errorf("WHOOP credentials file is missing client_id or client_secret")
	}
	return credentials, nil
}

func defaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".whoop"), nil
}
