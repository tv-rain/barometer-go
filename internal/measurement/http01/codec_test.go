package http01

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

func decodeParameters(input, options string) (task.Parameters, error) {
	c := Codec{}
	d := c.Definition()
	lim := jsondecode.Limits{MaxBytes: 65536, MaxDepth: 32, MaxNodes: 4096}
	i, e := jsondecode.Decode([]byte(input), d.Input, lim)
	if e != nil {
		return nil, e
	}
	o, e := jsondecode.Decode([]byte(options), d.Options, lim)
	if e != nil {
		return nil, e
	}
	return c.Normalize(i, o)
}

// requirement-ref: REQ-HTTP-01-0010
// requirement-ref: REQ-HTTP-01-0011
// requirement-ref: REQ-HTTP-01-0013
// requirement-ref: REQ-HTTP-01-0020
func TestCodecBoundsAndNormalization(t *testing.T) {
	base := "https://service.example/"
	for _, n := range []int{2047, 2048, 2049} {
		url := base + strings.Repeat("a", n-len(base))
		_, e := decodeParameters(fmt.Sprintf(`{"url":%q}`, url), `{}`)
		if (e == nil) != (n <= 2048) {
			t.Fatalf("URL bytes %d", n)
		}
	}
	for _, body := range []bool{false, true} {
		for _, count := range []int{0, 1, 8, 9} {
			for _, size := range []int{1, 256, 257} {
				list := make([]any, count)
				for i := range list {
					if body {
						list[i] = map[string]any{"type": "contains", "value": strings.Repeat("a", size)}
					} else {
						list[i] = map[string]any{"name": "X-Fixture", "type": "equals", "value": strings.Repeat("a", size)}
					}
				}
				key := "header_checks"
				if body {
					key = "body_checks"
				}
				b, _ := json.Marshal(map[string]any{"url": base, key: list})
				_, e := decodeParameters(string(b), `{}`)
				valid := count <= 8 && (count == 0 || size <= 256)
				if (e == nil) != valid {
					t.Fatalf("checks body=%t count=%d value=%d", body, count, size)
				}
			}
		}
	}
	for _, url := range []string{"HTTPS://SERVICE.Example:443", "https://service.example/", "https://[2606:4700:4700::1111]/"} {
		p, e := decodeParameters(fmt.Sprintf(`{"url":%q}`, url), `{}`)
		if e != nil {
			t.Fatal(url, e)
		}
		hp := p.(Parameters)
		if strings.Contains(hp.Target().URL, "SERVICE") || strings.Contains(hp.Target().URL, ":443") {
			t.Fatal("not canonical")
		}
	}
	for _, url := range []string{"https://service.example:0/", "https://service.example:65536/", "https://service.example:/", "https://service.example/#", "https://{id}.example/", "https://[::ffff:8.8.8.8]/"} {
		if _, e := decodeParameters(fmt.Sprintf(`{"url":%q}`, url), `{}`); e == nil {
			t.Fatal(url)
		}
	}
	p, e := decodeParameters(`{"url":"https://service.example/","body_checks":[{"type":"sha256","value":"`+strings.Repeat("a", 64)+`"}],"header_checks":[{"name":"X-Fixture","type":"present"}]}`, `{}`)
	if e != nil {
		t.Fatal(e)
	}
	hp := p.(Parameters)
	clone := hp.Target()
	clone.HeaderChecks[0].Name = "changed"
	clone.BodyChecks[0].Value = "changed"
	if hp.Target().HeaderChecks[0].Name != "x-fixture" || hp.Target().BodyChecks[0].Value != strings.Repeat("a", 64) {
		t.Fatal("mutable tuple")
	}
}
func TestProxyShape(t *testing.T) {
	for _, tc := range []struct {
		url, opt string
		ok       bool
	}{{"https://8.8.8.8/", `{}`, true}, {"https://[2606:4700:4700::1111]/", `{}`, true}, {"https://service.example/", `{}`, false}, {"https://8.8.8.8/", `{"address_family":"ipv4"}`, false}} {
		p, e := decodeParameters(fmt.Sprintf(`{"url":%q}`, tc.url), tc.opt)
		if e != nil {
			t.Fatal(e)
		}
		if (ProxyRule{}).Compatible(p) != tc.ok {
			t.Fatal(tc)
		}
	}
}
func FuzzHttp01Parameters(f *testing.F) {
	for _, in := range []string{`{"url":"https://8.8.8.8/"}`, `{"url":"https://service.example/","header_checks":[{"name":"X","type":"absent"}]}`, `{}`, `null`} {
		f.Add(in, `{}`)
	}
	f.Fuzz(func(t *testing.T, in, opt string) {
		p, e := decodeParameters(in, opt)
		if e == nil {
			v := p.(Parameters)
			if v.Target().URL == "" || v.body < 1 || v.body > 65536 || !validChecks(v.Target()) {
				t.Fatal("invalid normalized parameters")
			}
		}
	})
}
