package task

import (
	"crypto/sha256"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"
)

// Bounds of the local restriction inputs. They are implementation-local
// representation bounds, not public maxima.
const (
	MaxReleaseReviewedTasks = maxPolicyEntries
	MaxContextEndpoints     = 64
	maxEndpointBytes        = 2048
)

// ReleaseReviewedTask names one embedded task that the deployment's release
// review fixed to one run (REQ-LIMIT-0013): the embedded issuer and the
// SHA-256 digest of the exact serialized task bytes. Only a listed task may
// use the embedded one-time exception; a listing never relaxes any other
// limit or grants authority.
type ReleaseReviewedTask struct {
	trustedOnly
	IssuerID   string
	TaskSHA256 [sha256.Size]byte
}

// DeviceConstraints are local device constraints applied to every attempt in
// addition to the task's own (REQ-EXECUTION-0026). The stricter value wins:
// the higher minimum battery percent, and charging when either requires it.
// A nil MinimumBatteryPercent states no local battery minimum.
type DeviceConstraints struct {
	trustedOnly
	MinimumBatteryPercent *int
	ChargingOnly          bool
}

// Active reports whether any local device constraint is set.
func (d DeviceConstraints) Active() bool { return d.MinimumBatteryPercent != nil || d.ChargingOnly }

func (d DeviceConstraints) valid() bool {
	return d.MinimumBatteryPercent == nil || between(*d.MinimumBatteryPercent, 0, 100)
}

// ContextEligibility is local Network Context eligibility stricter than a
// task's (REQ-EXECUTION-0023, REQ-CONTEXT-FORMAT-0014). Paths names the
// network paths on which egress context may be acquired, and Endpoints the
// exact host-configured endpoint URLs acquisition may use. Neither path, or
// no endpoint, forbids acquisition. It only removes acquisition: it cannot
// add an endpoint the host did not configure and approve, or a path.
type ContextEligibility struct {
	trustedOnly
	Paths     Paths
	Endpoints []string
}

// Permits reports whether egress context may be acquired on the named
// network path ("direct" or "proxy").
func (c ContextEligibility) Permits(path string) bool {
	return path == "direct" && c.Paths.Direct || path == "proxy" && c.Paths.Proxy
}

// Narrow returns the configured endpoints, in order, that c permits.
func (c ContextEligibility) Narrow(configured ...string) []string {
	var out []string
	for _, e := range configured {
		if e != "" && slices.Contains(c.Endpoints, e) {
			out = append(out, e)
		}
	}
	return out
}

func (c ContextEligibility) valid() bool {
	if len(c.Endpoints) > MaxContextEndpoints || !unique(c.Endpoints) {
		return false
	}
	for _, e := range c.Endpoints {
		if !contextEndpoint(e) {
			return false
		}
	}
	return true
}

// contextEndpoint accepts the endpoint form the context lookup accepts: an
// absolute https URL with a host, no user information and no fragment.
func contextEndpoint(s string) bool {
	if len(s) == 0 || len(s) > maxEndpointBytes || !utf8.ValidString(s) || strings.ContainsAny(s, "\r\n\t ") {
		return false
	}
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Fragment == "" && u.Opaque == ""
}

func validReleaseReviewed(list []ReleaseReviewedTask, issuers map[string]bool) bool {
	if len(list) > MaxReleaseReviewedTasks || !unique(list) {
		return false
	}
	for _, r := range list {
		if !issuers[r.IssuerID] || r.TaskSHA256 == ([sha256.Size]byte{}) {
			return false
		}
	}
	return true
}

// ReleaseReviewed reports whether release review lists the embedded task of
// issuer with the given exact-bytes digest. An absent list lists nothing.
func (p LocalPolicy) ReleaseReviewed(issuer string, digest [sha256.Size]byte) bool {
	return p.valid && slices.Contains(p.input.ReleaseReviewed, ReleaseReviewedTask{IssuerID: issuer, TaskSHA256: digest})
}

// Device returns the local device constraints; the zero value sets none.
func (p LocalPolicy) Device() DeviceConstraints {
	d := p.input.Device
	if d.MinimumBatteryPercent != nil {
		n := *d.MinimumBatteryPercent
		d.MinimumBatteryPercent = &n
	}
	return d
}

// ContextEligibility returns the local Network Context eligibility, or false
// when local policy states none and the host configuration alone decides.
func (p LocalPolicy) ContextEligibility() (ContextEligibility, bool) {
	if p.input.Context == nil {
		return ContextEligibility{}, false
	}
	c := *p.input.Context
	c.Endpoints = slices.Clone(c.Endpoints)
	return c, true
}

func copyLocalInputs(in *PolicyInput) {
	in.ReleaseReviewed = slices.Clone(in.ReleaseReviewed)
	if in.Device.MinimumBatteryPercent != nil {
		n := *in.Device.MinimumBatteryPercent
		in.Device.MinimumBatteryPercent = &n
	}
	if in.Context != nil {
		c := *in.Context
		c.Endpoints = slices.Clone(c.Endpoints)
		in.Context = &c
	}
}
