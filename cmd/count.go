package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Show total workout count",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		resp, err := client.GetWorkoutCount(contextForCommand(cmd))
		if err != nil {
			return err
		}

		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, resp)
		case outputCompact:
			return output.PrintCompact(os.Stdout, []string{fmt.Sprintf("workout_count=%d", resp.WorkoutCount)})
		default:
			output.PrintKeyValueTable(os.Stdout, [][2]string{{"workout_count", fmt.Sprintf("%d", resp.WorkoutCount)}})
			return nil
		}
	},
}
