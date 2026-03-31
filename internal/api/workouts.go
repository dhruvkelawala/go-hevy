package api

import (
	"context"
	"fmt"
	"strconv"
)

func (c *Client) ListWorkouts(ctx context.Context, page, pageSize int) (*PaginatedWorkouts, error) {
	resp := &PaginatedWorkouts{}
	err := c.request(ctx, "GET", "/v1/workouts", map[string]string{
		"page":     strconv.Itoa(page),
		"pageSize": strconv.Itoa(pageSize),
	}, nil, resp)
	return resp, err
}

func (c *Client) GetWorkout(ctx context.Context, id string) (*Workout, error) {
	resp := &Workout{}
	err := c.request(ctx, "GET", fmt.Sprintf("/v1/workouts/%s", id), nil, nil, resp)
	return resp, err
}

func (c *Client) CreateWorkout(ctx context.Context, payload CreateWorkoutRequest) (*Workout, error) {
	resp := &Workout{}
	err := c.request(ctx, "POST", "/v1/workouts", nil, payload, resp)
	return resp, err
}

func (c *Client) UpdateWorkout(ctx context.Context, id string, payload CreateWorkoutRequest) (*Workout, error) {
	resp := &Workout{}
	err := c.request(ctx, "PUT", fmt.Sprintf("/v1/workouts/%s", id), nil, payload, resp)
	return resp, err
}

func (c *Client) GetWorkoutCount(ctx context.Context) (*WorkoutCount, error) {
	resp := &WorkoutCount{}
	err := c.request(ctx, "GET", "/v1/workouts/count", nil, nil, resp)
	return resp, err
}
