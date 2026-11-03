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
