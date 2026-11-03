package yh

import (
	"encoding/json"
	"fmt"
)

// CreateGroup makes a chat of the account's own and answers with its id.
func (c *Client) CreateGroup(name, intro string) (string, error) {
	var body []byte
	body = pbString(body, 2, name)
	body = pbString(body, 3, intro)

	data, err := c.do("POST", "/v1/group/create-group", body, true)
	if err != nil {
		return "", err
	}
	var status Status
	var id string
	fields(data, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			if v, ok := r.value(wire); ok {
				status = decodeStatus(v)
			}
			return true
		case 2:
			id = str(wire, r)
			return true
		}
		return false
	})
	if err := status.Err(); err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("yh: the group was made without an id")
	}
	return id, nil
}

// CreateBot makes a bot of the account's own. The route is the one the
// site edits a bot with: a bot with no id is a bot that does not exist yet.
func (c *Client) CreateBot(name, intro string) error {
	body, err := json.Marshal(map[string]any{
		"nickname":     name,
		"introduction": intro,
		"avatarUrl":    "",
		"private":      0,
	})
	if err != nil {
		return err
	}
	data, err := c.do("POST", "/v1/bot/web-edit-bot", body, false)
	if err != nil {
		return err
	}
	var out struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	if out.Code != 1 {
		return &APIError{Code: out.Code, Msg: out.Msg}
	}
	return nil
}
