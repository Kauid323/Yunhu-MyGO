package yh

// The decoders of messages: who sent one, what it holds, and the listing
// of them. A message from the listing and one the websocket pushes number
// their fields differently, so they are read by two functions.

func decodeSender(b []byte) Sender {
	var s Sender
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			s.ChatID = str(wire, r)
			return true
		case 2:
			s.ChatType = int(num(wire, r))
			return true
		case 3:
			s.Name = str(wire, r)
			return true
		case 4:
			s.AvatarURL = str(wire, r)
			return true
		case 5:
			s.VIP = flag(wire, r)
			return true
		}
		return false
	})
	return s
}

func decodeContent(b []byte) Content {
	var c Content
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			c.Text = str(wire, r)
			return true
		case 3:
			c.ImageURL = str(wire, r)
			return true
		case 4:
			c.FileName = str(wire, r)
			return true
		case 5:
			c.FileURL = str(wire, r)
			return true
		case 6:
			if v, ok := r.value(wire); ok {
				c.MentionedIDs = append(c.MentionedIDs, string(v))
			}
			return true
		case 7:
			c.Form = str(wire, r)
			return true
		case 8:
			c.QuoteText = str(wire, r)
			return true
		case 9:
			c.Image = str(wire, r)
			return true
		case 11:
			c.PostTitle = str(wire, r)
			return true
		case 12:
			c.PostContent = str(wire, r)
			return true
		case 18:
			c.FileSize = int64(num(wire, r))
			return true
		case 19:
			c.VideoURL = str(wire, r)
			return true
		case 21:
			c.AudioURL = str(wire, r)
			return true
		case 22:
			c.AudioTime = int(num(wire, r))
			return true
		case 25:
			c.StickerID = int64(num(wire, r))
			return true
		case 33:
			c.Width = int(num(wire, r))
			return true
		case 34:
			c.Height = int(num(wire, r))
			return true
		case 37:
			c.Tip = str(wire, r)
			return true
		}
		return false
	})
	return c
}

func decodeConversations(b []byte) ([]Conversation, string, Status) {
	var out []Conversation
	var md5 string
	var status Status
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			if v, ok := r.value(wire); ok {
				status = decodeStatus(v)
			}
			return true
		case 2:
			if v, ok := r.value(wire); ok {
				out = append(out, decodeConversation(v))
			}
			return true
		case 4:
			md5 = str(wire, r)
			return true
		}
		return false
	})
	return out, md5, status
}

func decodeConversation(b []byte) Conversation {
	var c Conversation
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			c.ChatID = str(wire, r)
			return true
		case 2:
			c.ChatType = int(num(wire, r))
			return true
		case 3:
			c.Remark = str(wire, r)
			return true
		case 4:
			c.Content = str(wire, r)
			return true
		case 5:
			c.TimestampMS = num(wire, r)
			return true
		case 6:
			c.Unread = int(num(wire, r))
			return true
		case 7:
			c.At = flag(wire, r)
			return true
		case 9:
			c.AvatarURL = str(wire, r)
			return true
		case 11:
			c.DoNotDisturb = flag(wire, r)
			return true
		case 12:
			c.SendTime = num(wire, r)
			return true
		case 13:
			c.VIP = flag(wire, r)
			return true
		case 15:
			c.Name = str(wire, r)
			return true
		case 16:
			c.CertLevel = int(num(wire, r))
			return true
		}
		return false
	})
	return c
}

func decodeMessages(b []byte) ([]*Message, Status) {
	var out []*Message
	var status Status
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			if v, ok := r.value(wire); ok {
				status = decodeStatus(v)
			}
			return true
		case 2:
			if v, ok := r.value(wire); ok {
				out = append(out, decodeMessage(v))
			}
			return true
		}
		return false
	})
	return out, status
}

// decodePushMessage reads a message the websocket pushes, whose fields are
// numbered differently from the listing's.
func decodePushMessage(b []byte) *Message {
	m := &Message{}
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			m.MsgID = str(wire, r)
			return true
		case 2:
			if v, ok := r.value(wire); ok {
				m.Sender = decodeSender(v)
			}
			return true
		case 4:
			m.ChatID = str(wire, r)
			return true
		case 6:
			if v, ok := r.value(wire); ok {
				m.Content = decodeContent(v)
			}
			return true
		case 7:
			m.ContentType = int(num(wire, r))
			return true
		case 8:
			m.TimestampMS = num(wire, r)
			return true
		case 10:
			m.RecalledAt = num(wire, r)
			return true
		case 11:
			m.QuoteMsgID = str(wire, r)
			return true
		case 12:
			m.MsgSeq = num(wire, r)
			return true
		case 14:
			m.EditedAt = num(wire, r)
			return true
		}
		return false
	})
	return m
}

func decodeMessage(b []byte) *Message {
	m := &Message{}
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			m.MsgID = str(wire, r)
			return true
		case 2:
			if v, ok := r.value(wire); ok {
				m.Sender = decodeSender(v)
			}
			return true
		case 3:
			m.Direction = str(wire, r)
			return true
		case 4:
			m.ContentType = int(num(wire, r))
			return true
		case 5:
			if v, ok := r.value(wire); ok {
				m.Content = decodeContent(v)
			}
			return true
		case 6:
			m.TimestampMS = num(wire, r)
			return true
		case 7:
			if v, ok := r.value(wire); ok {
				fields(v, func(f int, w int, rr *reader) bool {
					if f == 2 {
						m.CmdName = str(w, rr)
						return true
					}
					return false
				})
			}
			return true
		case 8:
			m.RecalledAt = num(wire, r)
			return true
		case 9:
			m.QuoteMsgID = str(wire, r)
			return true
		case 10:
			m.MsgSeq = num(wire, r)
			return true
		case 12:
			m.EditedAt = num(wire, r)
			return true
		}
		return false
	})
	return m
}
