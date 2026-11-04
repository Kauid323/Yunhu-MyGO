package yh

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
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

// SetProxy makes every request go through a proxy: kind is "http" or
// "socks5", and addr is the host and port it listens on ("127.0.0.1:7890").
// An empty kind puts the requests back on the direct way, as they were.
//
// The client is replaced rather than changed under the requests already
// running: a transport is not safe to swap while it is dialling.
func (c *Client) SetProxy(kind, addr string) error {
	addr = strings.TrimSpace(addr)
	var tr *http.Transport
	switch kind {
	case "":
	case "http", "https":
		u, err := url.Parse(addr)
		if err != nil {
			return err
		}
		if u.Scheme == "" {
			u.Scheme = "http"
		}
		if u.Host == "" {
			return fmt.Errorf("yh: proxy %q has no host", addr)
		}
		tr = &http.Transport{Proxy: http.ProxyURL(u)}
	case "socks5":
		d, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
		if err != nil {
			return err
		}
		tr = &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return d.Dial(network, address)
			},
		}
	default:
		return fmt.Errorf("yh: unknown proxy kind %q", kind)
	}
	c.mu.Lock()
	if tr == nil {
		c.hc = &http.Client{Timeout: 25 * time.Second}
	} else {
		c.hc = &http.Client{Timeout: 25 * time.Second, Transport: tr}
	}
	c.mu.Unlock()
	return nil
}

// HTTP is the client the requests go out with, proxy and all: the socket
// the chat is read over is opened with it too, so that one setting covers
// both.
func (c *Client) HTTP() *http.Client { return c.http() }

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

// MessagesAround returns the page a message is on, the message itself among
// them: the route lists from a message and its sequence rather than from
// the newest one, so a history can be opened where something was found
// instead of being paged back to it a page at a time.
//
// A sequence the route does not know is said as zero, which it reads as
// the message id alone being enough.
func (c *Client) MessagesAround(chatID string, chatType, count int, msgID string, seq uint64) ([]*Message, error) {
	var body []byte
	body = pbUint(body, 3, seq)
	body = pbUint(body, 4, uint64(chatType))
	body = pbString(body, 5, chatID)
	body = pbUint(body, 6, 0)
	body = pbUint(body, 7, uint64(count))
	body = pbString(body, 8, msgID)
	data, err := c.do("POST", "/v1/msg/list-message-by-mid-seq", body, true)
	if err != nil {
		return nil, err
	}
	msgs, status := decodeMessages(data)
	if err := status.Err(); err != nil {
		return nil, err
	}
	return oldestFirst(msgs), nil
}

// MessagesAfter returns the messages the server lists after a sequence: the
// page below a history that was opened in the middle of itself. The route
// takes no count of its own, so the page is as long as it is.
func (c *Client) MessagesAfter(chatID string, chatType int, seq uint64) ([]*Message, error) {
	var body []byte
	body = pbUint(body, 3, seq)
	body = pbUint(body, 4, uint64(chatType))
	body = pbString(body, 5, chatID)
	data, err := c.do("POST", "/v1/msg/list-message-by-seq", body, true)
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

// SendKind is the number the route is given for what a text is: words,
// markdown, or markup. One that says nothing means words, which is what a
// composer that has not been asked for a format holds.
func SendKind(kind int) int {
	if kind == 0 {
		return ContentText
	}
	return kind
}

// SendText sends a text message, and returns it as it will show locally.
// mentioned are the accounts the text names with an @: the server reads them
// from the content, and the chat that receives the message marks those names
// rather than showing them as words. kind is what the text is written in —
// words, markdown, or markup — and decides who reads it as more than words.
func (c *Client) SendText(chatID string, chatType int, text string, mentioned []string, kind int) (*Message, error) {
	kind = SendKind(kind)
	id := NewID()
	var content []byte
	content = pbString(content, 1, text)
	for _, who := range mentioned {
		content = pbString(content, 6, who)
	}

	var body []byte
	body = pbString(body, 2, id)
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	body = pbMessage(body, 5, content)
	body = pbUint(body, 6, uint64(kind))

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

// SendImage sends a picture to a chat: the key the bucket gave it, and the
// size it was sent in, which the chat uses to lay the message out before the
// picture itself is here.
//
// The key goes in the content's picture field rather than in the address
// field: a message that came back from the server carries the key there and
// the whole address in the other, so the address is what the server makes of
// the key, and sending an address of our own would be sending it twice.
func (c *Client) SendImage(chatID string, chatType int, key string, width, height, size int64) (*Message, error) {
	id := NewID()
	var content []byte
	content = pbString(content, 9, key)
	content = pbUint(content, 18, uint64(max(size, 0)))
	content = pbUint(content, 33, uint64(max(width, 0)))
	content = pbUint(content, 34, uint64(max(height, 0)))

	var body []byte
	body = pbString(body, 2, id)
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	body = pbMessage(body, 5, content)
	body = pbUint(body, 6, ContentImage)

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
		ContentType: ContentImage,
		Content: Content{
			Image:    key,
			ImageURL: MediaURL("image", key),
			Width:    int(width),
			Height:   int(height),
		},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}, nil
}

// SendFile sends a file to a chat: the key the bucket gave it, the name it
// had, and its size, which is what the message shows besides the name.
func (c *Client) SendFile(chatID string, chatType int, key, name string, size int64) (*Message, error) {
	id := NewID()
	var content []byte
	content = pbString(content, 4, name)
	content = pbString(content, 5, key)
	content = pbUint(content, 18, uint64(max(size, 0)))

	var body []byte
	body = pbString(body, 2, id)
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	body = pbMessage(body, 5, content)
	body = pbUint(body, 6, ContentFile)

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
		ContentType: ContentFile,
		Content: Content{
			FileName: name,
			FileURL:  MediaURL("file", key),
			FileSize: size,
		},
		TimestampMS: uint64(time.Now().UnixMilli()),
	}, nil
}

// SendVideo sends a clip to a chat. The key goes in the content's video
// field, and the same key in the media beside it: the media is what tells
// the chat the size and the kind of the clip, and a video sent without it
// is one the player has nothing to open. Both are the key, not an address —
// the server makes the address of it.
func (c *Client) SendVideo(chatID string, chatType int, key, name string, size int64) (*Message, error) {
	id := NewID()
	ext := uploadExt(name)
	var content []byte
	content = pbString(content, 19, key)
	content = pbUint(content, 18, uint64(max(size, 0)))

	var media []byte
	media = pbString(media, 1, key)
	media = pbString(media, 3, "video/"+ext)
	media = pbUint(media, 7, uint64(max(size, 0)))
	media = pbString(media, 8, key)
	media = pbString(media, 9, ext)

	var body []byte
	body = pbString(body, 2, id)
	body = pbString(body, 3, chatID)
	body = pbUint(body, 4, uint64(chatType))
	body = pbMessage(body, 5, content)
	body = pbUint(body, 6, ContentVideo)
	body = pbMessage(body, 9, media)

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
		ContentType: ContentVideo,
		Content: Content{
			VideoURL: MediaURL("video", key),
			FileName: name,
			FileSize: size,
		},
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
