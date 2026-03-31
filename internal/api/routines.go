package api

import (
	"context"
	"fmt"
	"strconv"
)

func (c *Client) ListRoutines(ctx context.Context, page, pageSize int) (*PaginatedRoutines, error) {
	resp := &PaginatedRoutines{}
	err := c.request(ctx, "GET", "/v1/routines", map[string]string{
		"page":     strconv.Itoa(page),
		"pageSize": strconv.Itoa(pageSize),
	}, nil, resp)
	return resp, err
}

func (c *Client) GetRoutine(ctx context.Context, id string) (*Routine, error) {
	resp := &RoutineResponse{}
	err := c.request(ctx, "GET", fmt.Sprintf("/v1/routines/%s", id), nil, nil, resp)
	if err != nil {
		return nil, err
	}
	return &resp.Routine, nil
}
