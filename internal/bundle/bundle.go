// Package bundle owns bounded, immutable Result Bundle production. It performs
// no transport, persistence, task selection, key derivation or platform calls.
package bundle

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/result"
)

const (
	Contract   = "0.1.0"
	MaxResults = 16
	MaxBytes   = 262144
)

var (
	ErrBinding     = errors.New("bundle: invalid local binding")
	ErrValue       = errors.New("bundle: invalid value")
	ErrUnavailable = errors.New("bundle: unavailable bundle")
)

// Builder pins trusted initialization, independently of task and result data.
// Capability availability gates the representation; it does not establish that
// derivation succeeded. A future private-state owner must supply only a valid
// HMAC output and non-decreasing epoch, or nil when linkage is unavailable.
type Builder struct {
	ready, linkage bool
	scope          string
}

// New selects only the exact bundle contract. It copies local policy without
// invoking the clock or protected-key capability or retaining either handle.
// requirement-ref: REQ-BUNDLE-0004
// requirement-ref: REQ-LINKAGE-0006
func New(contract string, initialization *config.Initialization) (*Builder, error) {
	if contract != Contract {
		return nil, jsondecode.ErrVersion
	}
	if initialization == nil || !initialization.Configuration().Valid() {
		return nil, ErrBinding
	}
	_, available := initialization.LinkageKeys()
	l := initialization.Configuration().Linkage()
	return &Builder{ready: true, linkage: available && l.Mode == "rotating_installation", scope: l.Scope}, nil
}

// Value owns the exact retry bytes and logical identity. Copies remain the same
// logical bundle; Build always creates a fresh identity, even for equal inputs.
// There is no mutation, caller-selected ID, or JSON restoration path. A future
// persistence owner needs a separately reviewed exact-byte restoration boundary.
type Value struct {
	id      string
	linkage *model.RotatingLinkage
	results []result.Value
	encoded []byte
}

func (v Value) MarshalJSON() ([]byte, error) {
	if len(v.encoded) == 0 {
		return nil, ErrUnavailable
	}
	return slices.Clone(v.encoded), nil
}
func (*Value) UnmarshalJSON([]byte) error { return model.ErrEncoding }
func (v Value) ID() string                { return v.id }

// View returns detached records. Editing them cannot change delivery bytes.
func (v Value) View() model.Bundle {
	if len(v.encoded) == 0 {
		return model.Bundle{}
	}
	b := model.Bundle{SchemaVersion: Contract, BundleID: v.id, Linkage: cloneLinkage(v.linkage), Results: make([]model.Result, len(v.results))}
	for i, r := range v.results {
		b.Results[i] = r.View()
	}
	return b
}

// Build accepts only immutable results already validated by their exact Output
// owner. Linkage is a trusted internal projection, never task/collector input.
// nil means absent even when local configuration permits rotating linkage.
// No invalid result is dropped and no oversized bundle is truncated or split.
// requirement-ref: REQ-BUNDLE-0001
// requirement-ref: REQ-BUNDLE-0002
// requirement-ref: REQ-BUNDLE-0003
// requirement-ref: REQ-BUNDLE-0006
// requirement-ref: REQ-BUNDLE-0007
// requirement-ref: REQ-BUNDLE-0010
// requirement-ref: REQ-BUNDLE-0011
// requirement-ref: REQ-BUNDLE-0016
// requirement-ref: REQ-BUNDLE-0017
// requirement-ref: REQ-BUNDLE-0018
// requirement-ref: REQ-BUNDLE-0021
// requirement-ref: REQ-BUNDLE-0022
// requirement-ref: REQ-BUNDLE-0023
// requirement-ref: REQ-BUNDLE-0025
// requirement-ref: REQ-BUNDLE-0026
// requirement-ref: REQ-LIMIT-0021
func (b *Builder) Build(results []result.Value, linkage *model.RotatingLinkage) (Value, error) {
	if b == nil || !b.ready {
		return Value{}, ErrBinding
	}
	if len(results) < 1 || len(results) > MaxResults {
		return Value{}, jsondecode.ErrBounds
	}
	l := cloneLinkage(linkage)
	var link []byte
	if l != nil {
		if !b.linkage || l.Scope != b.scope || l.Epoch < 0 {
			return Value{}, ErrValue
		}
		// Scope has already passed the exact reviewed initialization allowlist.
		link, _ = json.Marshal(struct {
			Mode  string `json:"mode"`
			Scope string `json:"scope"`
			Epoch int64  `json:"epoch"`
			Token string `json:"token"`
		}{"rotating_installation", l.Scope, l.Epoch, base64.RawURLEncoding.EncodeToString(l.Token[:])})
	}
	// Compute the complete size before allocating the output or generating an ID.
	// A 16-byte unpadded base64url ID is always 22 bytes.
	size := len(`{"schema_version":"0.1.0","bundle_id":"","results":[]}`) + 22
	if l != nil {
		size += len(`,"linkage":`) + len(link)
	}
	items := make([][]byte, len(results))
	for i, r := range results {
		data, err := r.MarshalJSON()
		if err != nil {
			return Value{}, ErrValue
		}
		items[i] = data
		size += len(data)
		if i > 0 {
			size++
		}
		if size > MaxBytes {
			return Value{}, jsondecode.ErrBounds
		}
	}
	id := newID()
	encoded := make([]byte, 0, size)
	encoded = append(encoded, `{"schema_version":"0.1.0","bundle_id":"`...)
	encoded = append(encoded, id...)
	encoded = append(encoded, '"')
	if l != nil {
		encoded = append(encoded, `,"linkage":`...)
		encoded = append(encoded, link...)
	}
	encoded = append(encoded, `,"results":[`...)
	for i, item := range items {
		if i > 0 {
			encoded = append(encoded, ',')
		}
		encoded = append(encoded, item...)
	}
	encoded = append(encoded, ']', '}')
	return Value{id, l, slices.Clone(results), encoded}, nil
}

func newID() string {
	for {
		var id [16]byte
		_, _ = rand.Read(id[:])
		if id != ([16]byte{}) {
			return base64.RawURLEncoding.EncodeToString(id[:])
		}
	}
}
func cloneLinkage(l *model.RotatingLinkage) *model.RotatingLinkage {
	if l == nil {
		return nil
	}
	c := *l
	return &c
}
