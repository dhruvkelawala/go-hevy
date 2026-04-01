package whoop

import "time"

type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UpdatedAt    time.Time `json:"updated_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type RecoveryCollection struct {
	Records   []RecoveryRecord `json:"records"`
	NextToken string           `json:"next_token"`
}

type RecoveryRecord struct {
	CycleID    int64         `json:"cycle_id"`
	SleepID    string        `json:"sleep_id"`
	UserID     int64         `json:"user_id"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	ScoreState string        `json:"score_state"`
	Score      RecoveryScore `json:"score"`
}

type RecoveryScore struct {
	UserCalibrating  bool    `json:"user_calibrating"`
	RecoveryScore    float64 `json:"recovery_score"`
	RestingHeartRate float64 `json:"resting_heart_rate"`
	HRVRMSSD         float64 `json:"hrv_rmssd_milli"`
	SPO2Percentage   float64 `json:"spo2_percentage,omitempty"`
	SkinTempCelsius  float64 `json:"skin_temp_celsius,omitempty"`
}
