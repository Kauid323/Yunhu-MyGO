package yh

// Group is what a group chat says of itself: what it is called, what it is
// about, how many are in it, and where it is filed.
type Group struct {
	ID        string
	Name      string
	AvatarURL string
	Intro     string
	Headcount int
	Category  string
}

// GroupInfo asks a group chat what it says of itself.
func (c *Client) GroupInfo(id string) (*Group, error) {
	var body []byte
	body = pbString(body, 2, id)

	data, err := c.do("POST", "/v1/group/info", body, true)
	if err != nil {
		return nil, err
	}
	g := &Group{ID: id}
	var status Status
	fields(data, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			if v, ok := r.value(wire); ok {
				status = decodeStatus(v)
			}
			return true
		case 2:
			if v, ok := r.value(wire); ok {
				decodeGroupData(v, g)
			}
			return true
		}
		return false
	})
	if err := status.Err(); err != nil {
		return nil, err
	}
	return g, nil
}

// Member is one person in a group chat: who they are, and how much they may
// do there.
type Member struct {
	ID        string
	Name      string
	AvatarURL string
	// Permission is how much the member may do: 100 for the owner, 2 for an
	// administrator, and nothing for everyone else.
	Permission int
}

// Members asks a group chat who is in it. The route pages, and a page is
// asked for by its number from one: what a mention needs is the people the
// reader is talking to, which are the first page of them.
func (c *Client) Members(id string, page, size int) ([]Member, error) {
	var data []byte
	data = pbUint(data, 1, uint64(max(size, 1)))
	data = pbUint(data, 2, uint64(max(page, 1)))
	var body []byte
	body = pbMessage(body, 2, data)
	body = pbString(body, 3, id)

	answer, err := c.do("POST", "/v1/group/list-member", body, true)
	if err != nil {
		return nil, err
	}
	var status Status
	var out []Member
	fields(answer, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			if v, ok := r.value(wire); ok {
				status = decodeStatus(v)
			}
			return true
		case 2:
			if v, ok := r.value(wire); ok {
				if m, ok := decodeMember(v); ok {
					out = append(out, m)
				}
			}
			return true
		}
		return false
	})
	if err := status.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// decodeMember reads one member of a group: the person is a message of its
// own, and the standing is beside it.
func decodeMember(b []byte) (Member, bool) {
	var m Member
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 2:
			if v, ok := r.value(wire); ok {
				fields(v, func(field, wire int, r *reader) bool {
					switch field {
					case 1:
						m.ID = str(wire, r)
						return true
					case 2:
						m.Name = str(wire, r)
						return true
					case 4:
						m.AvatarURL = str(wire, r)
						return true
					}
					return false
				})
			}
			return true
		case 3:
			m.Permission = int(num(wire, r))
			return true
		}
		return false
	})
	return m, m.ID != ""
}

// decodeGroupData reads what a group says of itself. The fields are the ones
// of the route's own message, numbered as its protobuf numbers them.
func decodeGroupData(b []byte, g *Group) {
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			g.ID = str(wire, r)
			return true
		case 2:
			g.Name = str(wire, r)
			return true
		case 3:
			g.AvatarURL = str(wire, r)
			return true
		case 5:
			g.Intro = str(wire, r)
			return true
		case 6:
			g.Headcount = int(num(wire, r))
			return true
		case 11:
			g.Category = str(wire, r)
			return true
		}
		return false
	})
}
