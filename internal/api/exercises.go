package api

import (
	"context"
	"fmt"
	"strconv"
)

func (c *Client) ListExercises(ctx context.Context, page, pageSize int) (*PaginatedExerciseTemplates, error) {
	resp := &PaginatedExerciseTemplates{}
	err := c.request(ctx, "GET", "/v1/exercise_templates", map[string]string{
		"page":     strconv.Itoa(page),
		"pageSize": strconv.Itoa(pageSize),
	}, nil, resp)
	return resp, err
}

func (c *Client) GetExercise(ctx context.Context, id string) (*ExerciseTemplate, error) {
	resp := &ExerciseTemplate{}
	err := c.request(ctx, "GET", fmt.Sprintf("/v1/exercise_templates/%s", id), nil, nil, resp)
	return resp, err
}

func (c *Client) GetExerciseHistory(ctx context.Context, id, startDate, endDate string) (*ExerciseHistoryResponse, error) {
	query := map[string]string{}
	if startDate != "" {
		query["start_date"] = startDate
	}
	if endDate != "" {
		query["end_date"] = endDate
	}

	resp := &ExerciseHistoryResponse{}
	err := c.request(ctx, "GET", fmt.Sprintf("/v1/exercise_history/%s", id), query, nil, resp)
	return resp, err
}
