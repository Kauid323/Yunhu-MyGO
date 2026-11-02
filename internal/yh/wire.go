// Package yh talks to the Yunhu (云湖) chat API.
//
// The endpoints documented at https://yh-api.yyyyt.top/api/v1 answer with
// protobuf on most routes, so this package encodes and decodes the few
// messages it needs on its own: no generated code, no protobuf dependency.
package yh

// Protocol-buffers wire format: varints, length-delimited fields, and
// skipping what a message does not use.

type reader struct {
	b []byte
	i int
}

func (r *reader) more() bool { return r.i < len(r.b) }

func (r *reader) varint() (uint64, bool) {
	var v uint64
	var shift uint
	for r.i < len(r.b) {
		c := r.b[r.i]
		r.i++
		if c < 0x80 {
			if shift >= 64 {
				return 0, false
			}
			return v | uint64(c)<<shift, true
		}
		v |= uint64(c&0x7f) << shift
		shift += 7
		if shift >= 70 {
			return 0, false
		}
	}
	return 0, false
}

func (r *reader) length() ([]byte, bool) {
	n, ok := r.varint()
	if !ok || n > uint64(len(r.b)-r.i) {
		return nil, false
	}
	b := r.b[r.i : r.i+int(n)]
	r.i += int(n)
	return b, true
}

func (r *reader) skip(wire int) bool {
	switch wire {
	case 0:
		_, ok := r.varint()
		return ok
	case 1:
		if len(r.b)-r.i < 8 {
			return false
		}
		r.i += 8
		return true
	case 2:
		_, ok := r.length()
		return ok
	case 5:
		if len(r.b)-r.i < 4 {
			return false
		}
		r.i += 4
		return true
	}
	return false
}

// fields walks the fields of a message, handing each to fn with a reader
// positioned at its value. fn reports whether it consumed the value; what
// it leaves unread is skipped.
func fields(b []byte, fn func(field, wire int, r *reader) bool) {
	r := &reader{b: b}
	for r.more() {
		key, ok := r.varint()
		if !ok {
			return
		}
		field, wire := int(key>>3), int(key&7)
		start := r.i
		if fn(field, wire, r) {
			continue
		}
		r.i = start
		if !r.skip(wire) {
			return
		}
	}
}

// value returns the string or bytes of a length-delimited field.
func (r *reader) value(wire int) ([]byte, bool) {
	if wire != 2 {
		return nil, false
	}
	return r.length()
}

func str(wire int, r *reader) string {
	v, ok := r.value(wire)
	if !ok {
		return ""
	}
	return string(v)
}

func num(wire int, r *reader) uint64 {
	if wire != 0 {
		return 0
	}
	v, _ := r.varint()
	return v
}

func flag(wire int, r *reader) bool {
	if wire != 0 {
		return false
	}
	v, _ := r.varint()
	return v != 0
}

// Encoding.

func pbKey(dst []byte, field, wire int) []byte {
	return pbVarintRaw(dst, uint64(field)<<3|uint64(wire))
}

func pbVarintRaw(dst []byte, v uint64) []byte {
	for v >= 0x80 {
		dst = append(dst, byte(v)|0x80)
		v >>= 7
	}
	return append(dst, byte(v))
}

// pbUint appends a varint field, skipping it when zero as proto3 does.
func pbUint(dst []byte, field int, v uint64) []byte {
	if v == 0 {
		return dst
	}
	dst = pbKey(dst, field, 0)
	return pbVarintRaw(dst, v)
}

func pbBool(dst []byte, field int, v bool) []byte {
	if !v {
		return dst
	}
	return pbUint(dst, field, 1)
}

func pbString(dst []byte, field int, s string) []byte {
	if s == "" {
		return dst
	}
	dst = pbKey(dst, field, 2)
	dst = pbVarintRaw(dst, uint64(len(s)))
	return append(dst, s...)
}

func pbMessage(dst []byte, field int, sub []byte) []byte {
	if len(sub) == 0 {
		return dst
	}
	dst = pbKey(dst, field, 2)
	dst = pbVarintRaw(dst, uint64(len(sub)))
	return append(dst, sub...)
}
