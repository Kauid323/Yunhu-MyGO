package yh

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Default hosts, as the API documentation gives them.
const (
	DefaultBase   = "https://chat-go.jwzhd.com"
	DefaultWSBase = "wss://chat-ws-go.jwzhd.com/ws"
)

// Client calls the Yunhu API with one account's token.
type Client struct {
	Base     string // https://chat-go.jwzhd.com
	WSBase   string // wss://chat-ws-go.jwzhd.com/ws
	Token    string
	DeviceID string
	Platform string

	mu  sync.Mutex
	hc  *http.Client
	md5 string
}

// New returns a client for the default hosts.
func New() *Client {
	return &Client{
		Base:     DefaultBase,
		WSBase:   DefaultWSBase,
		Platform: "windows",
		DeviceID: NewID(),
		hc:       &http.Client{Timeout: 25 * time.Second},
	}
}

// NewID makes a random identifier, as message ids and device ids are.
func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func (c *Client) http() *http.Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hc == nil {
		c.hc = &http.Client{Timeout: 25 * time.Second}
	}
	return c.hc
}

// do sends a request to path, with the account's token, and returns the
// body. Protobuf routes take and answer with protobuf; the others JSON.
func (c *Client) do(method, path string, body []byte, proto bool) ([]byte, error) {
	url := strings.TrimRight(c.Base, "/") + path
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		return nil, err
	}
	if body != nil {
		if proto {
			req.Header.Set("Content-Type", "application/x-protobuf")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if c.Token != "" {
		req.Header.Set("token", c.Token)
	}
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrResponse, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrStatus, resp.Status)
	}
	return data, nil
}

// jsonStatus is the JSON envelope of the routes that answer with JSON.
type jsonStatus struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (s jsonStatus) err() error {
	if s.Code == 1 {
		return nil
	}
	return &APIError{Code: s.Code, Msg: s.Msg}
}

// LoginEmail signs in with an email and a password, and returns the token.
func (c *Client) LoginEmail(email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"deviceId": c.DeviceID,
		"platform": c.Platform,
	})
	return c.login("/v1/user/email-login", body)
}

// LoginToken checks a token, and returns the account it belongs to.
func (c *Client) LoginToken(token string) (*User, error) {
	c.Token = token
	return c.UserInfo()
}

// UserInfo returns the signed-in account.
func (c *Client) UserInfo() (*User, error) {
	data, err := c.do("GET", "/v1/user/info", nil, true)
	if err != nil {
		return nil, err
	}
	u, status := decodeUser(data)
	if err := status.Err(); err != nil {
		return nil, err
	}
	if u.ID == "" {
		return nil, ErrNoAccount
	}
	return u, nil
}

// Introduction returns what the account says of itself, which the route the
// account is read from does not carry: it is the profile's, and the profile
// is a JSON route of its own.
func (c *Client) Introduction() (string, error) {
	data, err := c.do("POST", "/v1/user/get-user-data", []byte("{}"), false)
	if err != nil {
		return "", err
	}
	var out struct {
		jsonStatus
		Data struct {
			Data struct {
				Introduction string `json:"introduction"`
			} `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("%w: %v", ErrResponse, err)
	}
	if err := out.err(); err != nil {
		return "", err
	}
	return out.Data.Data.Introduction, nil
}

// Conversations lists the account's chats. md5 is the one the last listing
// returned: when nothing changed the server answers with an empty list.
func (c *Client) Conversations(md5 string) ([]Conversation, string, error) {
	var body []byte
	body = pbString(body, 2, md5)
	data, err := c.do("POST", "/v1/conversation/list", body, true)
	if err != nil {
		return nil, "", err
	}
	list, sum, status := decodeConversations(data)
	if err := status.Err(); err != nil {
		return nil, "", err
	}
	return list, sum, nil
}

// Messages returns the last count messages of a chat, oldest first.
func (c *Client) Messages(chatID string, chatType, count int) ([]*Message, error) {
	return c.messages(chatID, chatType, count, "")
}

// MessagesBefore returns count messages older than the one before, oldest
// first: the server lists newest first and pages backwards from a message
// id it does not include.
func (c *Client) MessagesBefore(chatID string, chatType, count int, before string) ([]*Message, error) {
	return c.messages(chatID, chatType, count, before)
}

func (c *Client) messages(chatID string, chatType, count int, before string) ([]*Message, error) {
	var body []byte
	body = pbUint(body, 2, uint64(count))
	if before != "" {
		body = pbString(body, 3, before)
	}
	body = pbUint(body, 4, uint64(chatType))
	body = pbString(body, 5, chatID)
	data, err := c.do("POST", "/v1/msg/list-message", body, true)
	if err != nil {
		return nil, err
	}
	msgs, status := decodeMessages(data)
	if err := status.Err(); err != nil {
		return nil, err
	}
	return oldestFirst(msgs), nil
}

// oldestFirst turns the server's newest-first page into the order a chat
// reads: oldest at the top.
func oldestFirst(msgs []*Message) []*Message {
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs
}

// SendText sends a text message, and returns it as it will show locally.
func (c *Client) SendText(chatID string, chatType int, text string) (*Message, error) {
	id := NewID()
	var content []byte
	content = pbString(content, 1, text)

	var body []byte
	body = pbString(body, 2, id)
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	body = pbMessage(body, 5, content)
	body = pbUint(body, 6, ContentText)

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
		ContentType: ContentText,
		Content:     Content{Text: text},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}, nil
}

// SendSticker sends one of the account's saved stickers to a chat: the
// picture the host keeps for it, and the id of the sticker itself, which the
// message carries so that it can be added back to the collection.
func (c *Client) SendSticker(chatID string, chatType int, image, expressionID string) (*Message, error) {
	id := NewID()
	var content []byte
	content = pbString(content, 9, image)
	content = pbString(content, 15, expressionID)

	var body []byte
	body = pbString(body, 2, id)
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	body = pbMessage(body, 5, content)
	body = pbUint(body, 6, ContentSticker)

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
		ContentType: ContentSticker,
		Content:     Content{Image: image},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}, nil
}

// MarkRead clears a conversation's unread count.
func (c *Client) MarkRead(chatID string) error {
	body, _ := json.Marshal(map[string]string{"chatId": chatID})
	data, err := c.do("POST", "/v1/conversation/dismiss-notification", body, false)
	if err != nil {
		return err
	}
	var res jsonStatus
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	return res.err()
}

// Logout signs the account out on this device.
func (c *Client) Logout() error {
	body, _ := json.Marshal(map[string]string{"device-id": c.DeviceID})
	_, err := c.do("POST", "/v1/user/logout", body, false)
	return err
}
