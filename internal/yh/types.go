package yh

import (
	"fmt"
	"time"
)

// Chat types, as the API numbers them.
const (
	ChatUser  = 1
	ChatGroup = 2
	ChatBot   = 3
)

// Content types, as the API numbers them.
const (
	ContentText     = 1
	ContentImage    = 2
	ContentMarkdown = 3
	ContentFile     = 4
	ContentForm     = 5
	ContentPost     = 6
	ContentSticker  = 7
	ContentHTML     = 8
	ContentTip      = 9
	// ContentVideo is a clip sent in a chat. The number is the client's
	// own: the route's comment lists the others and leaves this one out,
	// but a video that comes back from the server carries 10.
	ContentVideo = 10
	ContentAudio = 11
	ContentCall  = 13
)

// Status is the code and message every response carries.
type Status struct {
	Code int32
	Msg  string
}

// Err returns the status as an error, or nil when the code is 1.
func (s Status) Err() error {
	if s.Code == 1 {
		return nil
	}
	return &APIError{Code: int(s.Code), Msg: s.Msg}
}

// APIError is a failed request, with the server's code and message.
type APIError struct {
	Code int
	Msg  string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s (code %d)", e.Msg, e.Code)
}

// User is the signed-in account.
type User struct {
	ID        string
	Name      string
	AvatarURL string
	Email     string
	Phone     string
	Coin      float64
	VIP       bool
	// Introduction is what the account says of itself, as the profile has
	// it: the route the account comes from does not carry it, so it is
	// asked for on its own and may be empty.
	Introduction string
}

// Conversation is a chat in the conversation list.
type Conversation struct {
	ChatID       string
	ChatType     int
	Remark       string
	Name         string
	Content      string // last message preview
	TimestampMS  uint64 // when the chat entered the list
	SendTime     uint64 // when the last message was sent, in seconds
	Unread       int
	At           bool
	AvatarURL    string
	DoNotDisturb bool
	VIP          bool
	CertLevel    int
}

// Title is the name to show: the remark when there is one.
func (c Conversation) Title() string {
	if c.Remark != "" {
		return c.Remark
	}
	if c.Name != "" {
		return c.Name
	}
	return c.ChatID
}

// Time is when the conversation last moved.
func (c Conversation) Time() time.Time {
	if c.SendTime > 0 {
		return time.Unix(int64(c.SendTime), 0)
	}
	return time.UnixMilli(int64(c.TimestampMS))
}

// Sender is who sent a message.
type Sender struct {
	ChatID    string
	ChatType  int
	Name      string
	AvatarURL string
	VIP       bool
}

// Content is a message's payload: text, an image, a file, and so on.
type Content struct {
	Text         string
	ImageURL     string
	Image        string
	FileName     string
	FileURL      string
	FileSize     int64
	VideoURL     string
	AudioURL     string
	AudioTime    int
	MentionedIDs []string
	Form         string
	QuoteText    string
	QuoteMsgID   string
	PostTitle    string
	PostContent  string
	Tip          string
	StickerID    int64
	Width        int
	Height       int
}

// Message is one message of a conversation.
type Message struct {
	MsgID       string
	ChatID      string
	Sender      Sender
	Direction   string
	ContentType int
	Content     Content
	TimestampMS uint64
	MsgSeq      uint64
	QuoteMsgID  string
	RecalledAt  uint64
	EditedAt    uint64
	CmdName     string

	// Pending is set on a message still on its way to the server, Failed
	// when it did not get there.
	Pending bool
	Failed  string
}

// Mine reports whether the signed-in account sent the message.
func (m *Message) Mine(me string) bool {
	return me != "" && m.Sender.ChatID == me
}

// Time is when the message was sent.
func (m *Message) Time() time.Time {
	return time.UnixMilli(int64(m.TimestampMS))
}
