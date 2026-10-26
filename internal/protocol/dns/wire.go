// Package dns owns bounded DNS wire encoding, strict decoding and one-query
// operations. It owns no path, authority, retry, result envelope or persistence.
package dns

import (
	"encoding/binary"
	"errors"
	"net/netip"
	"slices"
	"strings"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

const MaxUDP = 4096
const MaxTCP = 65535

var ErrMalformed = errors.New("dns: malformed response")
var ErrMismatch = errors.New("dns: query mismatch")

type Question struct {
	ID         uint16
	Name, Type string
}

func qtype(s string) uint16 {
	if s == "A" {
		return 1
	}
	if s == "AAAA" {
		return 28
	}
	return 0
}

// Query emits one ordinary recursive IN query, without EDNS or identifying
// metadata. The caller supplies a fresh unpredictable ID for each operation.
// requirement-ref: REQ-DNS-01-0008
// requirement-ref: REQ-DNS-01-0048
func Query(q Question) ([]byte, error) {
	if qtype(q.Type) == 0 || !validName(q.Name) {
		return nil, ErrMalformed
	}
	b := make([]byte, 12, 12+len(q.Name)+6)
	binary.BigEndian.PutUint16(b, q.ID)
	b[2] = 1
	b[5] = 1
	for _, l := range strings.Split(q.Name, ".") {
		b = append(b, byte(len(l)))
		b = append(b, l...)
	}
	b = append(b, 0, 0, byte(qtype(q.Type)), 0, 1)
	return b, nil
}
func validName(n string) bool {
	if n == "" || len(n) > 253 {
		return false
	}
	for _, l := range strings.Split(n, ".") {
		if len(l) == 0 || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' {
			return false
		}
		for _, c := range l {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}

// name follows only backward compression pointers, with bounded label count,
// expanded length and pointer hops. No attacker-controlled recursion or slices.
func name(b []byte, at int) (string, int, error) {
	next := -1
	size, hops := 0, 0
	var labels []string
	for {
		if at < 0 || at >= len(b) || hops > 128 {
			return "", 0, ErrMalformed
		}
		hops++
		n := int(b[at])
		at++
		if n == 0 {
			if next < 0 {
				next = at
			}
			break
		}
		if n&0xc0 == 0xc0 {
			if at >= len(b) {
				return "", 0, ErrMalformed
			}
			target := (n&63)<<8 | int(b[at])
			if target >= at-1 || target < 12 {
				return "", 0, ErrMalformed
			}
			if next < 0 {
				next = at + 1
			}
			at = target
			continue
		}
		if n > 63 || at+n > len(b) {
			return "", 0, ErrMalformed
		}
		size += n + 1
		if size > 254 {
			return "", 0, ErrMalformed
		}
		// Escape non-LDH octets so a literal dot inside one DNS label cannot
		// masquerade as a separator in the correlated or retained name. Unrelated
		// wire names need not obey the hostname input grammar.
		var label strings.Builder
		for _, v := range b[at : at+n] {
			if v >= 'A' && v <= 'Z' {
				v += 'a' - 'A'
			}
			if v >= 'a' && v <= 'z' || v >= '0' && v <= '9' || v == '-' {
				label.WriteByte(v)
			} else {
				label.WriteByte('\\')
				label.WriteByte('0' + v/100)
				label.WriteByte('0' + v/10%10)
				label.WriteByte('0' + v%10)
			}
		}
		labels = append(labels, label.String())
		at += n
	}
	n := strings.Join(labels, ".")

	return n, next, nil
}

type record struct {
	owner, cname string
	address      netip.Addr
	kind         uint16
}

// Parse validates the entire frame, including skipped sections and EDNS TLVs.
// Only answers on the queried name's bounded CNAME chain are retained. Unknown
// types, TTLs, authority/additional records and raw messages never escape.
// requirement-ref: REQ-DNS-01-0010
// requirement-ref: REQ-DNS-01-0022
// requirement-ref: REQ-DNS-01-0036
// requirement-ref: REQ-DNS-01-0040
// requirement-ref: REQ-OBS-0046
func Parse(b []byte, q Question, maxBytes int) (observation.DNSResponse, error) {
	bad := func() (observation.DNSResponse, error) { return observation.DNSResponse{}, ErrMalformed }
	if len(b) >= 2 && binary.BigEndian.Uint16(b) != q.ID {
		return observation.DNSResponse{}, ErrMismatch
	}
	if len(b) > maxBytes || len(b) < 12 || maxBytes < 12 || maxBytes > MaxTCP {
		return bad()
	}
	if binary.BigEndian.Uint16(b) != q.ID {
		return observation.DNSResponse{}, ErrMismatch
	}
	flags := binary.BigEndian.Uint16(b[2:])
	if flags&0x8000 == 0 || flags&0x7800 != 0 || flags&0x0240 != 0 || binary.BigEndian.Uint16(b[4:]) != 1 {
		return bad()
	}
	n, off, e := name(b, 12)
	if e != nil || off+4 > len(b) {
		return bad()
	}
	if n != q.Name || binary.BigEndian.Uint16(b[off:]) != qtype(q.Type) || binary.BigEndian.Uint16(b[off+2:]) != 1 {
		return observation.DNSResponse{}, ErrMismatch
	}
	off += 4
	counts := [3]int{int(binary.BigEndian.Uint16(b[6:])), int(binary.BigEndian.Uint16(b[8:])), int(binary.BigEndian.Uint16(b[10:]))}
	total := counts[0] + counts[1] + counts[2]
	if total > len(b)/11 {
		return bad()
	}
	var records []record
	code := uint16(flags & 15)
	opt := false
	for section, count := range counts {
		for range count {
			owner, pos, err := name(b, off)
			if err != nil || pos+10 > len(b) {
				return bad()
			}
			kind, class := binary.BigEndian.Uint16(b[pos:]), binary.BigEndian.Uint16(b[pos+2:])
			ttl := binary.BigEndian.Uint32(b[pos+4:])
			length := int(binary.BigEndian.Uint16(b[pos+8:]))
			start, end := pos+10, pos+10+length
			if end > len(b) {
				return bad()
			}
			off = end
			r := record{owner: owner, kind: kind}
			switch kind {
			case 1:
				if length != 4 {
					return bad()
				}
				r.address = netip.AddrFrom4([4]byte(b[start:end]))
			case 28:
				if length != 16 {
					return bad()
				}
				r.address = netip.AddrFrom16([16]byte(b[start:end]))
			case 5:
				target, next, err := name(b, start)
				if err != nil || next != end || target == "" {
					return bad()
				}
				r.cname = target
			case 41:
				if section != 2 || owner != "" || opt {
					return bad()
				}
				opt = true
				code |= uint16(ttl>>24) << 4
				for j := start; j < end; {
					if j+4 > end {
						return bad()
					}
					sz := int(binary.BigEndian.Uint16(b[j+2:]))
					j += 4 + sz
					if j > end {
						return bad()
					}
				}
			}
			if section == 0 && class == 1 && (kind == qtype(q.Type) || kind == 5) {
				records = append(records, r)
			}
		}
	}
	if off != len(b) {
		return bad()
	}
	rcode, _ := observation.DNSResponseCode(code, true)
	out := observation.DNSResponse{RCode: rcode, Answers: []observation.DNSAnswer{}}
	current := q.Name
	seen := map[string]bool{}
	cnames := 0
	for {
		if seen[current] {
			return bad()
		}
		seen[current] = true
		next := ""
		for _, r := range records {
			if r.owner == current && r.kind == 5 {
				if next != "" && next != r.cname {
					return bad()
				}
				next = r.cname
			}
		}
		if next != "" {
			if !validName(next) {
				return bad()
			}
			for _, r := range records {
				if r.owner == current && r.kind == qtype(q.Type) {
					return bad()
				}
			}
			cnames++
			if cnames > 8 {
				return bad()
			}
			out.Answers = append(out.Answers, observation.DNSAnswer{Type: "CNAME", Value: next})
			current = next
			continue
		}
		break
	}
	// Preserve the original answer-section order after selecting the relevant chain.
	out.Answers = out.Answers[:0]
	for _, r := range records {
		if !seen[r.owner] {
			continue
		}
		var a observation.DNSAnswer
		if r.kind == 5 {
			a = observation.DNSAnswer{Type: "CNAME", Value: r.cname}
		} else {
			if !task.PublicAddress(r.address) || r.address.Is4In6() {
				continue
			}
			a = observation.DNSAnswer{Type: q.Type, Value: r.address.String()}
		}
		if !slices.Contains(out.Answers, a) {
			out.Answers = append(out.Answers, a)
		}
		if len(out.Answers) > 16 {
			return bad()
		}
	}
	return out, nil
}
