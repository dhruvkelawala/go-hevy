package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/go-hevy/internal/output"
	"github.com/spf13/cobra"
)

var historyStartDate string
var historyEndDate string

var historyCmd = &cobra.Command{
	Use:   "history <exercise-id>",
	Short: "Show exercise history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		resp, err := client.GetExerciseHistory(contextForCommand(cmd), args[0], historyStartDate, historyEndDate)
		if err != nil {
			return err
		}
		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, resp)
		case outputCompact:
			lines := make([]string, 0, len(resp.ExerciseHistory))
			for _, entry := range resp.ExerciseHistory {
				lines = append(lines, fmt.Sprintf("%s | %s | weight=%s | reps=%s", entry.WorkoutID, entry.WorkoutTitle, formatFloatPtr(entry.WeightKG), formatIntPtr(entry.Reps)))
			}
			return output.PrintCompact(os.Stdout, lines)
		default:
			rows := make([][]string, 0, len(resp.ExerciseHistory))
			for _, entry := range resp.ExerciseHistory {
				rows = append(rows, []string{entry.WorkoutID, entry.WorkoutTitle, formatTimestamp(entry.WorkoutStartTime), entry.SetType, formatFloatPtr(entry.WeightKG), formatIntPtr(entry.Reps), formatFloatPtr(entry.RPE)})
			}
			output.PrintTable(os.Stdout, []string{"Workout ID", "Workout", "Start", "Set Type", "Weight KG", "Reps", "RPE"}, rows)
			return nil
		}
	},
}

func init() {
	historyCmd.Flags().StringVar(&historyStartDate, "start-date", "", "Filter history from ISO-8601 timestamp")
	historyCmd.Flags().StringVar(&historyEndDate, "end-date", "", "Filter history to ISO-8601 timestamp")
}
