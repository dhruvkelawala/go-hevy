package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	appconfig "github.com/dhruvkelawala/hevy-cli/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively configure hevy",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)
		fmt.Fprint(os.Stdout, "Enter your Hevy API key: ")
		apiKey, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read api key: %w", err)
		}
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return fmt.Errorf("api key cannot be empty")
		}

		cfg := &appconfig.Config{APIKey: apiKey}
		if err := appconfig.Save(cfg); err != nil {
			return err
		}

		app.config = cfg
		path, _ := appconfig.ConfigPath()
		color.New(color.FgGreen).Fprintf(os.Stdout, "Saved API key to %s\n", path)
		return nil
	},
}
