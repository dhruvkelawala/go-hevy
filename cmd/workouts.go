package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/go-hevy/internal/api"
	"github.com/dhruvkelawala/go-hevy/internal/output"
	"github.com/spf13/cobra"
)

var workoutFile string

var workoutsCmd = &cobra.Command{
	Use:   "workouts",
	Short: "List recent workouts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositivePagination(app.page, app.pageSize, 10); err != nil {
			return err
		}

		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		resp, err := client.ListWorkouts(contextForCommand(cmd), app.page, app.pageSize)
		if err != nil {
			return err
		}

		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, resp)
		case outputCompact:
			lines := make([]string, 0, len(resp.Workouts))
			for _, workout := range resp.Workouts {
				lines = append(lines, fmt.Sprintf("%s | %s | %s", workout.ID, workout.Title, formatTimestamp(workout.StartTime)))
			}
			return output.PrintCompact(os.Stdout, lines)
		default:
			rows := make([][]string, 0, len(resp.Workouts))
			for _, workout := range resp.Workouts {
				rows = append(rows, []string{workout.ID, workout.Title, formatTimestamp(workout.StartTime), formatTimestamp(workout.EndTime)})
			}
			output.PrintTable(os.Stdout, []string{"ID", "Title", "Start", "End"}, rows)
			return nil
		}
	},
}

var workoutCmd = &cobra.Command{
	Use:   "workout",
	Short: "Get or mutate a workout",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		return runWorkoutGet(cmd, args[0])
	},
}

var workoutCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a workout from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if workoutFile == "" {
			return fmt.Errorf("workout file is required: use -f")
		}
		payload, err := readWorkoutRequestFile(workoutFile)
		if err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workout, err := client.CreateWorkout(contextForCommand(cmd), payload)
		if err != nil {
			return err
		}
		return renderWorkout(workout)
	},
}

var workoutUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a workout from a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if workoutFile == "" {
			return fmt.Errorf("workout file is required: use -f")
		}
		payload, err := readWorkoutRequestFile(workoutFile)
		if err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workout, err := client.UpdateWorkout(contextForCommand(cmd), args[0], payload)
		if err != nil {
			return err
		}
		return renderWorkout(workout)
	},
}

func init() {
	workoutCmd.AddCommand(workoutCreateCmd)
	workoutCmd.AddCommand(workoutUpdateCmd)
	workoutCreateCmd.Flags().StringVarP(&workoutFile, "file", "f", "", "Path to workout JSON file")
	workoutUpdateCmd.Flags().StringVarP(&workoutFile, "file", "f", "", "Path to workout JSON file")
}

func runWorkoutGet(cmd *cobra.Command, id string) error {
	client, err := clientFromConfig()
	if err != nil {
		return err
	}
	workout, err := client.GetWorkout(contextForCommand(cmd), id)
	if err != nil {
		return err
	}
	return renderWorkout(workout)
}

func renderWorkout(workout *api.Workout) error {
	if workout == nil {
		return fmt.Errorf("workout not found")
	}

	if app.outputMode == outputJSON {
		return output.PrintJSON(os.Stdout, workout)
	}

	compactLines := []string{
		fmt.Sprintf("%s | %s | %s | exercises=%d", workout.ID, workout.Title, formatTimestamp(workout.StartTime), len(workout.Exercises)),
	}
	if app.outputMode == outputCompact {
		return output.PrintCompact(os.Stdout, compactLines)
	}

	rows := [][2]string{{"id", workout.ID}, {"title", workout.Title}, {"description", output.ValueOrDash(workout.Description)}, {"start_time", formatTimestamp(workout.StartTime)}, {"end_time", formatTimestamp(workout.EndTime)}, {"created_at", formatTimestamp(workout.CreatedAt)}, {"updated_at", formatTimestamp(workout.UpdatedAt)}}
	output.PrintKeyValueTable(os.Stdout, rows)

	if len(workout.Exercises) > 0 {
		fmt.Fprintln(os.Stdout)
		exerciseRows := make([][]string, 0, len(workout.Exercises))
		for _, exercise := range workout.Exercises {
			exerciseRows = append(exerciseRows, []string{exercise.ExerciseTemplateID, output.ValueOrDash(exercise.Title), fmt.Sprintf("%d", len(exercise.Sets)), output.ValueOrDash(exercise.Notes)})
		}
		output.PrintTable(os.Stdout, []string{"Exercise ID", "Title", "Sets", "Notes"}, exerciseRows)
	}

	for _, exercise := range workout.Exercises {
		if len(exercise.Sets) == 0 {
			continue
		}
		fmt.Fprintf(os.Stdout, "\n%s\n", output.ValueOrDash(exercise.Title))
		setRows := make([][]string, 0, len(exercise.Sets))
		for _, set := range exercise.Sets {
			setRows = append(setRows, []string{fmt.Sprintf("%d", set.Index), set.Type, formatFloatPtr(set.WeightKG), formatIntPtr(set.Reps), formatIntPtr(set.DistanceMeters), formatIntPtr(set.DurationSeconds), formatFloatPtr(set.RPE)})
		}
		output.PrintTable(os.Stdout, []string{"#", "Type", "Weight KG", "Reps", "Distance M", "Duration S", "RPE"}, setRows)
	}

	return nil
}
