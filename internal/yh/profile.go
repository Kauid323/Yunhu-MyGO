package yh

import (
	"fmt"
)

// Profile is what is known about a user besides the messages they send: what
// they call themselves, what they say of themselves, and what the account has
// noted about them.
type Profile struct {
	ID        string
	Name      string
	AvatarURL string
	// Remark is the name this account keeps for them, which is empty when
	// it keeps none.
	Remark string
	// Introduction is what they say of themselves, and Gender is 1 for a
	// man, 2 for a woman and 3 for anything else.
	Introduction string
	Gender       int
	// City and District are where they say they are, which may be empty.
	City, District string
	VIP            bool
}

// UserProfile asks what is known about a user, by their id.
func (c *Client) UserProfile(id string) (*Profile, error) {
	if id == "" {
		return nil, ErrNoAccount
	}
	var body []byte
	body = pbString(body, 2, id)
	data, err := c.do("POST", "/v1/user/get-user", body, true)
	if err != nil {
		return nil, err
	}
	p, status := decodeProfile(data)
	if err := status.Err(); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: no such user", ErrResponse)
	}
	return p, nil
}
