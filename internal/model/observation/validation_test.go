package observation

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

// requirement-ref: REQ-JSON-0008
// requirement-ref: REQ-JSON-0011
func TestProtocolValidationPrecedenceAndStableEncoding(t *testing.T) {
	for _, tc := range []struct{ name, input string }{
		{"dns", dnsOK}, {"tcp", tcpOK}, {"tls", tlsOK}, {"http", httpOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := DecodeProtocol(Contract, []byte(tc.input), protocolLimits)
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			for range 20 {
				again, err := DecodeProtocol(Contract, canonical, protocolLimits)
				if err != nil {
					t.Fatal(err)
				}
				next, err := json.Marshal(again)
				if err != nil || !bytes.Equal(canonical, next) {
					t.Fatal("normalization changed on round trip")
				}
			}
			bad := strings.Replace(tc.input, `"duration_ms":`, `"private":"DO_NOT_RETAIN","duration_ms":`, 1)
			for _, contract := range []string{"9.0.0", Contract} {
				want := error(jsondecode.ErrField)
				if contract != Contract {
					want = ErrVersion
				}
				for range 20 {
					v, err := DecodeProtocol(contract, []byte(bad), protocolLimits)
					if !errors.Is(err, want) || len(err.Error()) > 128 {
						t.Fatalf("precedence: %v", err)
					}
					if _, err := v.MarshalJSON(); err == nil {
						t.Fatal("partial protocol value escaped")
					}
				}
			}
		})
	}
}

// requirement-ref: REQ-CONTEXT-FORMAT-0006
// requirement-ref: REQ-CONTEXT-FORMAT-0018
// requirement-ref: REQ-JSON-0002
func TestContextNormalizationStaysWithinASCIICodes(t *testing.T) {
	d, err := NewContextDecoder(Contract, ContextApprovals{ClassificationVersions: []string{"0.1.0"}, CountryCodes: []string{"US"}, RegionCodes: []string{"US-SC"}})
	if err != nil {
		t.Fatal(err)
	}
	path, err := NewPath(1, 0, "direct", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		country, region string
		valid           bool
	}{
		{"US", "US-SC", true}, {"us", "us-sc", true}, {"uS", "Us-sC", true},
		{"uſ", "us-sc", false}, {"us", "us-ſc", false}, {" us", "us-sc", false},
	} {
		t.Run(tc.country+"/"+tc.region, func(t *testing.T) {
			input := strings.ReplaceAll(strings.ReplaceAll(contextOK, "NL-NH", tc.region), `"NL"`, `"`+tc.country+`"`)
			v, err := d.Decode([]byte(input), path)
			if !tc.valid {
				if !errors.Is(err, ErrValue) {
					t.Fatalf("invalid code repaired: %v", err)
				}
				if _, err := v.MarshalJSON(); err == nil {
					t.Fatal("partial context escaped")
				}
				return
			}
			if err != nil || v.Egress().CountryCode != "US" || v.Egress().RegionCode != "US-SC" {
				t.Fatalf("normalization: %v", err)
			}
			canonical, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			again, err := d.Decode(canonical, path)
			if err != nil || !again.Matches(path) {
				t.Fatalf("path changed on round trip: %v", err)
			}
			next, err := json.Marshal(again)
			if err != nil || !bytes.Equal(canonical, next) {
				t.Fatal("context normalization not idempotent")
			}
		})
	}
}
