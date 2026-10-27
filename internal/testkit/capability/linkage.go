package capability

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"sync"
)

// LinkageKeys is a synthetic, explicitly test-scoped linkage key store. It is
// not protected storage and must never become a mobile fallback. Like the
// platform contract it creates its key lazily on derivation, deletes it on
// Delete and never returns it. A store made with a fixed synthetic key uses
// that key until its first Delete, and a random one afterwards. It records
// every derivation message, so a test can assert the exact HMAC input.
type LinkageKeys struct {
	mu       sync.Mutex
	key      []byte
	messages [][]byte
	created  int
	deletes  int
	fail     bool
	// OnDerive, when set, runs before each derivation without the lock held,
	// so a test can interleave another operation.
	OnDerive func()
}

// NewLinkageKeys returns a store holding the fixed synthetic key, as if it
// had already been created.
func NewLinkageKeys(key [32]byte) *LinkageKeys { return &LinkageKeys{key: key[:]} }

// Fail makes every later derivation and deletion fail until Recover.
func (k *LinkageKeys) Fail()    { k.mu.Lock(); k.fail = true; k.mu.Unlock() }
func (k *LinkageKeys) Recover() { k.mu.Lock(); k.fail = false; k.mu.Unlock() }

func (k *LinkageKeys) DeriveHMACSHA256(ctx context.Context, message []byte) ([32]byte, error) {
	if k.OnDerive != nil {
		k.OnDerive()
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	k.messages = append(k.messages, append([]byte(nil), message...))
	if k.fail || ctx.Err() != nil {
		return [32]byte{}, ErrUnavailable
	}
	if k.key == nil {
		k.key = make([]byte, 32)
		_, _ = rand.Read(k.key)
		k.created++
	}
	m := hmac.New(sha256.New, k.key)
	m.Write(message)
	var token [32]byte
	copy(token[:], m.Sum(nil))
	return token, nil
}

func (k *LinkageKeys) Delete(context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.fail {
		return ErrUnavailable
	}
	clear(k.key)
	k.key = nil
	k.deletes++
	return nil
}

// Messages returns copies of the derivation messages received so far.
func (k *LinkageKeys) Messages() [][]byte {
	k.mu.Lock()
	defer k.mu.Unlock()
	out := make([][]byte, len(k.messages))
	for i, m := range k.messages {
		out[i] = append([]byte(nil), m...)
	}
	return out
}

// Held reports whether a key exists, how many keys were created after a
// deletion or on first use, and how many deletions succeeded.
func (k *LinkageKeys) Held() (held bool, created, deletes int) {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.key != nil, k.created, k.deletes
}
