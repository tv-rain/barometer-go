package result

import (
	"encoding/json"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"slices"
)

// requirement-ref: REQ-RESULT-0021
// requirement-ref: REQ-RESULT-0026
// requirement-ref: REQ-RESULT-0027
// requirement-ref: REQ-RESULT-0029
// requirement-ref: REQ-RESULT-0041
// requirement-ref: REQ-RESULT-0043
func (d *Decoder) measurement(input []byte, o Output, state string, duration int) (*model.Measurement, error) {
	root, err := jsondecode.Decode(input, measurementSchema(o.Test.ID), jsondecode.Limits{MaxBytes: 65536, MaxDepth: 14, MaxNodes: 32768})
	if err != nil {
		return nil, err
	}
	var w measurementWire
	if json.Unmarshal(input, &w) != nil {
		return nil, ErrValue
	}
	obs, err := decodeObservations(w.Observations, o, duration)
	if err != nil {
		return nil, err
	}
	m := &model.Measurement{Outcome: w.Outcome, Observations: obs}
	switch o.Test.ID {
	case "NET-01":
		if len(obs) != 0 || !oneOf(m.Outcome, "snapshot", "platform_unavailable", "platform_error", "inconclusive") {
			return nil, ErrValue
		}
		if len(w.Access) != 0 {
			if !oneOf(m.Outcome, "snapshot", "inconclusive") {
				return nil, ErrValue
			}
			a, e := observation.DecodeAccess(Contract, w.Access, d.binding.AccessGeneration)
			if e != nil {
				return nil, e
			}
			m.Access = &a
			if m.Outcome == "snapshot" && a.Value().NetworkChanged {
				return nil, ErrValue
			}
		}
		if m.Outcome == "snapshot" && m.Access == nil {
			return nil, ErrValue
		}
	case "NET-05":
		if _, present := root.Lookup("stopped_before_step"); present && w.StoppedBeforeStep == "" {
			return nil, ErrValue
		}
		if err = d.composite(w, m, state, duration); err != nil {
			return nil, err
		}
	default:
		if len(obs) == 0 {
			return nil, ErrValue
		}
		for i, v := range obs {
			if v.Status() == "incomplete" || i < len(obs)-1 && v.Status() != "succeeded" {
				return nil, ErrValue
			}
		}
		if !validPrimitive(o, m) {
			return nil, ErrValue
		}
		if o.Test.ID == "NET-03" {
			m.PlatformValidation = w.PlatformValidation
			m.ResponseClass = w.ResponseClass
			_, hasResponse := root.Lookup("response_class")
			if !net03(m, hasResponse, o.Limits) {
				return nil, ErrValue
			}
		}
	}
	return m, nil
}

// requirement-ref: REQ-NET-05-0041
// requirement-ref: REQ-NET-05-0044
// requirement-ref: REQ-NET-05-0045
// requirement-ref: REQ-NET-05-0046
func (d *Decoder) composite(w measurementWire, m *model.Measurement, state string, duration int) error {
	p := d.binding.Profile
	if len(m.Observations) != 0 || w.Profile.ID != p.ID || w.Profile.Version != p.Version || !oneOf(w.NetworkValidation.State, "validated", "not_validated", "captive_portal_suspected", "unknown") || !oneOf(w.NetworkValidation.Source, "fresh", "cached", "platform") {
		return ErrValue
	}
	m.Profile = &model.Profile{ID: w.Profile.ID, Version: w.Profile.Version}
	m.NetworkValidation = &model.NetworkValidation{State: w.NetworkValidation.State, Source: w.NetworkValidation.Source}
	steps := *w.Steps
	if len(steps) > len(p.Steps) {
		return ErrValue
	}
	m.Steps = make([]model.Step, 0, len(steps))
	incomplete := false
	for i, s := range steps {
		expected := p.Steps[i]
		if s.ID != expected.ID || s.Test != testWireOf(expected.Output.Test) {
			return jsondecode.ErrVersion
		}
		v := model.Step{ID: s.ID, Test: expected.Output.Test, State: s.State}
		switch s.State {
		case "completed":
			if len(s.Measurement) == 0 {
				return ErrValue
			}
			child, err := d.measurement(s.Measurement, expected.Output, "completed", duration)
			if err != nil {
				return err
			}
			v.Measurement = child
		case "incomplete":
			if i != len(steps)-1 || len(s.Measurement) != 0 || state != "incomplete" {
				return ErrValue
			}
			incomplete = true
		default:
			return ErrValue
		}
		m.Steps = append(m.Steps, v)
	}
	if len(steps) < len(p.Steps) {
		if w.StoppedBeforeStep != p.Steps[len(steps)].ID {
			return ErrValue
		}
	} else if w.StoppedBeforeStep != "" {
		return ErrValue
	}
	m.StoppedBeforeStep = w.StoppedBeforeStep
	switch m.Outcome {
	case "completed":
		if state != "completed" || incomplete || len(steps) != len(p.Steps) {
			return ErrValue
		}
	case "incomplete":
		if state != "incomplete" || !incomplete && len(steps) == len(p.Steps) {
			return ErrValue
		}
	case "network_not_validated":
		if state != "completed" || len(steps) != 0 || !p.RequireValidation || !oneOf(w.NetworkValidation.State, "not_validated", "captive_portal_suspected") {
			return ErrValue
		}
	case "inconclusive":
		if state != "completed" || incomplete {
			return ErrValue
		}
	default:
		return ErrValue
	}
	if p.RequireValidation && w.NetworkValidation.State != "validated" && len(steps) != 0 {
		return ErrValue
	}
	return nil
}

// Exact initial contract dispatch, without runtime runners or a global registry.
func validPrimitive(o Output, m *model.Measurement) bool {
	obs := m.Observations
	last := obs[len(obs)-1]
	outcome := m.Outcome
	switch o.Test.ID {
	case "TCP-01":
		return len(obs) == 1 && last.Type() == "tcp_connect" && terminal(last, outcome, "connected")
	case "TLS-01":
		if !sequence(obs, []string{"tcp_connect", "tls_handshake"}) {
			return false
		}
		return tlsOutcome(last, outcome)
	case "DNS-01":
		if len(obs) != 1 || last.Type() != "dns_query" {
			return false
		}
		dns, _ := last.DNS()
		return dns.Engine == o.DNSMode && dns.QueryType == o.QueryType && dnsOutcome(last, outcome)
	case "DNS-03", "DNS-04":
		types := []string{"tcp_connect", "tls_handshake"}
		engine := "dot"
		if o.Test.ID == "DNS-03" {
			types = append(types, "http_exchange")
			engine = "doh"
		}
		types = append(types, "dns_query")
		if o.Bootstrap {
			types = append([]string{"dns_query"}, types...)
		}
		validSequence := sequence(obs, types)
		// A completed HTTP exchange with a non-success status is itself a
		// terminal DoH outcome, even though the DNS exchange never began.
		if o.Test.ID == "DNS-03" && last.Type() == "http_exchange" && outcome == "http_failure" {
			h, _ := last.HTTP()
			if h.StatusCode != nil && (*h.StatusCode < 200 || *h.StatusCode >= 300) {
				// Check the actual prefix without synthesizing a DNS observation.
				validSequence = len(obs) == len(types)-1
				for i, v := range obs {
					validSequence = validSequence && v.Type() == types[i]
				}
			}
		}
		if !validSequence {
			return false
		}
		for i, v := range obs {
			if v.Type() == "dns_query" {
				d, _ := v.DNS()
				expected := engine
				if o.Bootstrap && i == 0 {
					expected = "system"
				}
				if d.Engine != expected || i == len(types)-1 && d.QueryType != o.QueryType {
					return false
				}
			}
		}
		return encryptedDNSOutcome(last, outcome, o.Test.ID, o.Bootstrap && len(obs) == 1)
	case "HTTP-01":
		exchanges := 0
		for _, v := range obs {
			if v.Type() == "http_exchange" {
				exchanges++
			}
		}
		if exchanges > o.Limits.MaxRedirects+1 {
			return false
		}
		if !httpSequence(obs, true, true) {
			return false
		}
		return httpOutcome(last, outcome, o.Limits)
	case "NET-03":
		return httpSequence(obs, o.HTTPS, false) && oneOf(outcome, "validated", "captive_portal_suspected", "not_validated", "dns_failure", "connection_failure", "tls_failure", "unexpected_response", "inconclusive")
	}
	return false
}
func terminal(o observation.Observation, outcome, positive string) bool {
	return outcome == positive && o.Status() == "succeeded" || o.Status() == "failed" && outcome == o.Failure()
}
func sequence(obs []observation.Observation, types []string) bool {
	if len(obs) > len(types) {
		return false
	}
	for i, v := range obs {
		if v.Type() != types[i] {
			return false
		}
	}
	// A successful prefix is not a terminal test result.
	return len(obs) == len(types) || obs[len(obs)-1].Status() == "failed"
}
func tlsOutcome(o observation.Observation, outcome string) bool {
	if o.Type() == "tcp_connect" {
		return failed(o, outcome, map[string]string{"tcp_timeout": "timeout", "tcp_refused": "connection_refused", "tcp_reset": "connection_reset", "inconclusive": "inconclusive"})
	}
	return oneOf(outcome, "handshake_succeeded", "protocol_error", "certificate_untrusted", "certificate_name_mismatch", "certificate_time_invalid", "alpn_no_overlap", "inconclusive") && terminal(o, outcome, "handshake_succeeded") || failed(o, outcome, map[string]string{"tls_timeout": "timeout", "tls_alert": "alert", "tcp_reset": "connection_reset"})
}
func failed(o observation.Observation, outcome string, mapping map[string]string) bool {
	failure, ok := mapping[outcome]
	return ok && o.Status() == "failed" && o.Failure() == failure
}
func encryptedDNSOutcome(o observation.Observation, outcome, id string, bootstrap bool) bool {
	if o.Status() == "failed" && o.Failure() == "inconclusive" {
		return outcome == "inconclusive"
	}
	if bootstrap {
		return o.Status() == "failed" && outcome == "bootstrap_failure" || id == "DNS-03" && o.Failure() == "timeout" && outcome == "timeout"
	}
	switch o.Type() {
	case "tcp_connect":
		if id == "DNS-03" {
			return o.Status() == "failed" && outcome == "tcp_failure" || o.Failure() == "timeout" && outcome == "timeout"
		}
		return failed(o, outcome, map[string]string{"tcp_timeout": "timeout", "tcp_refused": "connection_refused", "tcp_reset": "connection_reset"})
	case "tls_handshake":
		return o.Status() == "failed" && outcome == "tls_failure" || o.Failure() == "timeout" && outcome == "timeout" || id == "DNS-04" && failed(o, outcome, map[string]string{"tcp_reset": "connection_reset"})
	case "http_exchange":
		h, _ := o.HTTP()
		nonSuccess := h.StatusCode != nil && (*h.StatusCode < 200 || *h.StatusCode >= 300)
		return (o.Status() == "failed" || nonSuccess) && outcome == "http_failure" || failed(o, outcome, map[string]string{"unexpected_content_type": "unexpected_content_type", "timeout": "timeout"})
	case "dns_query":
		if o.Status() == "succeeded" {
			d, _ := o.DNS()
			return outcome == "response" && len(d.Responses) == 1 && d.Responses[0].RCode == "NOERROR"
		}
		return failed(o, outcome, map[string]string{"dns_nxdomain": "nxdomain", "dns_servfail": "servfail", "dns_refused": "refused", "malformed_dns_response": "malformed_response", "query_mismatch": "query_mismatch", "timeout": "timeout"})
	}
	return false
}

func dnsOutcome(o observation.Observation, outcome string) bool {
	d, _ := o.DNS()
	if d.ResponseLimitReached {
		return outcome == "response_limit_reached"
	}
	conflict := false
	for i, r := range d.Responses {
		if i == 0 {
			continue
		}
		first := d.Responses[0]
		if r.RCode != first.RCode || !sameAnswers(first.Answers, r.Answers) {
			conflict = true
		}
	}
	if conflict {
		return outcome == "conflicting_responses"
	}
	if oneOf(outcome, "conflicting_responses", "response_limit_reached") {
		return false
	}
	if outcome == "resolved" {
		if o.Status() != "succeeded" {
			return false
		}
		// When answer retention was not approved, successful system/API resolution
		// can be represented without retained address values.
		return true
	}
	if outcome == "no_data" && o.Status() == "succeeded" {
		for _, r := range d.Responses {
			for _, a := range r.Answers {
				if a.Type == d.QueryType {
					return false
				}
			}
		}
		return true
	}
	return o.Status() == "failed" && outcome == o.Failure()
}
func sameAnswers(a, b []observation.DNSAnswer) bool {
	if len(a) != len(b) {
		return false
	}
	for _, value := range a {
		if !slices.Contains(b, value) {
			return false
		}
	}
	return true
}

// Redirect/reuse-aware operation ordering. A failure ends the sequence. HTTP-01
// permits additional bounded exchanges; NET-03 never follows redirects.
func httpSequence(obs []observation.Observation, https, redirects bool) bool {
	stage := 0
	exchanges := 0
	for _, o := range obs {
		switch o.Type() {
		case "dns_query":
			d, _ := o.DNS()
			if d.Engine != "system" || stage != 0 {
				return false
			}
			stage = 1
		case "tcp_connect":
			if stage > 1 {
				return false
			}
			stage = 2
		case "tls_handshake":
			if !https || stage != 2 {
				return false
			}
			stage = 3
		case "http_exchange":
			if https && stage != 3 && !(redirects && exchanges > 0 && stage == 0) || !https && stage != 2 {
				return false
			}
			exchanges++
			stage = 0
		default:
			return false
		}
		if exchanges > 1 && !redirects {
			return false
		}
	}
	last := obs[len(obs)-1]
	return last.Status() == "failed" || last.Type() == "http_exchange"
}
func httpOutcome(o observation.Observation, outcome string, l observation.ProtocolLimits) bool {
	switch o.Type() {
	case "dns_query":
		return o.Status() == "failed" && outcome == "dns_failure" || failed(o, outcome, map[string]string{"inconclusive": "inconclusive"})
	case "tcp_connect":
		return o.Status() == "failed" && outcome == "tcp_failure" || failed(o, outcome, map[string]string{"inconclusive": "inconclusive"})
	case "tls_handshake":
		return o.Status() == "failed" && outcome == "tls_failure" || failed(o, outcome, map[string]string{"inconclusive": "inconclusive"})
	}
	if o.Type() != "http_exchange" {
		return false
	}
	h, _ := o.HTTP()
	if outcome == "expected_response" {
		if o.Status() != "succeeded" {
			return false
		}
		count := l.HeaderChecks + l.BodyChecks
		if l.StatusCheck {
			count++
		}
		if len(h.Checks) != count {
			return false
		}
		for _, c := range h.Checks {
			if !c.Matched {
				return false
			}
		}
		return true
	}
	if outcome == "redirect_observed" {
		return o.Status() == "succeeded" && h.Redirect != nil && h.Redirect.FollowedCount == 0 && h.StatusCode != nil && oneOfStatus(*h.StatusCode, 301, 302, 303, 307, 308)
	}
	if outcome == "http_timeout" {
		return o.Status() == "failed" && o.Failure() == "timeout"
	}
	return o.Status() == "failed" && outcome == o.Failure() && oneOf(outcome, "protocol_error", "redirect_rejected", "redirect_limit_reached", "unexpected_status", "header_check_failed", "body_check_failed", "body_limit_reached", "unsupported_content_encoding", "inconclusive")
}
func oneOfStatus(n int, values ...int) bool { return slices.Contains(values, n) }
func net03(m *model.Measurement, hasResponse bool, limits observation.ProtocolLimits) bool {
	if !oneOf(m.PlatformValidation, "validated", "not_validated", "captive_portal", "unknown") {
		return false
	}
	last := m.Observations[len(m.Observations)-1]
	h, isHTTP := last.HTTP()
	response := isHTTP && h.StatusCode != nil
	if hasResponse != response || hasResponse && !oneOf(m.ResponseClass, "expected", "redirect", "content_mismatch", "unexpected") {
		return false
	}
	if m.ResponseClass == "expected" && !httpOutcome(last, "expected_response", limits) {
		return false
	}
	if m.ResponseClass == "redirect" && (h.Redirect == nil || h.Redirect.FollowedCount != 0 || h.StatusCode == nil || !oneOfStatus(*h.StatusCode, 301, 302, 303, 307, 308)) {
		return false
	}
	switch m.Outcome {
	case "validated":
		return m.ResponseClass == "expected" && last.Status() == "succeeded"
	case "captive_portal_suspected":
		return m.PlatformValidation == "captive_portal" || oneOf(m.ResponseClass, "redirect", "content_mismatch")
	case "not_validated":
		return m.PlatformValidation == "not_validated" && !oneOf(m.ResponseClass, "expected", "redirect", "content_mismatch")
	case "unexpected_response":
		return m.ResponseClass == "unexpected"
	case "dns_failure":
		return last.Type() == "dns_query" && last.Status() == "failed"
	case "connection_failure":
		return last.Type() == "tcp_connect" && last.Status() == "failed"
	case "tls_failure":
		return last.Type() == "tls_handshake" && last.Status() == "failed"
	case "inconclusive":
		return true
	}
	return false
}
