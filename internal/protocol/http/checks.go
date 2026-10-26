package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime"
	stdhttp "net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tv-rain/barometer-go/internal/model/observation"
)

type HeaderCheck struct{ Name, Type, Value string }
type BodyCheck struct{ Type, Value string }

func validate(c Config) (*url.URL, error) {
	if len(c.URL) > 2048 || !utf8.ValidString(c.URL) || c.MaxBodyBytes < 1 || c.MaxBodyBytes > 65536 || c.MaxRedirects < 0 || c.MaxRedirects > 2 || len(c.HeaderChecks) > 8 || len(c.BodyChecks) > 8 || c.ExpectedStatus != 0 && (c.ExpectedStatus < 100 || c.ExpectedStatus > 599) || len(c.PostBody) > 65535 || len(c.PostBody) > 0 && c.MaxRedirects != 0 {
		return nil, ErrInterrupted
	}
	u, e := url.Parse(c.URL)
	if e != nil || u.Scheme != "https" && u.Scheme != "http" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" || strings.ContainsAny(u.Host, "\r\n\t ") {
		return nil, ErrInterrupted
	}
	if len(u.String()) > 2048 {
		return nil, ErrInterrupted
	}
	if u.Port() != "" {
		p, e := strconv.Atoi(u.Port())
		if e != nil || p < 1 || p > 65535 {
			return nil, ErrInterrupted
		}
	}
	if u.Scheme == "http" && (len(c.PostBody) > 0 || c.MaxRedirects > 0) {
		return nil, ErrInterrupted
	}
	for _, h := range c.HeaderChecks {
		if !headerName(h.Name) || sensitiveHeader(h.Name) {
			return nil, ErrInterrupted
		}
		switch h.Type {
		case "present", "absent":
			if h.Value != "" {
				return nil, ErrInterrupted
			}
		case "equals", "contains":
			if !literal(h.Value) {
				return nil, ErrInterrupted
			}
		default:
			return nil, ErrInterrupted
		}
	}
	for _, b := range c.BodyChecks {
		switch b.Type {
		case "contains", "prefix", "suffix":
			if !literal(b.Value) {
				return nil, ErrInterrupted
			}
		case "sha256":
			decoded, e := hex.DecodeString(b.Value)
			if e != nil || len(decoded) != 32 || strings.ToLower(b.Value) != b.Value {
				return nil, ErrInterrupted
			}
		default:
			return nil, ErrInterrupted
		}
	}
	return u, nil
}
func literal(s string) bool { return len(s) > 0 && len(s) <= 256 && utf8.ValidString(s) }
func headerName(s string) bool {
	if len(s) == 0 || len(s) > 256 {
		return false
	}
	for _, c := range []byte(s) {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c))) {
			return false
		}
	}
	return true
}
func sensitiveHeader(s string) bool {
	switch strings.ToLower(s) {
	case "authorization", "proxy-authorization", "www-authenticate", "proxy-authenticate", "cookie", "set-cookie", "authentication-info", "proxy-authentication-info":
		return true
	}
	return false
}

// evaluate retains only decidable indexed check outcomes. No decompression,
// body drain or extra byte probe occurs at the application body bound.
// requirement-ref: REQ-OBS-0028
// requirement-ref: REQ-OBS-0029
// requirement-ref: REQ-OBS-0030
// requirement-ref: REQ-OBS-0031
// requirement-ref: REQ-HTTP-01-0037
func evaluate(r *stdhttp.Response, d *observation.HTTPDetails, c Config) ([]byte, string) {
	failure := ""
	if c.ExpectedStatus != 0 {
		matched := r.StatusCode == c.ExpectedStatus
		d.Checks = append(d.Checks, observation.Check{Kind: "status", Matched: matched})
		if !matched {
			failure = "unexpected_status"
		}
	}
	for i, h := range c.HeaderChecks {
		values := r.Header.Values(h.Name)
		matched := false
		switch h.Type {
		case "present":
			matched = len(values) > 0
		case "absent":
			matched = len(values) == 0
		case "equals", "contains":
			for _, v := range values {
				if h.Type == "equals" && v == h.Value || h.Type == "contains" && strings.Contains(v, h.Value) {
					matched = true
				}
			}
		}
		d.Checks = append(d.Checks, observation.Check{Kind: "header", Index: i, Matched: matched})
		if !matched && failure == "" {
			failure = "header_check_failed"
		}
	}
	encoding := strings.TrimSpace(strings.Join(r.Header.Values("Content-Encoding"), ","))
	if encoding != "" && !strings.EqualFold(encoding, "identity") {
		if len(c.BodyChecks) > 0 || len(c.PostBody) > 0 {
			d.ContentEncoding = "unsupported"
			return nil, "unsupported_content_encoding"
		}
	} else {
		d.ContentEncoding = "identity"
	}
	if len(c.PostBody) > 0 {
		media, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if e != nil || media != "application/dns-message" {
			return nil, "unexpected_content_type"
		}
	}
	body := make([]byte, c.MaxBodyBytes)
	n, complete := 0, false
	d.Body = &observation.Body{}
	for n < len(body) {
		k, e := r.Body.Read(body[n:])
		n += k
		d.Body.BytesProcessed = n
		if e == io.EOF {
			complete = true
			break
		}
		if e != nil {
			return nil, ioFailure(e)
		}
		if k == 0 {
			return nil, "inconclusive"
		}
	}
	body = body[:n]
	if r.ContentLength >= 0 {
		if complete && int64(n) != r.ContentLength {
			return nil, "protocol_error"
		}
		complete = int64(n) == r.ContentLength
	}
	d.Body.Truncated = !complete && n == c.MaxBodyBytes
	for i, b := range c.BodyChecks {
		matched, decided := bodyCheck(body, complete, b)
		if !decided {
			continue
		}
		d.Checks = append(d.Checks, observation.Check{Kind: "body", Index: i, Matched: matched})
		if !matched && failure == "" {
			failure = "body_check_failed"
		}
	}
	if d.Body.Truncated {
		decided := 0
		for _, v := range d.Checks {
			if v.Kind == "body" {
				decided++
			}
		}
		if decided < len(c.BodyChecks) && failure == "" {
			failure = "body_limit_reached"
		}
	}
	return body, failure
}
func bodyCheck(body []byte, complete bool, c BodyCheck) (bool, bool) {
	v := []byte(c.Value)
	switch c.Type {
	case "contains":
		if bytes.Contains(body, v) {
			return true, true
		}
		return false, complete
	case "prefix":
		if len(body) < len(v) && !bytes.HasPrefix(v, body) {
			return false, true
		}
		if len(body) >= len(v) {
			return bytes.HasPrefix(body, v), true
		}
		return false, complete
	case "suffix":
		return bytes.HasSuffix(body, v), complete
	case "sha256":
		sum := sha256.Sum256(body)
		return hex.EncodeToString(sum[:]) == c.Value, complete
	}
	return false, false
}
