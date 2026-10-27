package tls01

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

var optionLimits = jsondecode.Limits{MaxBytes: 8192, MaxDepth: 8, MaxNodes: 64}

func alpnOptions(ids ...string) string {
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = fmt.Sprintf("%q", id)
	}
	return `{"alpn_protocols":[` + strings.Join(quoted, ",") + `]}`
}

func normalizeALPN(t *testing.T, options string) (Parameters, error) {
	t.Helper()
	c := Codec{}
	d := c.Definition()
	in, err := jsondecode.Decode([]byte(`{"ip_address":"8.8.8.8","port":443,"server_name":"measurement.example"}`), d.Input, optionLimits)
	if err != nil {
		t.Fatal(err)
	}
	opt, err := jsondecode.Decode([]byte(options), d.Options, optionLimits)
	if err != nil {
		return Parameters{}, err
	}
	p, err := c.Normalize(in, opt)
	if err != nil {
		return Parameters{}, err
	}
	return p.(Parameters), nil
}

// M-TLS-01-OPTIONS ALPN bounds in isolation. Only h2, http/1.1 and dot are
// registered offer identifiers (REQ-OBS-0049), and duplicates are rejected,
// so no admissible offer can hold more than three entries or an identifier
// longer than eight bytes. Through the Client a nine-entry list therefore
// always also repeats or is unregistered, and a 33-byte identifier is always
// unregistered. The count bound is isolated here at the closed options
// descriptor, which applies it before membership: 8 distinct well-formed
// identifiers pass the descriptor and 9 do not. The 1..32 byte bound has no
// separate descriptor stage; the registered set, which lies inside it, is the
// stricter check, so 32- and 33-byte identifiers are both rejected and the
// registered identifiers at both ends of the set's length range are accepted.
// requirement-ref: REQ-TLS-01-0016
// requirement-ref: REQ-OBS-0049
func TestMTls01ALPNBoundsIsolated(t *testing.T) {
	d := Codec{}.Definition()
	var distinct []string
	for i := 0; i < 9; i++ {
		distinct = append(distinct, fmt.Sprintf("p%d", i))
	}
	if _, err := jsondecode.Decode([]byte(alpnOptions(distinct[:8]...)), d.Options, optionLimits); err != nil {
		t.Fatal("8 distinct identifiers rejected by the count bound:", err)
	}
	if _, err := jsondecode.Decode([]byte(alpnOptions(distinct...)), d.Options, optionLimits); err == nil {
		t.Fatal("9 distinct identifiers passed the count bound")
	}
	// Past the descriptor, membership still refuses unregistered entries.
	if _, err := normalizeALPN(t, alpnOptions(distinct[:8]...)); err == nil {
		t.Fatal("unregistered identifiers admitted")
	}
	for _, n := range []int{1, 32, 33} {
		if _, err := normalizeALPN(t, alpnOptions(strings.Repeat("a", n))); err == nil {
			t.Fatalf("unregistered %d-byte identifier admitted", n)
		}
	}
	for _, tc := range [][]string{{}, {"h2"}, {"dot"}, {"http/1.1"}, {"dot", "http/1.1", "h2"}} {
		p, err := normalizeALPN(t, alpnOptions(tc...))
		if err != nil || strings.Join(p.ALPN(), ",") != strings.Join(tc, ",") {
			t.Fatalf("%v: %v %v", tc, p.ALPN(), err)
		}
	}
	for _, tc := range [][]string{{"h2", "h2"}, {"dot", "http/1.1", "h2", "h2"}, {""}} {
		if _, err := normalizeALPN(t, alpnOptions(tc...)); err == nil {
			t.Fatalf("%v admitted", tc)
		}
	}
}
