package observation

import (
	"encoding/json"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

type TLSDetails struct {
	AddressFamily         string `json:"address_family"`
	TLSVersion            string `json:"tls_version,omitempty"`
	ALPN                  string `json:"alpn,omitempty"`
	CertificateValidation string `json:"certificate_validation"`
	Alert                 string `json:"alert,omitempty"`
}
type tlsWire TLSDetails

func (TLSDetails) MarshalJSON() ([]byte, error) { return nil, ErrImport }
func (o Observation) TLS() (TLSDetails, bool) {
	if o.tls == nil {
		return TLSDetails{}, false
	}
	return *o.tls, true
}

func NewTLS(contract, status string, duration int, failure string, d TLSDetails) (Observation, error) {
	if contract != Contract {
		return Observation{}, ErrVersion
	}
	for _, s := range []string{status, failure, d.AddressFamily, d.TLSVersion, d.ALPN, d.CertificateValidation, d.Alert} {
		if len(s) > 32 {
			return Observation{}, ErrValue
		}
	}
	b, err := json.Marshal(struct {
		Type     string  `json:"type"`
		Status   string  `json:"status"`
		Duration int     `json:"duration_ms"`
		Failure  string  `json:"failure,omitempty"`
		Details  tlsWire `json:"details"`
	}{"tls_handshake", status, duration, failure, tlsWire(d)})
	if err != nil {
		return Observation{}, ErrValue
	}
	return DecodeProtocol(contract, b, ProtocolLimits{})
}

// requirement-ref: REQ-OBS-0043
// requirement-ref: REQ-OBS-0045
// requirement-ref: REQ-OBS-0050
func validateTLS(o Observation, v *jsondecode.Value, d TLSDetails) error {
	if !oneOf(d.AddressFamily, "ipv4", "ipv6") || !oneOf(d.CertificateValidation, "trusted", "untrusted", "name_mismatch", "time_invalid", "not_completed", "inconclusive") {
		return ErrValue
	}
	if _, ok := v.Lookup("tls_version"); ok && !oneOf(d.TLSVersion, "TLS1.2", "TLS1.3", "unknown") {
		return ErrValue
	}
	if _, ok := v.Lookup("alpn"); ok && !oneOf(d.ALPN, "http/1.1", "h2", "dot", "unknown") {
		return ErrValue
	}
	if _, ok := v.Lookup("alert"); ok && (o.status != "failed" || o.failure != "alert" || !oneOf(d.Alert, "handshake_failure", "certificate_error", "protocol_error", "access_denied", "internal_error", "other")) {
		return ErrValue
	}
	if o.status == "succeeded" && d.CertificateValidation != "trusted" {
		return ErrValue
	}
	if o.status == "failed" && !oneOf(o.failure, "timeout", "connection_reset", "alert", "protocol_error", "certificate_untrusted", "certificate_name_mismatch", "certificate_time_invalid", "alpn_no_overlap", "inconclusive") {
		return ErrValue
	}
	switch o.failure {
	case "certificate_untrusted":
		if d.CertificateValidation != "untrusted" {
			return ErrValue
		}
	case "certificate_name_mismatch":
		if d.CertificateValidation != "name_mismatch" {
			return ErrValue
		}
	case "certificate_time_invalid":
		if d.CertificateValidation != "time_invalid" {
			return ErrValue
		}
	}
	return nil
}

// These pure mappings consume typed facts supplied by a reviewed protocol/API
// adapter. They cannot establish negotiation or peer-alert reception themselves.
// requirement-ref: REQ-OBS-0048
func NegotiatedTLSVersion(version uint16, exposed bool) (string, bool) {
	if !exposed {
		return "", false
	}
	switch version {
	case 0x0303:
		return "TLS1.2", true
	case 0x0304:
		return "TLS1.3", true
	}
	return "unknown", true
}

// requirement-ref: REQ-OBS-0049
func NegotiatedALPN(protocol string, exposed bool) (string, bool) {
	if !exposed || protocol == "" {
		return "", false
	}
	if oneOf(protocol, "http/1.1", "h2", "dot") {
		return protocol, true
	}
	return "unknown", true
}

// ValidateALPNOffer is a pure membership/count check for future registered-test
// validators. It does not admit tasks, choose targets, or create an ALPN offer.
// requirement-ref: REQ-OBS-0049
func ValidateALPNOffer(contract string, protocols []string) error {
	if contract != Contract {
		return ErrVersion
	}
	if len(protocols) > 8 {
		return ErrValue
	}
	seen := map[string]bool{}
	for _, p := range protocols {
		if !oneOf(p, "http/1.1", "h2", "dot") || seen[p] {
			return ErrValue
		}
		seen[p] = true
	}
	return nil
}

// PeerAlertClass takes a directly received description code; exposed=false
// means omit the detail. Local errors/alerts must not be passed as peer facts.
// requirement-ref: REQ-OBS-0050
func PeerAlertClass(code uint8, exposed bool) (string, bool) {
	if !exposed {
		return "", false
	}
	switch code {
	case 40, 71, 86, 120:
		return "handshake_failure", true
	case 42, 43, 44, 45, 46, 48, 113, 116:
		return "certificate_error", true
	case 10, 20, 22, 47, 50, 51, 70, 109, 110:
		return "protocol_error", true
	case 49, 112, 115:
		return "access_denied", true
	case 80:
		return "internal_error", true
	default:
		return "other", true
	}
}
