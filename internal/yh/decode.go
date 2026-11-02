package yh

// The decoders of what is not a message: the status every response
// carries, the account, and the conversation list. They read the protobuf
// wire format with the reader in wire.go, field by field, and never fail:
// a field that is not there is left as it was.

// decodeStatusResponse reads the status a protobuf route answers with. The
// StatusResponse envelope holds it in field 1; a route that answers with the
// status itself has no such field, so both shapes are read.
func decodeStatusResponse(b []byte) Status {
	var nested []byte
	var wrapped bool
	fields(b, func(field, wire int, r *reader) bool {
		// request_id is field 1 too, but a number: only a message there is
		// the envelope.
		if field == 1 && wire == 2 {
			if v, ok := r.value(wire); ok {
				nested, wrapped = v, true
			}
			return true
		}
		return false
	})
	if wrapped {
		return decodeStatus(nested)
	}
	return decodeStatus(b)
}

func decodeStatus(b []byte) Status {
	var s Status
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 2:
			s.Code = int32(num(wire, r))
			return true
		case 3:
			s.Msg = str(wire, r)
			return true
		}
		return false
	})
	return s
}

func decodeUser(b []byte) (*User, Status) {
	u := &User{}
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
				decodeUserData(v, u)
			}
			return true
		}
		return false
	})
	return u, status
}

func decodeUserData(b []byte, u *User) {
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			u.ID = str(wire, r)
			return true
		case 2:
			u.Name = str(wire, r)
			return true
		case 4:
			u.AvatarURL = str(wire, r)
			return true
		case 6:
			u.Phone = str(wire, r)
			return true
		case 7:
			u.Email = str(wire, r)
			return true
		case 9:
			u.VIP = flag(wire, r)
			return true
		}
		return false
	})
}

// decodeProfile reads what the route that answers about a user says: the
// account itself, the note this account keeps about them, and what they say of
// themselves — three messages of their own, the last two of which may be
// missing.
func decodeProfile(b []byte) (*Profile, Status) {
	p := &Profile{}
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
				decodeProfileData(v, p)
			}
			return true
		}
		return false
	})
	return p, status
}

func decodeProfileData(b []byte, p *Profile) {
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 1:
			p.ID = str(wire, r)
			return true
		case 2:
			p.Name = str(wire, r)
			return true
		case 4:
			p.AvatarURL = str(wire, r)
			return true
		case 13:
			p.VIP = flag(wire, r)
			return true
		case 18:
			if v, ok := r.value(wire); ok {
				decodeRemark(v, p)
			}
			return true
		case 19:
			if v, ok := r.value(wire); ok {
				decodeProfileInfo(v, p)
			}
			return true
		}
		return false
	})
}

// decodeRemark reads the note this account keeps about a user.
func decodeRemark(b []byte, p *Profile) {
	fields(b, func(field, wire int, r *reader) bool {
		if field == 1 {
			p.Remark = str(wire, r)
			return true
		}
		return false
	})
}

// decodeProfileInfo reads what a user says of themselves.
func decodeProfileInfo(b []byte, p *Profile) {
	fields(b, func(field, wire int, r *reader) bool {
		switch field {
		case 2:
			p.Introduction = str(wire, r)
			return true
		case 3:
			p.Gender = int(num(wire, r))
			return true
		case 5:
			p.City = str(wire, r)
			return true
		case 6:
			p.District = str(wire, r)
			return true
		}
		return false
	})
}
