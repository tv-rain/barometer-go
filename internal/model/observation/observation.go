// Package observation owns minimized protocol and network domain values.
// It performs no network activity, execution, caching, or interpretation.
package observation

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

const Contract = "0.1.0"

var (
	ErrVersion     = errors.New("observation: unsupported contract")
	ErrValue       = errors.New("observation: invalid value")
	ErrUnavailable = errors.New("observation: unavailable model")
	ErrImport      = errors.New("observation: use explicit contract decoder")
)

// ProtocolLimits come from the validated enclosing test configuration. They
// contain no targets or check values. Zero checks means none were configured.
type ProtocolLimits struct {
	MaxBodyBytes     int
	MaxRedirects     int
	StatusCheck      bool
	HeaderChecks     int
	BodyChecks       int
	MaxDNSResponses  int
	RetainDNSAnswers bool
}

type TCPDetails struct {
	AddressFamily string `json:"address_family"`
}
type Redirect struct {
	FollowedCount         int    `json:"followed_count"`
	TerminalLocationClass string `json:"terminal_location_class"`
}
type Body struct {
	BytesProcessed int  `json:"bytes_processed"`
	Truncated      bool `json:"truncated"`
}
type Check struct {
	Kind    string `json:"kind"`
	Index   int    `json:"index"`
	Matched bool   `json:"matched"`
}
type HTTPDetails struct {
	AddressFamily   string    `json:"address_family"`
	Protocol        string    `json:"protocol"`
	StatusCode      *int      `json:"status_code,omitempty"`
	Redirect        *Redirect `json:"redirect,omitempty"`
	Body            *Body     `json:"body,omitempty"`
	Checks          []Check   `json:"checks,omitempty"`
	ContentEncoding string    `json:"content_encoding,omitempty"`
}

// Observation is immutable; zero values cannot be serialized.
type Observation struct {
	typ, status, failure string
	duration             int
	tcp                  *TCPDetails
	http                 *HTTPDetails
	dns                  *DNSDetails
	tls                  *TLSDetails
}

// DecodeProtocol selects a versionless nested schema using the enclosing
// result contract. It never guesses a version or publishes a partial value.
// requirement-ref: REQ-OBS-0003
// requirement-ref: REQ-OBS-0004
// requirement-ref: REQ-JSON-0008
// requirement-ref: REQ-OBS-0044
// requirement-ref: REQ-OBS-0051
func DecodeProtocol(contract string, input []byte, limits ProtocolLimits) (Observation, error) {
	if contract != Contract {
		return Observation{}, ErrVersion
	}
	if limits.MaxBodyBytes < 0 || limits.MaxBodyBytes > 65536 || limits.MaxRedirects < 0 || limits.MaxRedirects > 2 || limits.HeaderChecks < 0 || limits.HeaderChecks > 8 || limits.BodyChecks < 0 || limits.BodyChecks > 8 || limits.MaxDNSResponses < 0 || limits.MaxDNSResponses > 3 {
		return Observation{}, ErrValue
	}
	v, err := jsondecode.DecodeSelected(input, protocolSelection, jsondecode.Limits{MaxBytes: 65536, MaxDepth: 8, MaxNodes: 256})
	if err != nil {
		return Observation{}, err
	}
	o := Observation{typ: field(v, "type"), status: field(v, "status"), failure: field(v, "failure"), duration: number(v, "duration_ms")}
	_, hasFailure := v.Lookup("failure")
	if o.duration%250 != 0 || !oneOf(o.status, "succeeded", "failed", "incomplete") || o.status == "failed" && !hasFailure || o.status != "failed" && hasFailure {
		return Observation{}, ErrValue
	}
	details, _ := v.Lookup("details")
	switch o.typ {
	case "tcp_connect":
		d := TCPDetails{AddressFamily: field(details, "address_family")}
		if !oneOf(d.AddressFamily, "ipv4", "ipv6") || hasFailure && !oneOf(o.failure, "timeout", "connection_refused", "connection_reset", "network_unreachable", "address_unreachable", "permission_denied", "socket_error", "inconclusive") {
			return Observation{}, ErrValue
		}
		o.tcp = &d
	case "dns_query":
		if len(input) > 16384 {
			return Observation{}, jsondecode.ErrBounds
		}
		var wire struct {
			Details DNSDetails `json:"details"`
		}
		if json.Unmarshal(input, &wire) != nil {
			return Observation{}, ErrValue
		}
		if err := validateDNS(o, details, wire.Details, limits); err != nil {
			return Observation{}, err
		}
		o.dns = &wire.Details
	case "tls_handshake":
		var wire struct {
			Details TLSDetails `json:"details"`
		}
		if json.Unmarshal(input, &wire) != nil {
			return Observation{}, ErrValue
		}
		if err := validateTLS(o, details, wire.Details); err != nil {
			return Observation{}, err
		}
		o.tls = &wire.Details
	case "http_exchange":
		// Strict schema validation has already established field types/presence.
		var wire struct {
			Details HTTPDetails `json:"details"`
		}
		if json.Unmarshal(input, &wire) != nil {
			return Observation{}, ErrValue
		}
		d := wire.Details
		if err := validateHTTP(o, details, d, limits); err != nil {
			return Observation{}, err
		}
		o.http = &d
	}
	return o, nil
}

// requirement-ref: REQ-OBS-0037
// requirement-ref: REQ-OBS-0038
func validateHTTP(o Observation, v *jsondecode.Value, d HTTPDetails, l ProtocolLimits) error {
	if !oneOf(d.AddressFamily, "ipv4", "ipv6", "unknown") || !oneOf(d.Protocol, "http/1.1", "h2", "unknown") {
		return ErrValue
	}
	if _, present := v.Lookup("content_encoding"); present && !oneOf(d.ContentEncoding, "identity", "unsupported", "unknown") {
		return ErrValue
	}
	if o.status == "failed" && !oneOf(o.failure, "timeout", "protocol_error", "redirect_rejected", "redirect_limit_reached", "unexpected_status", "header_check_failed", "body_check_failed", "body_limit_reached", "unsupported_content_encoding", "unexpected_content_type", "unexpected_response", "inconclusive") {
		return ErrValue
	}
	if d.Redirect != nil && (d.Redirect.FollowedCount > l.MaxRedirects || !oneOf(d.Redirect.TerminalLocationClass, "registered_https", "unregistered_https", "non_https", "credentials", "prohibited_host", "invalid")) {
		return ErrValue
	}
	if d.Body != nil && (d.Body.BytesProcessed > l.MaxBodyBytes || d.Body.Truncated && d.Body.BytesProcessed != l.MaxBodyBytes) {
		return ErrValue
	}
	if _, present := v.Lookup("checks"); present && len(d.Checks) == 0 {
		return ErrValue
	}
	seen := map[Check]bool{}
	failedStatus, failedHeader, failedBody := false, false, false
	for _, c := range d.Checks {
		key := c
		key.Matched = false
		if seen[key] {
			return ErrValue
		}
		seen[key] = true
		switch c.Kind {
		case "status":
			if !l.StatusCheck || c.Index != 0 || d.StatusCode == nil {
				return ErrValue
			}
			failedStatus = failedStatus || !c.Matched
		case "header":
			if c.Index >= l.HeaderChecks || d.StatusCode == nil {
				return ErrValue
			}
			failedHeader = failedHeader || !c.Matched
		case "body":
			if c.Index >= l.BodyChecks || d.Body == nil {
				return ErrValue
			}
			failedBody = failedBody || !c.Matched
		default:
			return ErrValue
		}
	}
	if o.status == "succeeded" && (failedStatus || failedHeader || failedBody || d.StatusCode == nil || d.ContentEncoding == "unsupported") {
		return ErrValue
	}
	switch o.failure {
	case "redirect_rejected", "redirect_limit_reached":
		if d.Redirect == nil {
			return ErrValue
		}
	case "unexpected_status":
		if !failedStatus {
			return ErrValue
		}
	case "header_check_failed":
		if !failedHeader {
			return ErrValue
		}
	case "body_check_failed":
		if !failedBody {
			return ErrValue
		}
	case "body_limit_reached":
		if d.Body == nil || !d.Body.Truncated {
			return ErrValue
		}
	case "unsupported_content_encoding":
		if d.ContentEncoding != "unsupported" {
			return ErrValue
		}
	}
	return nil
}

func (o Observation) Type() string    { return o.typ }
func (o Observation) Status() string  { return o.status }
func (o Observation) DurationMS() int { return o.duration }
func (o Observation) Failure() string { return o.failure }
func (o Observation) TCP() (TCPDetails, bool) {
	if o.tcp == nil {
		return TCPDetails{}, false
	}
	return *o.tcp, true
}
func (o Observation) HTTP() (HTTPDetails, bool) {
	if o.http == nil {
		return HTTPDetails{}, false
	}
	d := *o.http
	if d.StatusCode != nil {
		n := *d.StatusCode
		d.StatusCode = &n
	}
	if d.Redirect != nil {
		r := *d.Redirect
		d.Redirect = &r
	}
	if d.Body != nil {
		b := *d.Body
		d.Body = &b
	}
	d.Checks = slices.Clone(d.Checks)
	return d, true
}

// MarshalJSON exposes only closed, normalized wire objects, never raw input.
// requirement-ref: REQ-OBS-0018
// requirement-ref: REQ-OBS-0031
func (o Observation) MarshalJSON() ([]byte, error) {
	var details any
	if o.tcp != nil {
		details = tcpWire(*o.tcp)
	} else if o.http != nil {
		details = httpEncoding(*o.http)
	} else if o.dns != nil {
		details = dnsEncoding(*o.dns)
	} else if o.tls != nil {
		details = tlsWire(*o.tls)
	} else {
		return nil, ErrUnavailable
	}
	return json.Marshal(struct {
		Type     string `json:"type"`
		Status   string `json:"status"`
		Duration int    `json:"duration_ms"`
		Failure  string `json:"failure,omitempty"`
		Details  any    `json:"details"`
	}{o.typ, o.status, o.duration, o.failure, details})
}
func (*Observation) UnmarshalJSON([]byte) error { return ErrImport }

// Detached views are deliberately not alternate JSON serialization paths.
// Only their validated containing model can produce a wire representation.
func (TCPDetails) MarshalJSON() ([]byte, error)  { return nil, ErrImport }
func (HTTPDetails) MarshalJSON() ([]byte, error) { return nil, ErrImport }
func (Redirect) MarshalJSON() ([]byte, error)    { return nil, ErrImport }
func (Body) MarshalJSON() ([]byte, error)        { return nil, ErrImport }
func (Check) MarshalJSON() ([]byte, error)       { return nil, ErrImport }

type tcpWire TCPDetails
type redirectWire Redirect
type bodyWire Body
type checkWire Check
type httpWire struct {
	AddressFamily   string        `json:"address_family"`
	Protocol        string        `json:"protocol"`
	StatusCode      *int          `json:"status_code,omitempty"`
	Redirect        *redirectWire `json:"redirect,omitempty"`
	Body            *bodyWire     `json:"body,omitempty"`
	Checks          []checkWire   `json:"checks,omitempty"`
	ContentEncoding string        `json:"content_encoding,omitempty"`
}

func httpEncoding(d HTTPDetails) httpWire {
	w := httpWire{AddressFamily: d.AddressFamily, Protocol: d.Protocol, StatusCode: d.StatusCode, Redirect: (*redirectWire)(d.Redirect), Body: (*bodyWire)(d.Body), ContentEncoding: d.ContentEncoding}
	for _, c := range d.Checks {
		w.Checks = append(w.Checks, checkWire(c))
	}
	return w
}

// NewHTTP validates a minimized producer value using the enclosing test's bounds.
func NewHTTP(contract, status string, duration int, failure string, d HTTPDetails, limits ProtocolLimits) (Observation, error) {
	if contract != Contract {
		return Observation{}, ErrVersion
	}
	if len(status) > 16 || len(failure) > 32 || len(d.AddressFamily) > 16 || len(d.Protocol) > 16 || len(d.ContentEncoding) > 16 || len(d.Checks) > 17 || d.Redirect != nil && len(d.Redirect.TerminalLocationClass) > 32 {
		return Observation{}, ErrValue
	}
	b, err := json.Marshal(struct {
		Type     string   `json:"type"`
		Status   string   `json:"status"`
		Duration int      `json:"duration_ms"`
		Failure  string   `json:"failure,omitempty"`
		Details  httpWire `json:"details"`
	}{"http_exchange", status, duration, failure, httpEncoding(d)})
	if err != nil {
		return Observation{}, ErrValue
	}
	return DecodeProtocol(contract, b, limits)
}

func oneOf(s string, values ...string) bool { return slices.Contains(values, s) }
func field(v *jsondecode.Value, name string) string {
	x, ok := v.Lookup(name)
	if !ok {
		return ""
	}
	return x.Text()
}
