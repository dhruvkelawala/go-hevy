package cmd

import (
	"fmt"
	"os"

	appconfig "github.com/dhruvkelawala/go-hevy/internal/config"
	"github.com/dhruvkelawala/go-hevy/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or update configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, map[string]any{
				"config_path":    configRows()[0][1],
				"api_key":        appconfig.Redact(app.config.EffectiveAPIKey()),
				"api_key_source": configSource(),
			})
		case outputCompact:
			return output.PrintCompact(os.Stdout, []string{
				fmt.Sprintf("config_path=%s", configRows()[0][1]),
				fmt.Sprintf("api_key=%s", appconfig.Redact(app.config.EffectiveAPIKey())),
				fmt.Sprintf("api_key_source=%s", configSource()),
			})
		default:
			output.PrintKeyValueTable(os.Stdout, configRows())
			return nil
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set key <value>",
	Short: "Set the stored API key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "key" {
			return fmt.Errorf("unsupported config key %q", args[0])
		}
		cfg := &appconfig.Config{APIKey: args[1]}
		if err := appconfig.Save(cfg); err != nil {
			return err
		}
		app.config = cfg
		printLine("Updated config key %q", args[0])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
