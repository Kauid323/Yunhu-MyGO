package yh

// A frame from the websocket: the envelope it comes in, and the piece of a
// message still being written that a bot streams.

// WsEnvelope is a frame from the websocket: its command and payload.
type WsEnvelope struct {
	Cmd   string
	Value []byte
}

// decodeWsWrapper reads the info and the Any payload of a websocket frame.
func decodeWsWrapper(b []byte) WsEnvelope {
	var e WsEnvelope
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			if v, ok := r.value(wire); ok {
				fields(v, func(f int, w int, rr *reader) bool {
					if f == 2 {
						e.Cmd = str(w, rr)
						return true
					}
					return false
				})
			}
			return true
		case 2:
			if v, ok := r.value(wire); ok {
				fields(v, func(f int, w int, rr *reader) bool {
					if f == 2 {
						e.Value, _ = rr.value(w)
						return true
					}
					return false
				})
			}
			return true
		}
		return false
	})
	return e
}

// Stream is a piece of a message still being written, as a bot streams it.
type Stream struct {
	MsgID   string
	ChatID  string
	Content string
}

func decodeStream(b []byte) Stream {
	var s Stream
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			s.MsgID = str(wire, r)
			return true
		case 3:
			s.ChatID = str(wire, r)
			return true
		case 4:
			s.Content = str(wire, r)
			return true
		}
		return false
	})
	return s
}
