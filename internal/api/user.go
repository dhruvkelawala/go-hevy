package api

import (
	"context"
	"encoding/json"
)

func (c *Client) GetMe(ctx context.Context) (map[string]any, error) {
	resp := map[string]any{}
	err := c.request(ctx, "GET", "/v1/user/info", nil, nil, &resp)
	if err != nil {
		return nil, err
	}
	if data, ok := resp["data"].(map[string]any); ok {
		return data, nil
	}
	return resp, nil
}

func DecodeUserInfo(data map[string]any) (*UserInfo, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	user := &UserInfo{}
	if err := json.Unmarshal(raw, user); err != nil {
		return nil, err
	}
	user.Raw = data
	return user, nil
}
