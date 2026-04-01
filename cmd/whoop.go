package cmd

import (
	"context"
	"fmt"

	"github.com/dhruvkelawala/hevy-cli/internal/whoop"
)

type whoopRecoverySnapshot struct {
	RecoveryScore float64 `json:"recovery_score"`
	HRVRMSSD      float64 `json:"hrv_rmssd_milli"`
	RestingHR     float64 `json:"resting_heart_rate"`
}

type whoopHistoryPoint struct {
	Day           string  `json:"day"`
	RecoveryScore float64 `json:"recovery_score"`
}

func fetchWhoopRecovery(ctx context.Context, days int) (*whoop.RecoveryCollection, error) {
	client, err := whoop.NewClient()
	if err != nil {
		return nil, err
	}
	return client.ListRecoveries(ctx, days)
}

func parseWhoopSnapshot(resp *whoop.RecoveryCollection) *whoopRecoverySnapshot {
	if resp == nil || len(resp.Records) == 0 {
		return nil
	}
	record := resp.Records[0]
	if record.ScoreState != "SCORED" && record.Score.RecoveryScore == 0 {
		return nil
	}
	return &whoopRecoverySnapshot{
		RecoveryScore: record.Score.RecoveryScore,
		HRVRMSSD:      record.Score.HRVRMSSD,
		RestingHR:     record.Score.RestingHeartRate,
	}
}

func parseWhoopHistory(resp *whoop.RecoveryCollection) []whoopHistoryPoint {
	if resp == nil {
		return nil
	}
	history := make([]whoopHistoryPoint, 0, len(resp.Records))
	for _, record := range resp.Records {
		label := "-"
		if !record.CreatedAt.IsZero() {
			label = record.CreatedAt.Local().Format("Mon")
		}
		history = append(history, whoopHistoryPoint{Day: label, RecoveryScore: record.Score.RecoveryScore})
	}
	return history
}

func whoopUnavailableMessage(err error) string {
	if err == nil {
		return "WHOOP unavailable. Falling back to training-only readiness."
	}
	return fmt.Sprintf("WHOOP unavailable (%s). Falling back to training-only readiness.", err)
}

func whoopStatus(score float64) (string, string) {
	switch {
	case score >= 67:
		return "GREEN", "Full send. Heavy compounds OK."
	case score >= 34:
		return "YELLOW", "Moderate session. Reduce volume 20%, skip heavy singles."
	default:
		return "RED", "Active recovery only. Light cardio or mobility."
	}
}
