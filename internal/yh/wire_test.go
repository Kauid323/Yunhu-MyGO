package yh

import (
	"encoding/hex"
	"testing"
)

// TestDecodeStatusResponse checks that the StatusResponse envelope is
// unwrapped: the status lives in field 1, and reading the outer fields
// instead reports a failure for a message that was sent.
func TestDecodeStatusResponse(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		code int32
		msg  string
	}{
		{
			// The reply to a send to an unknown chat, as the server sends it:
			// StatusResponse{status{code:4404, msg:"用户/群/机器人不是您的好友"}}.
			name: "wrapped failure",
			hex:  "0a2b10b4221a26e794a8e688b72fe7bea42fe69cbae599a8e4babae4b88de698afe682a8e79a84e5a5bde58f8b",
			code: 4404,
			msg:  "用户/群/机器人不是您的好友",
		},
		{
			// StatusResponse{status{code:1, msg:"success"}}.
			name: "wrapped success",
			hex:  "0a0b10011a0773756363657373",
			code: 1,
			msg:  "success",
		},
		{
			// A route that answers with the status itself, no envelope.
			name: "bare status",
			hex:  "10011a0773756363657373",
			code: 1,
			msg:  "success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatal(err)
			}
			got := decodeStatusResponse(b)
			if got.Code != tt.code || got.Msg != tt.msg {
				t.Errorf("got code=%d msg=%q, want code=%d msg=%q",
					got.Code, got.Msg, tt.code, tt.msg)
			}
			if err := got.Err(); (err == nil) != (tt.code == 1) {
				t.Errorf("Err() = %v for code %d", err, tt.code)
			}
		})
	}
}

// TestOldestFirst checks that a newest-first page is turned round.
func TestOldestFirst(t *testing.T) {
	msgs := []*Message{
		{MsgID: "c", MsgSeq: 3},
		{MsgID: "b", MsgSeq: 2},
		{MsgID: "a", MsgSeq: 1},
	}
	got := oldestFirst(msgs)
	want := []string{"a", "b", "c"}
	for i, id := range want {
		if got[i].MsgID != id {
			t.Fatalf("index %d is %q, want %q (order: %v)", i, got[i].MsgID, id, ids(got))
		}
	}
}

func ids(msgs []*Message) []string {
	out := make([]string, len(msgs))
	for i, m := range msgs {
		out[i] = m.MsgID
	}
	return out
}
