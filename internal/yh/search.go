package yh

import (
	"encoding/json"
	"strings"
	"time"
)

// SearchHit is one message the search of a chat found: what was said, by
// whom, and when. The route answers a page of them, newest first.
//
// The content of a message is the message's own body, which the route
// hands back as it is kept: a text says its words, a picture says where it
// is, and both are said in a little JSON of their own.
type SearchHit struct {
	ID        string
	Sequence  int64
	Name      string
	AvatarURL string
	Content   string
	// URL is where a picture of the message is, when it is one.
	URL  string
	Type int
	Time time.Time
}

// SearchChat asks for the messages of a chat that match word, of the kind
// the tabs of the search window name: "" or "all" for any, "2" for
// pictures, "4" for files, "7" for stickers. before is the newest message
// it may answer with, so that a search can be asked for again from where
// it stopped.
func (c *Client) SearchChat(chatID string, chatType int, word, kind string, before time.Time, size int) ([]SearchHit, error) {
	if kind == "" {
		kind = "all"
	}
	if size <= 0 {
		size = 40
	}
	at := int64(9999999999999)
	if !before.IsZero() {
		at = before.UnixMilli()
	}
	body, err := json.Marshal(map[string]any{
		"word":      word,
		"chatId":    chatID,
		"chatType":  chatType,
		"type":      kind,
		"size":      size,
		"time":      at,
		"direction": 1,
	})
	if err != nil {
		return nil, err
	}
	data, err := c.do("POST", "/v1/search/chat-search", body, false)
	if err != nil {
		return nil, err
	}
	var out struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			List []struct {
				ID        string `json:"id"`
				Sequence  int64  `json:"sequence"`
				Name      string `json:"name"`
				AvatarURL string `json:"avatarUrl"`
				Content   string `json:"content"`
				Type      string `json:"type"`
				Time      int64  `json:"time"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out.Code != 1 {
		return nil, &APIError{Code: out.Code, Msg: out.Msg}
	}
	hits := make([]SearchHit, 0, len(out.Data.List))
	for _, it := range out.Data.List {
		text, url := splitContent(it.Content)
		hit := SearchHit{
			ID:        it.ID,
			Sequence:  it.Sequence,
			Name:      it.Name,
			AvatarURL: it.AvatarURL,
			Content:   text,
			URL:       url,
			Time:      time.UnixMilli(it.Time),
		}
		hit.Type = atoi(it.Type)
		hits = append(hits, hit)
	}
	return hits, nil
}

// splitContent reads what a message says. The route hands the body back as
// it is kept, which for most messages is a little JSON: the words of a
// text, or where a picture is. A body that is not that is the words.
func splitContent(raw string) (text, url string) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") {
		return raw, ""
	}
	var c struct {
		Text     string `json:"text"`
		Image    string `json:"image"`
		ImageURL string `json:"imageUrl"`
	}
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return raw, ""
	}
	switch {
	case c.Image != "":
		return c.Text, c.Image
	case c.ImageURL != "":
		return c.Text, c.ImageURL
	}
	return c.Text, ""
}

// atoi reads the number a field holds, which the route writes as a string
// in one place and as a number in another.
func atoi(s string) int {
	s = strings.TrimSpace(s)
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
