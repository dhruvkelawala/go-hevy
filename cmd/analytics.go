package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
)

type personalRecord struct {
	Exercise string    `json:"exercise"`
	WeightKG float64   `json:"weight_kg"`
	Reps     int       `json:"reps"`
	Date     time.Time `json:"date"`
}

type volumePoint struct {
	WorkoutID string    `json:"workout_id"`
	Date      time.Time `json:"date"`
	Label     string    `json:"label"`
	VolumeKG  float64   `json:"volume_kg"`
}

type weekSummary struct {
	StartDate              time.Time           `json:"start_date"`
	EndDate                time.Time           `json:"end_date"`
	SessionCount           int                 `json:"session_count"`
	TotalVolumeKG          float64             `json:"total_volume_kg"`
	AverageDurationMinutes float64             `json:"average_duration_minutes"`
	MuscleGroups           map[string]int      `json:"muscle_groups"`
	MuscleDays             map[string][]string `json:"muscle_days"`
}

type workoutExerciseSummary struct {
	Title      string  `json:"title"`
	Sets       int     `json:"sets"`
	TotalReps  int     `json:"total_reps"`
	VolumeKG   float64 `json:"volume_kg"`
	BestWeight float64 `json:"best_weight_kg"`
}

type workoutDiff struct {
	LeftID    string                   `json:"left_id"`
	RightID   string                   `json:"right_id"`
	Overlap   []exerciseDiff           `json:"overlap"`
	OnlyLeft  []workoutExerciseSummary `json:"only_left"`
	OnlyRight []workoutExerciseSummary `json:"only_right"`
}

type exerciseDiff struct {
	Exercise        string  `json:"exercise"`
	LeftBestWeight  float64 `json:"left_best_weight_kg"`
	RightBestWeight float64 `json:"right_best_weight_kg"`
	WeightChange    float64 `json:"weight_change_kg"`
	LeftTotalReps   int     `json:"left_total_reps"`
	RightTotalReps  int     `json:"right_total_reps"`
	RepChange       int     `json:"rep_change"`
	LeftVolumeKG    float64 `json:"left_volume_kg"`
	RightVolumeKG   float64 `json:"right_volume_kg"`
	VolumeChangeKG  float64 `json:"volume_change_kg"`
}

type searchResult struct {
	WorkoutID    string   `json:"workout_id"`
	WorkoutTitle string   `json:"workout_title"`
	StartTime    string   `json:"start_time"`
	MatchedOn    []string `json:"matched_on"`
	Score        int      `json:"score"`
}

func parseAPITime(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return t.Local(), true
}

func weekStart(t time.Time) time.Time {
	local := t.Local()
	year, month, day := local.Date()
	midnight := time.Date(year, month, day, 0, 0, 0, 0, local.Location())
	offset := (int(midnight.Weekday()) + 6) % 7
	return midnight.AddDate(0, 0, -offset)
}

func calculateWorkoutStreak(workouts []api.Workout) (int, time.Time) {
	weeks := map[time.Time]bool{}
	latest := time.Time{}
	for _, workout := range workouts {
		start, ok := parseAPITime(workout.StartTime)
		if !ok {
			continue
		}
		ws := weekStart(start)
		weeks[ws] = true
		if latest.IsZero() || ws.After(latest) {
			latest = ws
		}
	}
	if latest.IsZero() {
		return 0, time.Time{}
	}
	count := 0
	since := latest
	for current := latest; weeks[current]; current = current.AddDate(0, 0, -7) {
		count++
		since = current
	}
	return count, since
}

func findPersonalRecord(exercise string, entries []api.ExerciseHistoryEntry) (personalRecord, bool) {
	best := personalRecord{Exercise: exercise}
	found := false
	for _, entry := range entries {
		if entry.WeightKG == nil {
			continue
		}
		reps := 0
		if entry.Reps != nil {
			reps = *entry.Reps
		}
		date, ok := parseAPITime(entry.WorkoutStartTime)
		if !ok {
			continue
		}
		candidate := personalRecord{Exercise: exercise, WeightKG: *entry.WeightKG, Reps: reps, Date: date}
		if !found || candidate.WeightKG > best.WeightKG || (candidate.WeightKG == best.WeightKG && candidate.Reps > best.Reps) || (candidate.WeightKG == best.WeightKG && candidate.Reps == best.Reps && candidate.Date.Before(best.Date)) {
			best = candidate
			found = true
		}
	}
	return best, found
}

func buildVolumePoints(entries []api.ExerciseHistoryEntry) []volumePoint {
	byWorkout := map[string]volumePoint{}
	for _, entry := range entries {
		if entry.WeightKG == nil || entry.Reps == nil {
			continue
		}
		date, ok := parseAPITime(entry.WorkoutStartTime)
		if !ok {
			continue
		}
		point := byWorkout[entry.WorkoutID]
		point.WorkoutID = entry.WorkoutID
		point.Date = date
		point.Label = date.Format("Jan 02")
		point.VolumeKG += *entry.WeightKG * float64(*entry.Reps)
		byWorkout[entry.WorkoutID] = point
	}
	points := make([]volumePoint, 0, len(byWorkout))
	for _, point := range byWorkout {
		points = append(points, point)
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points
}

func renderVolumeChart(title string, points []volumePoint, maxPoints int) []string {
	if maxPoints > 0 && len(points) > maxPoints {
		points = points[len(points)-maxPoints:]
	}
	lines := []string{fmt.Sprintf("%s volume — last %d sessions", title, len(points))}
	if len(points) == 0 {
		return append(lines, "No volume history found.")
	}
	maxVolume := 0.0
	for _, point := range points {
		if point.VolumeKG > maxVolume {
			maxVolume = point.VolumeKG
		}
	}
	for _, point := range points {
		width := 1
		if maxVolume > 0 {
			width = int((point.VolumeKG / maxVolume) * 14)
			if width < 1 {
				width = 1
			}
		}
		lines = append(lines, fmt.Sprintf("%s  %s%s  %s", point.Label, formatWeight(point.VolumeKG), app.weightUnit, strings.Repeat("█", width)))
	}
	return lines
}

func workoutDurationMinutes(workout api.Workout) float64 {
	start, okStart := parseAPITime(workout.StartTime)
	end, okEnd := parseAPITime(workout.EndTime)
	if !okStart || !okEnd || !end.After(start) {
		return 0
	}
	return end.Sub(start).Minutes()
}

func workoutVolume(workout api.Workout) float64 {
	total := 0.0
	for _, exercise := range workout.Exercises {
		for _, set := range exercise.Sets {
			if set.WeightKG != nil && set.Reps != nil {
				total += *set.WeightKG * float64(*set.Reps)
			}
		}
	}
	return total
}

func summarizeWeek(workouts []api.Workout, catalog map[string]api.ExerciseTemplate, now time.Time, previous bool) weekSummary {
	start := weekStart(now)
	if previous {
		start = start.AddDate(0, 0, -7)
	}
	end := start.AddDate(0, 0, 7)
	summary := weekSummary{StartDate: start, EndDate: end.Add(-time.Nanosecond), MuscleGroups: map[string]int{}, MuscleDays: map[string][]string{}}
	var durationTotal float64
	for _, workout := range workouts {
		startedAt, ok := parseAPITime(workout.StartTime)
		if !ok || startedAt.Before(start) || !startedAt.Before(end) {
			continue
		}
		summary.SessionCount++
		summary.TotalVolumeKG += workoutVolume(workout)
		durationTotal += workoutDurationMinutes(workout)
		weekday := startedAt.Format("Mon")
		seenMuscles := map[string]bool{}
		for _, exercise := range workout.Exercises {
			muscle := primaryMuscleForExercise(exercise, catalog)
			if muscle == "" || seenMuscles[muscle] {
				continue
			}
			seenMuscles[muscle] = true
			summary.MuscleGroups[muscle]++
			if !containsString(summary.MuscleDays[muscle], weekday) {
				summary.MuscleDays[muscle] = append(summary.MuscleDays[muscle], weekday)
			}
		}
	}
	if summary.SessionCount > 0 {
		summary.AverageDurationMinutes = durationTotal / float64(summary.SessionCount)
	}
	for muscle := range summary.MuscleDays {
		sort.Strings(summary.MuscleDays[muscle])
	}
	return summary
}

func primaryMuscleForExercise(exercise api.Exercise, catalog map[string]api.ExerciseTemplate) string {
	if template, ok := catalog[exercise.ExerciseTemplateID]; ok && strings.TrimSpace(template.PrimaryMuscleGroup) != "" {
		return strings.ToLower(template.PrimaryMuscleGroup)
	}
	title := strings.ToLower(exercise.Title)
	switch {
	case strings.Contains(title, "bench"), strings.Contains(title, "press"), strings.Contains(title, "fly"):
		return "chest"
	case strings.Contains(title, "row"), strings.Contains(title, "pull"), strings.Contains(title, "lat"):
		return "back"
	case strings.Contains(title, "squat"), strings.Contains(title, "deadlift"), strings.Contains(title, "leg"):
		return "legs"
	case strings.Contains(title, "shoulder"), strings.Contains(title, "lateral raise"), strings.Contains(title, "overhead"):
		return "shoulders"
	case strings.Contains(title, "curl"), strings.Contains(title, "tricep"), strings.Contains(title, "bicep"):
		return "arms"
	default:
		return ""
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func summarizeWorkoutExercises(workout api.Workout) map[string]workoutExerciseSummary {
	summary := map[string]workoutExerciseSummary{}
	for _, exercise := range workout.Exercises {
		entry := summary[exercise.Title]
		entry.Title = exercise.Title
		for _, set := range exercise.Sets {
			entry.Sets++
			if set.Reps != nil {
				entry.TotalReps += *set.Reps
			}
			if set.WeightKG != nil {
				if *set.WeightKG > entry.BestWeight {
					entry.BestWeight = *set.WeightKG
				}
				if set.Reps != nil {
					entry.VolumeKG += *set.WeightKG * float64(*set.Reps)
				}
			}
		}
		summary[exercise.Title] = entry
	}
	return summary
}

func compareWorkouts(left, right api.Workout) workoutDiff {
	leftSummary := summarizeWorkoutExercises(left)
	rightSummary := summarizeWorkoutExercises(right)
	diff := workoutDiff{LeftID: left.ID, RightID: right.ID}
	for title, leftEntry := range leftSummary {
		if rightEntry, ok := rightSummary[title]; ok {
			diff.Overlap = append(diff.Overlap, exerciseDiff{
				Exercise:        title,
				LeftBestWeight:  leftEntry.BestWeight,
				RightBestWeight: rightEntry.BestWeight,
				WeightChange:    rightEntry.BestWeight - leftEntry.BestWeight,
				LeftTotalReps:   leftEntry.TotalReps,
				RightTotalReps:  rightEntry.TotalReps,
				RepChange:       rightEntry.TotalReps - leftEntry.TotalReps,
				LeftVolumeKG:    leftEntry.VolumeKG,
				RightVolumeKG:   rightEntry.VolumeKG,
				VolumeChangeKG:  rightEntry.VolumeKG - leftEntry.VolumeKG,
			})
			continue
		}
		diff.OnlyLeft = append(diff.OnlyLeft, leftEntry)
	}
	for title, rightEntry := range rightSummary {
		if _, ok := leftSummary[title]; !ok {
			diff.OnlyRight = append(diff.OnlyRight, rightEntry)
		}
	}
	sort.Slice(diff.Overlap, func(i, j int) bool { return diff.Overlap[i].Exercise < diff.Overlap[j].Exercise })
	sort.Slice(diff.OnlyLeft, func(i, j int) bool { return diff.OnlyLeft[i].Title < diff.OnlyLeft[j].Title })
	sort.Slice(diff.OnlyRight, func(i, j int) bool { return diff.OnlyRight[i].Title < diff.OnlyRight[j].Title })
	return diff
}

func searchWorkouts(workouts []api.Workout, query string) []searchResult {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return nil
	}
	results := []searchResult{}
	for _, workout := range workouts {
		result := searchResult{WorkoutID: workout.ID, WorkoutTitle: workout.Title, StartTime: workout.StartTime}
		if score := fuzzyScore(workout.Title, needle); score > 0 {
			result.Score += score + 10
			result.MatchedOn = append(result.MatchedOn, "title")
		}
		for _, exercise := range workout.Exercises {
			if score := fuzzyScore(exercise.Title, needle); score > 0 {
				result.Score += score
				result.MatchedOn = append(result.MatchedOn, exercise.Title)
			}
		}
		if result.Score > 0 {
			result.MatchedOn = dedupeStrings(result.MatchedOn)
			results = append(results, result)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].StartTime > results[j].StartTime
		}
		return results[i].Score > results[j].Score
	})
	return results
}

func fuzzyScore(value, needle string) int {
	title := strings.ToLower(strings.TrimSpace(value))
	switch {
	case title == needle:
		return 100
	case strings.HasPrefix(title, needle):
		return 75
	case strings.Contains(title, needle):
		return 50
	default:
		compact := strings.ReplaceAll(title, " ", "")
		compactNeedle := strings.ReplaceAll(needle, " ", "")
		if strings.Contains(compact, compactNeedle) {
			return 25
		}
		return 0
	}
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func renderCalendar(year int, month time.Month, workoutDays map[int]bool) []string {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	lastDay := first.AddDate(0, 1, -1).Day()
	lines := []string{first.Format("January 2006"), "Mo Tu We Th Fr Sa Su"}
	week := make([]string, 0, 7)
	padding := (int(first.Weekday()) + 6) % 7
	for i := 0; i < padding; i++ {
		week = append(week, "  ")
	}
	for day := 1; day <= lastDay; day++ {
		cell := fmt.Sprintf("%2d", day)
		if workoutDays[day] {
			cell = fmt.Sprintf("[%d]", day)
			if day < 10 {
				cell = fmt.Sprintf("[%d]", day)
			}
		}
		week = append(week, cell)
		if len(week) == 7 {
			lines = append(lines, strings.Join(week, " "))
			week = make([]string, 0, 7)
		}
	}
	if len(week) > 0 {
		for len(week) < 7 {
			week = append(week, "  ")
		}
		lines = append(lines, strings.Join(week, " "))
	}
	return lines
}
