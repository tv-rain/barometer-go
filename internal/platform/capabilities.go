// Package platform owns narrow internal capability contracts consumed by the
// portable core. Concrete reviewed adapters are not implemented here.
package platform

import (
	"context"
	"time"
)

// Clock is installed internally by a reviewed integration. Unavailable or
// invalid time is an error, never an implicit system-clock fallback. Implementors
// honor cancellation and a finite caller deadline without background work.
type Clock interface {
	Now(context.Context) (time.Time, error)
}

// LinkageKeyStore never exports the installation secret. Derivation lazily
// creates protected, application-scoped, non-synchronizing key material; Delete
// is idempotent. Implementors honor finite deadlines and fail when protection
// is unavailable. Ordinary-file or host-provided key fallback is forbidden.
type LinkageKeyStore interface {
	DeriveHMACSHA256(context.Context, []byte) ([32]byte, error)
	Delete(context.Context) error
}
