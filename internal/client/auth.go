package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Login аутентифицирует пользователя и сохраняет полученный токен.
func (c *Client) Login(ctx context.Context, email, password string) error {
	respBody, err := c.do(ctx, "POST", "/users/login", map[string]string{
		"email":    email,
		"password": password,
	}, false)
	if err != nil {
		return err
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return fmt.Errorf("client: decode login: %w", err)
	}
	c.token = out.Token
	return nil
}

// Register создаёт нового пользователя.
func (c *Client) Register(ctx context.Context, email, password string) error {
	_, err := c.do(ctx, "POST", "/users", map[string]string{
		"email":    email,
		"password": password,
	}, false)
	return err
}
