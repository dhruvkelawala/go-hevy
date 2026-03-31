package api

type PaginatedWorkouts struct {
	Page      int       `json:"page"`
	PageCount int       `json:"page_count"`
	Workouts  []Workout `json:"workouts"`
}

type PaginatedRoutines struct {
	Page      int       `json:"page"`
	PageCount int       `json:"page_count"`
	Routines  []Routine `json:"routines"`
}

type PaginatedExerciseTemplates struct {
	Page              int                `json:"page"`
	PageCount         int                `json:"page_count"`
	ExerciseTemplates []ExerciseTemplate `json:"exercise_templates"`
}

type WorkoutCount struct {
	WorkoutCount int `json:"workout_count"`
}

type Workout struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartTime   string     `json:"start_time"`
	EndTime     string     `json:"end_time"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
	IsPrivate   bool       `json:"is_private,omitempty"`
	Exercises   []Exercise `json:"exercises,omitempty"`
}

type Routine struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Notes     string            `json:"notes,omitempty"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Exercises []RoutineExercise `json:"exercises,omitempty"`
}

type RoutineResponse struct {
	Routine Routine `json:"routine"`
}

type Exercise struct {
	Index              int    `json:"index,omitempty"`
	Title              string `json:"title"`
	Notes              string `json:"notes,omitempty"`
	ExerciseTemplateID string `json:"exercise_template_id"`
	SupersetID         *int   `json:"superset_id,omitempty"`
	SupersetsID        *int   `json:"supersets_id,omitempty"`
	Sets               []Set  `json:"sets,omitempty"`
}

type RoutineExercise struct {
	Exercise
	RestSeconds *int `json:"rest_seconds,omitempty"`
}

type Set struct {
	Index           int       `json:"index,omitempty"`
	Type            string    `json:"type"`
	WeightKG        *float64  `json:"weight_kg,omitempty"`
	Reps            *int      `json:"reps,omitempty"`
	DistanceMeters  *int      `json:"distance_meters,omitempty"`
	DurationSeconds *int      `json:"duration_seconds,omitempty"`
	RPE             *float64  `json:"rpe,omitempty"`
	CustomMetric    *float64  `json:"custom_metric,omitempty"`
	RepRange        *RepRange `json:"rep_range,omitempty"`
}

type RepRange struct {
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
}

type ExerciseTemplate struct {
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Type                  string   `json:"type"`
	PrimaryMuscleGroup    string   `json:"primary_muscle_group"`
	SecondaryMuscleGroups []string `json:"secondary_muscle_groups"`
	IsCustom              bool     `json:"is_custom"`
}

type ExerciseHistoryResponse struct {
	ExerciseHistory []ExerciseHistoryEntry `json:"exercise_history"`
}

type ExerciseHistoryEntry struct {
	WorkoutID          string   `json:"workout_id"`
	WorkoutTitle       string   `json:"workout_title"`
	WorkoutStartTime   string   `json:"workout_start_time"`
	WorkoutEndTime     string   `json:"workout_end_time"`
	ExerciseTemplateID string   `json:"exercise_template_id"`
	WeightKG           *float64 `json:"weight_kg,omitempty"`
	Reps               *int     `json:"reps,omitempty"`
	DistanceMeters     *int     `json:"distance_meters,omitempty"`
	DurationSeconds    *int     `json:"duration_seconds,omitempty"`
	RPE                *float64 `json:"rpe,omitempty"`
	CustomMetric       *float64 `json:"custom_metric,omitempty"`
	SetType            string   `json:"set_type"`
}

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Timezone    string `json:"timezone"`

	Raw map[string]any `json:"-"`
}

type APIError struct {
	Error string `json:"error"`
}

type CreateWorkoutRequest struct {
	Workout WorkoutInput `json:"workout"`
}

type WorkoutInput struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	StartTime   string                 `json:"start_time"`
	EndTime     string                 `json:"end_time"`
	IsPrivate   bool                   `json:"is_private"`
	Exercises   []WorkoutExerciseInput `json:"exercises"`
}

type WorkoutExerciseInput struct {
	ExerciseTemplateID string            `json:"exercise_template_id"`
	SupersetID         *int              `json:"superset_id,omitempty"`
	Notes              string            `json:"notes,omitempty"`
	Sets               []WorkoutSetInput `json:"sets"`
}

type WorkoutSetInput struct {
	Type            string   `json:"type"`
	WeightKG        *float64 `json:"weight_kg,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	DistanceMeters  *int     `json:"distance_meters,omitempty"`
	DurationSeconds *int     `json:"duration_seconds,omitempty"`
	CustomMetric    *float64 `json:"custom_metric,omitempty"`
	RPE             *float64 `json:"rpe,omitempty"`
}
