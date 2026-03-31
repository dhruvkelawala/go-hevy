package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dhruvkelawala/go-hevy/internal/api"
	appconfig "github.com/dhruvkelawala/go-hevy/internal/config"
	"github.com/dhruvkelawala/go-hevy/internal/output"
)

func requirePositivePagination(page, pageSize, maxPageSize int) error {
	if page < 1 {
		return fmt.Errorf("page must be 1 or greater")
	}
	if pageSize < 1 || pageSize > maxPageSize {
		return fmt.Errorf("page-size must be between 1 and %d", maxPageSize)
	}
	return nil
}

func readWorkoutRequestFile(path string) (api.CreateWorkoutRequest, error) {
	var payload api.CreateWorkoutRequest
	data, err := os.ReadFile(path)
	if err != nil {
		return payload, fmt.Errorf("read file: %w", err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("parse JSON file: %w", err)
	}
	return payload, nil
}

func formatTimestamp(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return t.Local().Format("2006-01-02 15:04")
}

func formatFloatPtr(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *v)
}

func formatIntPtr(v *int) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *v)
}

func printObject(v any, tableRows [][2]string, compactLines []string) error {
	switch app.outputMode {
	case outputJSON:
		return output.PrintJSON(os.Stdout, v)
	case outputCompact:
		return output.PrintCompact(os.Stdout, compactLines)
	default:
		output.PrintKeyValueTable(os.Stdout, tableRows)
		return nil
	}
}

func configRows() [][2]string {
	path, _ := appconfig.ConfigPath()
	return [][2]string{
		{"config_path", output.ValueOrDash(path)},
		{"api_key", output.ValueOrDash(appconfig.Redact(app.config.EffectiveAPIKey()))},
		{"api_key_source", configSource()},
	}
}

func configSource() string {
	if strings.TrimSpace(os.Getenv(appconfig.EnvAPIKey)) != "" {
		return "environment"
	}
	if app.config != nil && app.config.HasAPIKey() {
		return "config_file"
	}
	return "unset"
}
