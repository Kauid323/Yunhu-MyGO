package yh

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// ErrNotConnected is what writing asks for before the socket is up.
var ErrNotConnected = errors.New("the connection is not up")

// State is how the live connection stands. It is a code rather than words:
// the window is the only thing that writes words, and it writes them in
// the language the desktop is in.
type State string

const (
	StateOffline  State = "offline"  // nothing has connected yet
	StateOnline   State = "online"   // the socket is up and pushing
	StateRetrying State = "retrying" // it dropped, and is being remade
)

// Event is something the server pushes over the websocket.
type Event struct {
	Cmd     string
	Message *Message // push_message, edit_message, blocked_message
	Stream  Stream   // stream_message
}

// Conn is the account's live connection: messages arrive as they are sent,
// and the connection is remade whenever it drops.
type Conn struct {
	client *Client
	userID string

	mu      sync.Mutex
	ws      *websocket.Conn
	state   State
	onState func(State)
	onEvent func(Event)
}

// NewConn makes a live connection for the account userID.
func (c *Client) NewConn(userID string, onEvent func(Event), onState func(State)) *Conn {
	return &Conn{client: c, userID: userID, state: StateOffline, onEvent: onEvent, onState: onState}
}

func (c *Conn) setState(s State) {
	c.mu.Lock()
	changed := c.state != s
	c.state = s
	c.mu.Unlock()
	if changed && c.onState != nil {
		c.onState(s)
	}
}

// State is the connection's state, as the view shows it.
func (c *Conn) State() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// Run keeps the connection up until ctx is done.
func (c *Conn) Run(ctx context.Context) {
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		err := c.session(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			c.setState(StateRetrying)
		}
		// A connection that held for a while is not a failing one: the
		// next attempt starts over rather than waiting the longest.
		if time.Since(start) > 30*time.Second {
			backoff = time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 20*time.Second {
			backoff *= 2
		}
	}
}

// session is one connection: dial, sign in, then read until it drops.
func (c *Conn) session(ctx context.Context) error {
	dialCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(dialCtx, c.client.WSBase, nil)
	if err != nil {
		return err
	}
	defer ws.Close(websocket.StatusNormalClosure, "")
	ws.SetReadLimit(8 << 20)

	c.mu.Lock()
	c.ws = ws
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.ws = nil
		c.mu.Unlock()
	}()

	login, _ := json.Marshal(map[string]any{
		"seq": NewID(),
		"cmd": "login",
		"data": map[string]string{
			"userId":   c.userID,
			"token":    c.client.Token,
			"platform": c.client.Platform,
			"deviceId": c.client.DeviceID,
		},
	})
	if err := ws.Write(ctx, websocket.MessageText, login); err != nil {
		return err
	}
	// The socket is up and the login frame is sent. The server acknowledges
	// nothing: it only starts pushing, so waiting for an ack would leave
	// the window saying it is not connected while messages arrive.
	c.setState(StateOnline)

	hbCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()
	go c.heartbeat(hbCtx)

	for {
		typ, data, err := ws.Read(ctx)
		if err != nil {
			return err
		}
		if typ == websocket.MessageText {
			continue
		}
		c.dispatch(data)
	}
}

// heartbeat keeps the connection alive while it shows as connected.
func (c *Conn) heartbeat(ctx context.Context) {
	beat, _ := json.Marshal(map[string]any{"seq": NewID(), "cmd": "heartbeat", "data": map[string]string{}})
	t := time.NewTicker(25 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := c.write(beat); err != nil {
				return
			}
		}
	}
}

func (c *Conn) write(b []byte) error {
	c.mu.Lock()
	ws := c.ws
	c.mu.Unlock()
	if ws == nil {
		return ErrNotConnected
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return ws.Write(ctx, websocket.MessageText, b)
}

// dispatch turns a binary frame into an event.
func (c *Conn) dispatch(b []byte) {
	env := decodeWsWrapper(b)
	if env.Cmd == "" {
		return
	}
	ev := Event{Cmd: env.Cmd}
	switch env.Cmd {
	case "push_message", "edit_message", "blocked_message":
		if len(env.Value) == 0 {
			return
		}
		ev.Message = decodePushMessage(env.Value)
	case "stream_message":
		ev.Stream = decodeStream(env.Value)
	case "heartbeat_ack":
		return
	}
	if c.onEvent != nil {
		c.onEvent(ev)
	}
}

// Close drops the connection.
func (c *Conn) Close() {
	c.mu.Lock()
	ws := c.ws
	c.ws = nil
	c.mu.Unlock()
	if ws != nil {
		ws.Close(websocket.StatusNormalClosure, "")
	}
}
