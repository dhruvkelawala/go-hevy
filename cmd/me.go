package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show authenticated user profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		profile, err := client.GetMe(contextForCommand(cmd))
		if err != nil {
			return err
		}

		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, profile)
		case outputCompact:
			parts := make([]string, 0, len(profile))
			keys := make([]string, 0, len(profile))
			for key := range profile {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				parts = append(parts, fmt.Sprintf("%s=%v", key, profile[key]))
			}
			return output.PrintCompact(os.Stdout, []string{strings.Join(parts, " ")})
		default:
			rows := make([][2]string, 0, len(profile))
			keys := make([]string, 0, len(profile))
			for key := range profile {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				rows = append(rows, [2]string{key, fmt.Sprintf("%v", profile[key])})
			}
			output.PrintKeyValueTable(os.Stdout, rows)
			return nil
		}
	},
}
