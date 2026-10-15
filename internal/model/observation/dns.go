package observation

import (
	"encoding/json"
	"net/netip"
	"slices"
	"strings"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

type DNSAnswer struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type DNSResponse struct {
	ArrivalIndex    int         `json:"arrival_index"`
	ArrivalOffsetMS int         `json:"arrival_offset_ms"`
	RCode           string      `json:"rcode"`
	Answers         []DNSAnswer `json:"answers"`
}
type DNSDetails struct {
	Engine               string        `json:"engine"`
	AddressFamily        string        `json:"address_family,omitempty"`
	QueryType            string        `json:"query_type"`
	Responses            []DNSResponse `json:"responses"`
	ResponseLimitReached bool          `json:"response_limit_reached,omitempty"`
}

// requirement-ref: REQ-OBS-0009
// requirement-ref: REQ-OBS-0012
// requirement-ref: REQ-OBS-0013
// requirement-ref: REQ-OBS-0042
func validateDNS(o Observation, v *jsondecode.Value, d DNSDetails, l ProtocolLimits) error {
	if !oneOf(d.Engine, "system", "udp", "tcp", "doh", "dot") || !oneOf(d.QueryType, "A", "AAAA") {
		return ErrValue
	}
	if _, ok := v.Lookup("address_family"); ok && (!oneOf(d.AddressFamily, "ipv4", "ipv6", "unknown") || d.Engine == "system") {
		return ErrValue
	}
	if o.status == "failed" && !oneOf(o.failure, "nxdomain", "no_data", "servfail", "refused", "timeout", "resolver_error", "malformed_response", "unexpected_source", "query_mismatch", "inconclusive") {
		return ErrValue
	}
	maxResponses := l.MaxDNSResponses
	if maxResponses == 0 {
		maxResponses = 3
	}
	if d.Engine != "udp" {
		maxResponses = 1
	}
	if len(d.Responses) > maxResponses {
		return ErrValue
	}
	if x, ok := v.Lookup("response_limit_reached"); ok && (!x.Bool() || d.Engine != "udp" || len(d.Responses) != maxResponses) {
		return ErrValue
	}
	previous := 0
	for i, r := range d.Responses {
		if r.ArrivalIndex != i+1 || r.ArrivalOffsetMS%250 != 0 || r.ArrivalOffsetMS > o.duration || r.ArrivalOffsetMS < previous || !oneOf(r.RCode, "NOERROR", "FORMERR", "SERVFAIL", "NXDOMAIN", "NOTIMP", "REFUSED", "OTHER") {
			return ErrValue
		}
		previous = r.ArrivalOffsetMS
		cnames := 0
		seen := map[DNSAnswer]bool{}
		for _, a := range r.Answers {
			if !l.RetainDNSAnswers || seen[a] {
				return ErrValue
			}
			seen[a] = true
			switch a.Type {
			case "CNAME":
				cnames++
				if cnames > 8 || !canonicalDNSName(a.Value) {
					return ErrValue
				}
			case "A", "AAAA":
				ip, err := netip.ParseAddr(a.Value)
				if err != nil || ip.String() != a.Value || !publicAnswer(ip) || a.Type != d.QueryType || a.Type == "A" && !ip.Is4() || a.Type == "AAAA" && !ip.Is6() {
					return ErrValue
				}
			default:
				return ErrValue
			}
		}
	}
	// A multi-response UDP outcome is selected by the runner (including conflict
	// and count-limit precedence), not by a model inspecting one response alone.
	if len(d.Responses) == 1 && !d.ResponseLimitReached && d.Responses[0].RCode != "NOERROR" && o.status != "incomplete" {
		expected := dnsCodeFailure(d.Responses[0].RCode)
		if o.status != "failed" || o.failure != expected {
			return ErrValue
		}
	}
	return nil
}

func dnsCodeFailure(code string) string {
	switch code {
	case "NXDOMAIN":
		return "nxdomain"
	case "SERVFAIL":
		return "servfail"
	case "REFUSED":
		return "refused"
	}
	return "inconclusive"
}

// DNSResponseCode normalizes a complete, reliably exposed effective code, not
// just the header's low bits. The caller owns parsing, correlation and evidence.
// False means no response record can be synthesized from this API result.
// requirement-ref: REQ-OBS-0046
func DNSResponseCode(code uint16, exposed bool) (string, bool) {
	if !exposed {
		return "", false
	}
	if code < 6 {
		return [...]string{"NOERROR", "FORMERR", "SERVFAIL", "NXDOMAIN", "NOTIMP", "REFUSED"}[code], true
	}
	return "OTHER", true
}

// NewDNS consumes already normalized answers from a reviewed non-sensitive
// query. It removes exact duplicates in first-occurrence order; bounds are
// checked before publication, without truncating any distinct answer.
func NewDNS(contract, status string, duration int, failure string, d DNSDetails, l ProtocolLimits) (Observation, error) {
	if contract != Contract {
		return Observation{}, ErrVersion
	}
	if len(status) > 16 || len(failure) > 32 || len(d.Engine) > 8 || len(d.AddressFamily) > 8 || len(d.QueryType) > 4 {
		return Observation{}, ErrValue
	}
	if len(d.Responses) > 3 {
		return Observation{}, ErrValue
	}
	for _, r := range d.Responses {
		if len(r.Answers) > 16 || len(r.RCode) > 16 {
			return Observation{}, ErrValue
		}
		for _, a := range r.Answers {
			if len(a.Type) > 8 || len(a.Value) > 253 {
				return Observation{}, ErrValue
			}
		}
	}
	d = cloneDNS(d)
	for i, r := range d.Responses {
		if len(r.Answers) > 16 {
			return Observation{}, ErrValue
		}
		seen := map[DNSAnswer]bool{}
		out := make([]DNSAnswer, 0, len(r.Answers))
		for _, a := range r.Answers {
			if !seen[a] {
				seen[a] = true
				out = append(out, a)
			}
		}
		d.Responses[i].Answers = out
	}
	b, err := json.Marshal(struct {
		Type     string  `json:"type"`
		Status   string  `json:"status"`
		Duration int     `json:"duration_ms"`
		Failure  string  `json:"failure,omitempty"`
		Details  dnsWire `json:"details"`
	}{"dns_query", status, duration, failure, dnsEncoding(d)})
	if err != nil {
		return Observation{}, ErrValue
	}
	return DecodeProtocol(contract, b, l)
}
func (o Observation) DNS() (DNSDetails, bool) {
	if o.dns == nil {
		return DNSDetails{}, false
	}
	return cloneDNS(*o.dns), true
}
func cloneDNS(d DNSDetails) DNSDetails {
	d.Responses = slices.Clone(d.Responses)
	for i := range d.Responses {
		d.Responses[i].Answers = slices.Clone(d.Responses[i].Answers)
	}
	return d
}
func (DNSDetails) MarshalJSON() ([]byte, error)  { return nil, ErrImport }
func (DNSResponse) MarshalJSON() ([]byte, error) { return nil, ErrImport }
func (DNSAnswer) MarshalJSON() ([]byte, error)   { return nil, ErrImport }

type dnsAnswerWire DNSAnswer
type dnsResponseWire struct {
	ArrivalIndex    int             `json:"arrival_index"`
	ArrivalOffsetMS int             `json:"arrival_offset_ms"`
	RCode           string          `json:"rcode"`
	Answers         []dnsAnswerWire `json:"answers"`
}
type dnsWire struct {
	Engine               string            `json:"engine"`
	AddressFamily        string            `json:"address_family,omitempty"`
	QueryType            string            `json:"query_type"`
	Responses            []dnsResponseWire `json:"responses"`
	ResponseLimitReached bool              `json:"response_limit_reached,omitempty"`
}

func dnsEncoding(d DNSDetails) dnsWire {
	w := dnsWire{Engine: d.Engine, AddressFamily: d.AddressFamily, QueryType: d.QueryType, Responses: make([]dnsResponseWire, len(d.Responses)), ResponseLimitReached: d.ResponseLimitReached}
	for i, r := range d.Responses {
		w.Responses[i] = dnsResponseWire{r.ArrivalIndex, r.ArrivalOffsetMS, r.RCode, make([]dnsAnswerWire, len(r.Answers))}
		for j, a := range r.Answers {
			w.Responses[i].Answers[j] = dnsAnswerWire(a)
		}
	}
	return w
}
func canonicalDNSName(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	if _, err := netip.ParseAddr(s); err == nil {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}

// A conservative fixed retention policy, not a reachability claim or a live
// IANA query. Listed aggregates are excluded even where exceptions exist.
var excludedAnswers = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("3fff::/20"),
}

var globalIPv6Answers = netip.MustParsePrefix("2000::/3")

func publicAnswer(a netip.Addr) bool {
	if !a.IsGlobalUnicast() || a.IsPrivate() || a.Zone() != "" || a.Is4In6() {
		return false
	}
	if a.Is6() && !globalIPv6Answers.Contains(a) {
		return false
	}
	for _, p := range excludedAnswers {
		if p.Contains(a) {
			return false
		}
	}
	return true
}
