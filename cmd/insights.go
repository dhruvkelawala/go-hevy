package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var prAll bool
var weekPrev bool

var streakCmd = &cobra.Command{
	Use:   "streak",
	Short: "Show your current weekly workout streak",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		streak, since := calculateWorkoutStreak(workouts)
		response := map[string]any{"weeks": streak}
		if !since.IsZero() {
			response["since"] = since.Format(time.RFC3339)
		}
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, response)
		}
		if streak == 0 {
			printLine("Current streak: 0 weeks")
			return nil
		}
		printLine("Current streak: %d weeks (since %s)", streak, since.Format("Jan 2"))
		return nil
	},
}

var prCmd = &cobra.Command{
	Use:   "pr <exercise-name>",
	Short: "Show personal records by exercise",
	Args: func(cmd *cobra.Command, args []string) error {
		if prAll {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		if prAll {
			workouts, err := fetchWorkouts(client, 1, 0, true)
			if err != nil {
				return err
			}
			detailed, err := fetchWorkoutDetails(client, workouts)
			if err != nil {
				return err
			}
			records := aggregatePRsFromWorkouts(detailed)
			if app.outputMode == outputJSON {
				return output.PrintJSON(os.Stdout, records)
			}
			rows := make([][]string, 0, len(records))
			for _, record := range records {
				rows = append(rows, []string{record.Exercise, formatWeight(record.WeightKG), fmt.Sprintf("%d", record.Reps), record.Date.Format("2006-01-02")})
			}
			output.PrintTable(os.Stdout, []string{"Exercise", weightHeader(), "Reps", "Date"}, rows)
			return nil
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		exercise := pickExerciseByName(exercises, args[0])
		if exercise == nil {
			return fmt.Errorf("no exercise found matching %q", args[0])
		}
		history, err := client.GetExerciseHistory(contextForCommand(cmd), exercise.ID, "", "")
		if err != nil {
			return err
		}
		record, ok := findPersonalRecord(exercise.Title, history.ExerciseHistory)
		if !ok {
			return fmt.Errorf("no weighted history found for %q", exercise.Title)
		}
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, record)
		}
		output.PrintKeyValueTable(os.Stdout, [][2]string{{"exercise", record.Exercise}, {"weight", formatWeight(record.WeightKG)}, {"reps", fmt.Sprintf("%d", record.Reps)}, {"date", record.Date.Format("2006-01-02")}})
		return nil
	},
}

var weekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show a weekly training summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		detailed, err := fetchWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		summary := summarizeWeek(detailed, exerciseCatalogMap(exercises), time.Now(), weekPrev)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, summary)
		}
		muscles := sortedMuscleSummaries(summary.MuscleGroups)
		output.PrintKeyValueTable(os.Stdout, [][2]string{{"week", summary.StartDate.Format("2006-01-02") + " to " + summary.EndDate.Format("2006-01-02")}, {"sessions", fmt.Sprintf("%d", summary.SessionCount)}, {"total_volume", formatWeight(summary.TotalVolumeKG)}, {"avg_duration_min", fmt.Sprintf("%.1f", summary.AverageDurationMinutes)}, {"muscle_groups", strings.Join(muscles, ", ")}})
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [workout-id] [workout-id]",
	Short: "Compare two workouts side by side (defaults to last two)",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		// If no IDs provided, use the two most recent workouts
		if len(args) < 2 {
			workouts, err := fetchWorkouts(client, 1, 0, false)
			if err != nil {
				return err
			}
			if len(workouts) < 2 {
				return fmt.Errorf("need at least 2 workouts to compare; found %d", len(workouts))
			}
			if len(args) == 0 {
				args = []string{workouts[0].ID, workouts[1].ID}
			} else {
				args = append(args, workouts[1].ID)
			}
		}
		left, err := client.GetWorkout(contextForCommand(cmd), args[0])
		if err != nil {
			return err
		}
		right, err := client.GetWorkout(contextForCommand(cmd), args[1])
		if err != nil {
			return err
		}
		diff := compareWorkouts(*left, *right)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, diff)
		}
		printLine("%s vs %s", left.Title, right.Title)
		if len(diff.Overlap) > 0 {
			rows := make([][]string, 0, len(diff.Overlap))
			for _, item := range diff.Overlap {
				rows = append(rows, []string{item.Exercise, formatWeight(item.LeftBestWeight), formatWeight(item.RightBestWeight), signedFloat(item.WeightChange), fmt.Sprintf("%d→%d", item.LeftTotalReps, item.RightTotalReps), signedFloat(item.VolumeChangeKG)})
			}
			output.PrintTable(os.Stdout, []string{"Exercise", "Left Best", "Right Best", "Δ Weight", "Reps", "Δ Volume"}, rows)
		}
		if len(diff.OnlyLeft) > 0 {
			printLine("Only in %s: %s", left.Title, joinExerciseTitles(diff.OnlyLeft))
		}
		if len(diff.OnlyRight) > 0 {
			printLine("Only in %s: %s", right.Title, joinExerciseTitles(diff.OnlyRight))
		}
		return nil
	},
}

var volumeCmd = &cobra.Command{
	Use:   "volume <exercise-name>",
	Short: "Show volume over time for an exercise",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		exercise := pickExerciseByName(exercises, args[0])
		if exercise == nil {
			return fmt.Errorf("no exercise found matching %q", args[0])
		}
		history, err := client.GetExerciseHistory(contextForCommand(cmd), exercise.ID, "", "")
		if err != nil {
			return err
		}
		points := buildVolumePoints(history.ExerciseHistory)
		lines := renderVolumeChart(exercise.Title, points, 8)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, map[string]any{"exercise": exercise, "points": points, "chart": lines})
		}
		return output.PrintCompact(os.Stdout, lines)
	},
}

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's workout if one exists",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 10, false)
		if err != nil {
			return err
		}
		today := time.Now().Format("2006-01-02")
		for _, workout := range workouts {
			startedAt, ok := parseAPITime(workout.StartTime)
			if ok && startedAt.Format("2006-01-02") == today {
				return runWorkoutGet(cmd, workout.ID)
			}
		}
		printLine("No workout today yet.")
		return nil
	},
}

var musclesCmd = &cobra.Command{
	Use:   "muscles",
	Short: "Show muscle groups hit this week",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		detailed, err := fetchWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		summary := summarizeWeek(detailed, exerciseCatalogMap(exercises), time.Now(), false)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, summary.MuscleDays)
		}
		// Collect all muscle groups from data, sorted alphabetically
		allMuscles := make([]string, 0, len(summary.MuscleDays))
		for muscle := range summary.MuscleDays {
			allMuscles = append(allMuscles, muscle)
		}
		sort.Strings(allMuscles)
		if len(allMuscles) == 0 {
			printLine("No muscles hit this week")
		}
		for _, muscle := range allMuscles {
			days := summary.MuscleDays[muscle]
			label := strings.ReplaceAll(muscle, "_", " ")
			printLine("%-14s %s (%s)", label, strings.Repeat("█", summary.MuscleGroups[muscle]*2), strings.Join(days, ", "))
		}
		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search workouts by title or exercise name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		detailed, err := fetchWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		results := searchWorkouts(detailed, args[0])
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, results)
		}
		rows := make([][]string, 0, len(results))
		for _, result := range results {
			rows = append(rows, []string{result.WorkoutID, result.WorkoutTitle, formatTimestamp(result.StartTime), strings.Join(result.MatchedOn, ", ")})
		}
		output.PrintTable(os.Stdout, []string{"Workout ID", "Title", "Date", "Matched On"}, rows)
		return nil
	},
}

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Show an ASCII workout calendar for the current month",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		now := time.Now()
		workoutDays := map[int]bool{}
		for _, workout := range workouts {
			startedAt, ok := parseAPITime(workout.StartTime)
			if !ok || startedAt.Month() != now.Month() || startedAt.Year() != now.Year() {
				continue
			}
			workoutDays[startedAt.Day()] = true
		}
		lines := renderCalendar(now.Year(), now.Month(), workoutDays)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, map[string]any{"month": now.Month().String(), "year": now.Year(), "workout_days": workoutDays, "calendar": lines})
		}
		return output.PrintCompact(os.Stdout, lines)
	},
}

func init() {
	prCmd.Flags().BoolVar(&prAll, "all", false, "Show PRs for all exercises")
	weekCmd.Flags().BoolVar(&weekPrev, "prev", false, "Show the previous week")
	rootCmd.AddCommand(streakCmd)
	rootCmd.AddCommand(prCmd)
	rootCmd.AddCommand(weekCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(volumeCmd)
	rootCmd.AddCommand(todayCmd)
	rootCmd.AddCommand(musclesCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(calendarCmd)
}

func aggregatePRsFromWorkouts(workouts []api.Workout) []personalRecord {
	bestByExercise := map[string]personalRecord{}
	for _, workout := range workouts {
		date, ok := parseAPITime(workout.StartTime)
		if !ok {
			continue
		}
		for _, exercise := range workout.Exercises {
			for _, set := range exercise.Sets {
				if set.WeightKG == nil {
					continue
				}
				reps := 0
				if set.Reps != nil {
					reps = *set.Reps
				}
				candidate := personalRecord{Exercise: exercise.Title, WeightKG: *set.WeightKG, Reps: reps, Date: date}
				current, ok := bestByExercise[exercise.Title]
				if !ok || candidate.WeightKG > current.WeightKG || (candidate.WeightKG == current.WeightKG && candidate.Reps > current.Reps) {
					bestByExercise[exercise.Title] = candidate
				}
			}
		}
	}
	records := make([]personalRecord, 0, len(bestByExercise))
	for _, record := range bestByExercise {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Exercise < records[j].Exercise })
	return records
}

func sortedMuscleSummaries(groups map[string]int) []string {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, fmt.Sprintf("%s(%d)", key, groups[key]))
	}
	return result
}

func signedFloat(value float64) string {
	if value > 0 {
		return "+" + formatWeight(value)
	}
	return formatWeight(value)
}

func joinExerciseTitles(values []workoutExerciseSummary) string {
	titles := make([]string, 0, len(values))
	for _, value := range values {
		titles = append(titles, value.Title)
	}
	return strings.Join(titles, ", ")
}
