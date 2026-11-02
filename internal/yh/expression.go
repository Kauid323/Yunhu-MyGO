package yh

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Expression is a sticker the account has saved: the picture it is, which the
// sticker host serves, and the id a message about it carries, which the
// server keeps so that the message can be added back to the collection.
type Expression struct {
	// ID is the sticker's own id. The message that carries it takes it as
	// a string, which is what the proto says, so it is kept as one.
	ID string
	// URL is the picture as the host names it, usually a relative path.
	URL string
}

// Expressions lists the stickers the account has saved: what the panel over
// the composer shows.
func (c *Client) Expressions() ([]Expression, error) {
	data, err := c.do("POST", "/v1/expression/list", []byte("{}"), false)
	if err != nil {
		return nil, err
	}
	var out struct {
		jsonStatus
		Data struct {
			Expression []struct {
				ID  int64  `json:"id"`
				URL string `json:"url"`
			} `json:"expression"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrResponse, err)
	}
	if err := out.err(); err != nil {
		return nil, err
	}
	list := make([]Expression, 0, len(out.Data.Expression))
	for _, e := range out.Data.Expression {
		if e.URL == "" {
			continue
		}
		list = append(list, Expression{ID: strconv.FormatInt(e.ID, 10), URL: e.URL})
	}
	return list, nil
}
