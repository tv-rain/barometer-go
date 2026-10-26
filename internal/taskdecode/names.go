package taskdecode

import (
	"net/netip"
	"strings"

	"golang.org/x/net/idna"
)

// Explicit nontransitional UTS #46 lookup mapping with STD3, context, bidi and
// DNS-length checks. Dependency and Unicode tables are pinned by go.mod.
var names = idna.New(idna.MapForLookup(), idna.Transitional(false), idna.StrictDomainName(true), idna.ValidateLabels(true), idna.CheckHyphens(true), idna.CheckJoiners(true), idna.BidiRule(), idna.VerifyDNSLength(true))

// DNSName deterministically normalizes a task-supplied DNS name to lowercase
// ASCII of at most 253 bytes. It rejects IP literals, wildcards and other
// non-STD3 labels, single-label and trailing-dot names, and special-use and
// local suffixes. It never resolves the name.
func DNSName(s string) (string, bool) {
	if len(s) == 0 || len(s) > 4096 || strings.HasSuffix(s, ".") {
		return "", false
	}
	n, e := names.ToASCII(s)
	n = strings.ToLower(n)
	if e != nil || len(n) > 253 || !strings.Contains(n, ".") || strings.HasSuffix(n, ".") {
		return "", false
	}
	if _, e := netip.ParseAddr(n); e == nil {
		return "", false
	}
	for _, suffix := range []string{"local", "localhost", "internal", "lan", "home", "home.arpa", "onion", "invalid", "test"} {
		if n == suffix || strings.HasSuffix(n, "."+suffix) {
			return "", false
		}
	}
	return n, true
}
