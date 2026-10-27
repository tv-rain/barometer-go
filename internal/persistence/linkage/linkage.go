// Package linkage derives the optional rotating result-linkage projection
// from the protected-key capability and owns the in-memory linkage epoch
// state. The installation secret never reaches it: derivation runs through
// platform.LinkageKeyStore. It stores no token between derivations, persists
// nothing, sends nothing and has no fallback key storage, so every failure
// omits linkage. Linkage is pseudonymous, not anonymous.
package linkage

import (
	"context"
	"strconv"
	"sync"

	"github.com/tv-rain/barometer-go/internal/config"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/platform"
)

// purpose is the fixed domain-separation prefix of REQ-LINKAGE-0010.
const purpose = "barometer-result-linkage-v1"

// Derive returns the linkage of a bundle created now, or nil when linkage is
// disabled or unavailable. It is bounded by ctx.
type Derive func(context.Context) *model.RotatingLinkage

// Source holds the epoch floor of one client: the highest epoch it has used
// for the current scope and rotation period. The floor is the only linkage
// state kept between derivations. It lives for the client, so a later
// generation with the same configuration never reuses an earlier epoch;
// another scope or period, or Erase, starts it again. Source is safe for
// concurrent use and holds no lock across a capability call.
type Source struct {
	mu     sync.Mutex
	scope  string
	period int64
	floor  int64  // -1 before any epoch is used
	erased uint64 // counts Erase calls
}

func NewSource() *Source { return &Source{floor: -1} }

func (*Source) String() string   { return "linkage.Source" }
func (*Source) GoString() string { return "linkage.Source" }

// For returns the derivation of one prepared initialization, or nil when it
// does not enable rotating linkage or has no protected-key capability. Only
// trusted initialization enables linkage; no task or result value reaches it.
// requirement-ref: REQ-LINKAGE-0004
// requirement-ref: REQ-LINKAGE-0006
// requirement-ref: REQ-LINKAGE-0043
// requirement-ref: REQ-LINKAGE-0044
func (s *Source) For(init *config.Initialization) Derive {
	if s == nil || init == nil {
		return nil
	}
	keys, ok := init.LinkageKeys()
	l := init.Configuration().Linkage()
	if !ok || l.Mode != "rotating_installation" || l.RotationPeriodSeconds <= 0 {
		return nil
	}
	clock, scope, period := init.Clock(), l.Scope, int64(l.RotationPeriodSeconds)
	return func(ctx context.Context) *model.RotatingLinkage {
		return s.derive(ctx, clock, keys, scope, period)
	}
}

// derive computes E = floor(unix / period) from the trusted clock and
// HMAC-SHA-256(secret, purpose || 0x00 || scope || 0x00 || decimal(E))
// inside the key store. An unavailable or negative time, an epoch below the
// floor, a capability failure and an Erase during derivation each omit
// linkage. The message carries only the purpose, the reviewed deployment
// scope and the epoch.
// requirement-ref: REQ-LINKAGE-0010
// requirement-ref: REQ-LINKAGE-0011
// requirement-ref: REQ-LINKAGE-0012
// requirement-ref: REQ-LINKAGE-0014
// requirement-ref: REQ-LINKAGE-0015
// requirement-ref: REQ-LINKAGE-0037
func (s *Source) derive(ctx context.Context, clock platform.Clock, keys platform.LinkageKeyStore, scope string, period int64) *model.RotatingLinkage {
	if ctx.Err() != nil {
		return nil
	}
	now, err := clock.Now(ctx)
	if err != nil || ctx.Err() != nil || now.Unix() < 0 {
		return nil
	}
	epoch := now.Unix() / period
	s.mu.Lock()
	if s.scope != scope || s.period != period {
		s.scope, s.period, s.floor = scope, period, -1
	}
	erased, usable := s.erased, epoch >= s.floor
	s.mu.Unlock()
	if !usable {
		return nil
	}
	message := make([]byte, 0, len(purpose)+len(scope)+22)
	message = append(message, purpose...)
	message = append(message, 0)
	message = append(message, scope...)
	message = append(message, 0)
	message = strconv.AppendInt(message, epoch, 10)
	token, err := keys.DeriveHMACSHA256(ctx, message)
	if err != nil || ctx.Err() != nil {
		clear(token[:])
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// A clear or a configuration change during derivation, or a concurrent
	// derivation of a later epoch, discards this token.
	if s.erased != erased || s.scope != scope || s.period != period || epoch < s.floor {
		clear(token[:])
		return nil
	}
	s.floor = epoch
	return &model.RotatingLinkage{Scope: scope, Epoch: epoch, Token: token}
}

// Erase drops the epoch floor and discards any derivation in progress. It is
// the linkage eraser of persistence.Clear: it creates no secret and sends
// nothing. Erasing an empty source succeeds.
// requirement-ref: REQ-LINKAGE-0023
func (s *Source) Erase(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scope, s.period, s.floor = "", 0, -1
	s.erased++
	return nil
}
