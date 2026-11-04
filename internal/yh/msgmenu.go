package yh

import (
	"encoding/json"
	"fmt"
	"time"
)

// DeleteMessage takes a message off the account's own history. It is the
// delete of the client, not a recall: the message stays for everyone else.
func (c *Client) DeleteMessage(chatID string, chatType int, msgID string) error {
	body, err := json.Marshal(map[string]any{
		"chatId":   chatID,
		"chatType": chatType,
		"msgId":    msgID,
	})
	if err != nil {
		return err
	}
	data, err := c.do("POST", "/v1/msg/delete", body, false)
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

// RecallMessages takes messages back for everyone in the chat, as the
// client's recall does: they stay in the history as a line that says they
// were recalled, and the words are gone. The route takes a list, so picking
// several messages and taking them back is one request.
func (c *Client) RecallMessages(chatID string, chatType int, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	var body []byte
	for _, id := range ids {
		body = pbString(body, 2, id)
	}
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	data, err := c.do("POST", "/v1/msg/recall-msg", body, true)
	if err != nil {
		return err
	}
	return decodeStatusResponse(data).Err()
}

// ForwardMessage sends a message that was said in one chat to others, as
// the client's forward does: the message itself is not copied, it is
// pointed at.
func (c *Client) ForwardMessage(msgID string, chatType int, to []Target) error {
	if len(to) == 0 {
		return fmt.Errorf("yh: nothing to forward to")
	}
	body, err := json.Marshal(map[string]any{
		"msgId":    msgID,
		"chatType": chatType,
		"receive":  to,
	})
	if err != nil {
		return err
	}
	data, err := c.do("POST", "/v1/msg/msg-forward", body, false)
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

// Target is a chat a message can be forwarded to.
type Target struct {
	ChatID   string `json:"chatId"`
	ChatType int    `json:"chatType"`
}

// SendQuoted sends a text that answers another: the message it answers is
// pointed at, and the words it said are carried along, as the client does
// so that the quote reads the same to everyone. mentioned are the accounts
// the text names with an @, as SendText takes them, and kind is what the
// text is written in — words, markdown, or markup.
func (c *Client) SendQuoted(chatID string, chatType int, text, quoteID, quoteText string, mentioned []string, kind int) (*Message, error) {
	kind = SendKind(kind)
	id := NewID()
	var content []byte
	content = pbString(content, 1, text)
	for _, who := range mentioned {
		content = pbString(content, 6, who)
	}
	content = pbString(content, 8, quoteText)

	var body []byte
	body = pbString(body, 2, id)
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	body = pbMessage(body, 5, content)
	body = pbUint(body, 6, uint64(kind))
	body = pbString(body, 8, quoteID)

	data, err := c.do("POST", "/v1/msg/send-message", body, true)
	if err != nil {
		return nil, err
	}
	if err := decodeStatusResponse(data).Err(); err != nil {
		return nil, err
	}
	return &Message{
		MsgID:       id,
		ChatID:      chatID,
		ContentType: kind,
		Content:     Content{Text: text, QuoteText: quoteText},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}, nil
}
