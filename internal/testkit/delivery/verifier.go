// Package delivery owns a synthetic remote-delivery verifier for tests only.
// Production code must not import it. Its envelope is a test format, not a
// reviewed authentication or signature mechanism, and confers no authority.
package delivery

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// ErrUnverified is the single verification error. It carries no detail.
var ErrUnverified = errors.New("testkit/delivery: delivery not verified")

// Verifier models an authenticated, integrity-protected channel for one
// synthetic issuer. The envelope is an HMAC-SHA256 tag over the task bytes
// followed by those bytes. Issuer, generation and issuance time are what the
// channel established, not values read from the task.
//
// IssuerAuthenticated false models transport encryption without issuer
// authentication (REQ-TRUST-0005): integrity may hold, but Verify never
// returns a source, so no record can be installed.
type Verifier struct {
	Key                 []byte
	IssuerID            string
	PolicyGeneration    uint64
	IssuedAt            task.Timestamp
	IssuerAuthenticated bool
}

// Seal returns an envelope for exactTask under v.Key.
func (v Verifier) Seal(exactTask []byte) []byte {
	return append(v.tag(exactTask), exactTask...)
}

func (v Verifier) tag(exactTask []byte) []byte {
	m := hmac.New(sha256.New, v.Key)
	m.Write(exactTask)
	return m.Sum(nil)
}

// Verify returns the authenticated remote source and an owned copy of the
// exact task bytes, or ErrUnverified.
func (v Verifier) Verify(ctx context.Context, delivery []byte) (task.SourceContext, []byte, error) {
	if ctx.Err() != nil || len(v.Key) == 0 || len(delivery) <= sha256.Size {
		return task.SourceContext{}, nil, ErrUnverified
	}
	tag, body := delivery[:sha256.Size], delivery[sha256.Size:]
	if !hmac.Equal(tag, v.tag(body)) || !v.IssuerAuthenticated {
		return task.SourceContext{}, nil, ErrUnverified
	}
	issued := v.IssuedAt
	source, err := task.NewSourceContext(task.SourceInput{Origin: task.Remote, IssuerID: v.IssuerID, PolicyGeneration: v.PolicyGeneration, Authenticated: true, IssuedAt: &issued})
	if err != nil {
		return task.SourceContext{}, nil, ErrUnverified
	}
	return source, append([]byte(nil), body...), nil
}
