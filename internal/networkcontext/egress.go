package networkcontext

import (
	"context"
	"crypto/x509"
	"errors"
	"net/netip"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Registered values (REQ-CONTEXT-0015, REQ-CONTEXT-0022).
const (
	// MaxResponseBytes bounds the endpoint's JSON response body.
	MaxResponseBytes = 4096
	// OperationTimeout bounds each primary or fallback operation.
	OperationTimeout = 5 * time.Second
	// DefaultFreshness is the in-memory cache lifetime this build uses.
	DefaultFreshness = 300 * time.Second
	// MaxFreshness is the registered maximum cache lifetime.
	MaxFreshness = 900 * time.Second
)

// Reviewed implementation inputs (DEV-647): the most bytes of the response
// status line and headers, and the most cache entries kept at once. One
// access-network generation needs at most a direct and a proxy entry; the
// rest is slack for a proxy generation that has not yet been forgotten.
const (
	MaxHeaderBytes  = 8192
	maxCacheEntries = 8
)

var (
	ErrLookup   = errors.New("networkcontext: invalid lookup configuration")
	errResponse = errors.New("networkcontext: invalid endpoint response")
)

// Fetch makes one HTTPS GET to endpoint through the attempt's path and
// returns the body of a valid response, at most MaxResponseBytes. The socket
// owner, internal/execution, supplies it bound to a dialer of the attempt's
// own path; this package opens no connection. ctx carries the operation
// bound. roots nil selects the system trust roots.
type Fetch func(ctx context.Context, endpoint *url.URL, roots *x509.CertPool) ([]byte, error)

// Clock is the engine clock: it times cache freshness and the operation
// bound, so both follow the same time source as the attempt deadline.
type Clock interface {
	Now() time.Time
	AfterFunc(d time.Duration, f func()) (stop func() bool)
}

// Endpoints are the host-configured primary and optional fallback URLs. The
// host configuration has already validated them against reviewed approvals
// (REQ-CONTEXT-0007); no task can supply or change them.
type Endpoints struct{ Primary, Fallback string }

// Lookup acquires egress classifications for one initialized generation. It
// is immutable; its cache entries belong to the cache epoch it was created
// in, so a generation never reads or writes another generation's entries.
type Lookup struct {
	endpoints []*url.URL // primary, then the optional fallback
	approvals observation.ContextApprovals
	roots     *x509.CertPool // nil selects the system roots
	clock     Clock
	cache     *Cache
	epoch     uint64
	timeout   time.Duration
	freshness time.Duration
}

// NewLookup binds the endpoints, the reviewed classification approvals
// (classification versions, country and region codes; connection codes come
// from the path) and the trust roots to cache. Without classification
// approvals every response is invalid, so egress stays unavailable.
func NewLookup(e Endpoints, a observation.ContextApprovals, roots *x509.CertPool, clock Clock, cache *Cache) (*Lookup, error) {
	if clock == nil || cache == nil {
		return nil, ErrLookup
	}
	a.ConnectionCodes = nil
	if _, err := observation.NewContextDecoder(observation.Contract, a); err != nil {
		return nil, ErrLookup
	}
	l := &Lookup{approvals: a, roots: roots, clock: clock, cache: cache, epoch: cache.Epoch(), timeout: OperationTimeout, freshness: DefaultFreshness}
	for i, raw := range []string{e.Primary, e.Fallback} {
		if raw == "" && i == 1 {
			break
		}
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
			return nil, ErrLookup
		}
		l.endpoints = append(l.endpoints, u)
	}
	return l, nil
}

// classification is one valid endpoint response. It holds no IP address,
// endpoint or request detail.
type classification struct {
	country, region, version string
	asn                      *uint32
}

// acquire returns the classification for the key in the bounded order of
// REQ-CONTEXT-0043: a fresh cache entry, otherwise one request to the
// primary, otherwise one request to the fallback through the same dialer.
// It reports the source, or false when no endpoint produced a valid
// response. It never retries an endpoint, never changes the path and caches
// no failure. Each operation ends with ctx and within the operation bound.
// requirement-ref: REQ-CONTEXT-0015
// requirement-ref: REQ-CONTEXT-0043
// requirement-ref: REQ-CONTEXT-0021
func (l *Lookup) acquire(ctx context.Context, key cacheKey, fetch Fetch, valid func(classification) bool) (classification, string, bool) {
	if c, ok := l.cache.get(l.epoch, key, l.clock.Now(), l.freshness); ok && valid(c) {
		return c, "cache", true
	}
	for i, u := range l.endpoints {
		if ctx.Err() != nil {
			break
		}
		if key.path == "proxy" {
			if a, err := netip.ParseAddr(u.Hostname()); err != nil || !task.PublicAddress(a) {
				continue
			}
		}
		c, err := l.fetch(ctx, u, fetch)
		if err != nil || !valid(c) {
			continue
		}
		l.cache.put(l.epoch, key, c, l.clock.Now())
		return c, [...]string{"primary", "fallback"}[i], true
	}
	return classification{}, "", false
}

func (l *Lookup) fetch(ctx context.Context, u *url.URL, fetch Fetch) (classification, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := l.clock.AfterFunc(l.timeout, cancel)
	defer stop()
	body, err := fetch(ctx, u, l.roots)
	if err != nil || len(body) > MaxResponseBytes || ctx.Err() != nil {
		return classification{}, errResponse
	}
	return decodeResponse(body)
}

var responseSchema = &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
	"country_code":           {Schema: &jsondecode.Schema{Kind: jsondecode.String}},
	"region_code":            {Schema: &jsondecode.Schema{Kind: jsondecode.String}},
	"asn":                    {Schema: &jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "4294967295"}},
	"classification_version": {Schema: &jsondecode.Schema{Kind: jsondecode.String}, Required: true},
}}

// decodeResponse strictly decodes the endpoint object (REQ-CONTEXT-0013,
// REQ-CONTEXT-0014): unknown fields, duplicates and wrong types are invalid.
// A response that omits every optional classification is valid
// (REQ-CONTEXT-0016). Syntax and approval of the values are checked when the
// classification is bound to a path.
// requirement-ref: REQ-CONTEXT-0013
// requirement-ref: REQ-CONTEXT-0014
// requirement-ref: REQ-CONTEXT-0016
func decodeResponse(body []byte) (classification, error) {
	v, err := jsondecode.Decode(body, responseSchema, jsondecode.Limits{MaxBytes: MaxResponseBytes, MaxDepth: 2, MaxNodes: 8})
	if err != nil {
		return classification{}, errResponse
	}
	text := func(name string) string {
		if x, ok := v.Lookup(name); ok {
			return x.Text()
		}
		return ""
	}
	c := classification{country: text("country_code"), region: text("region_code"), version: text("classification_version")}
	if x, ok := v.Lookup("asn"); ok {
		n, err := strconv.ParseUint(x.Text(), 10, 32)
		if err != nil {
			return classification{}, errResponse
		}
		asn := uint32(n)
		c.asn = &asn
	}
	for _, name := range []string{"country_code", "region_code"} {
		if x, ok := v.Lookup(name); ok && x.Text() == "" {
			return classification{}, errResponse
		}
	}
	return c, nil
}

// cacheKey identifies one access-network generation and selected-path
// generation. It holds no address or persistent identifier (REQ-CONTEXT-0024);
// both generations are local counters that are never serialized.
type cacheKey struct {
	network, proxy uint64
	path           string
}

type cacheEntry struct {
	value classification
	at    time.Time
}

// Cache is the in-memory egress cache of one client. It is never persisted.
// Clear drops every entry and starts a new epoch, so a lookup created before
// it can neither read nor write entries afterwards: shutdown,
// reinitialization, policy replacement and clearing all clear it. Entries of
// another access-network generation are dropped when a newer one is stored,
// and ForgetProxy drops proxy entries when the proxy changes.
// requirement-ref: REQ-CONTEXT-0022
// requirement-ref: REQ-CONTEXT-0023
// requirement-ref: REQ-CONTEXT-0024
// requirement-ref: REQ-PROXY-0016
type Cache struct {
	mu      sync.Mutex
	epoch   uint64
	entries map[cacheKey]cacheEntry
}

func NewCache() *Cache { return &Cache{epoch: 1} }

// Epoch is the current epoch, which a new Lookup binds to.
func (c *Cache) Epoch() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.epoch
}

// Clear drops every entry and starts a new epoch.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.epoch++
	c.entries = nil
}

// Erase is the cache's clearing eraser. It sends nothing and cannot fail.
func (c *Cache) Erase(context.Context) error {
	c.Clear()
	return nil
}

// ForgetProxy drops every proxy entry: a proxy replacement or clearing
// invalidates the previous generation's contexts (REQ-PROXY-0016).
func (c *Cache) ForgetProxy() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.entries {
		if k.path == "proxy" {
			delete(c.entries, k)
		}
	}
}

// Close clears the cache when it is still in l's epoch, so a generation that
// ends drops its entries without disturbing a successor's.
func (l *Lookup) Close() {
	if l == nil {
		return
	}
	l.cache.mu.Lock()
	defer l.cache.mu.Unlock()
	if l.cache.epoch == l.epoch {
		l.cache.epoch++
		l.cache.entries = nil
	}
}

// Len reports the number of entries, for tests and clearing checks.
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

func (c *Cache) get(epoch uint64, k cacheKey, now time.Time, freshness time.Duration) (classification, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[k]
	if !ok || epoch != c.epoch {
		return classification{}, false
	}
	// An entry is reused up to and including its freshness bound
	// (REQ-LIMIT-0023) and never after the registered maximum.
	if age := now.Sub(e.at); age < 0 || age > min(freshness, MaxFreshness) {
		delete(c.entries, k)
		return classification{}, false
	}
	return e.value, true
}

func (c *Cache) put(epoch uint64, k cacheKey, v classification, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if epoch != c.epoch {
		return
	}
	if c.entries == nil {
		c.entries = make(map[cacheKey]cacheEntry)
	}
	for old := range c.entries {
		if old.network != k.network || len(c.entries) >= maxCacheEntries {
			delete(c.entries, old)
		}
	}
	c.entries[k] = cacheEntry{value: v, at: now}
}
