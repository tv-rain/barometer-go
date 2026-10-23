package networkcontext_test

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/networkcontext"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

// Access-context family cases (DEV-644). The fake platform network holds
// synthetic canaries for every forbidden item beside the access state it
// reports. Logs, cache entries and egress-endpoint requests named by the case
// plan do not exist in this build: the core writes no log, and the context
// cache and egress lookup belong to DEV-647, so those surfaces are not
// exercised. These are synthetic core contract cases, not platform evidence.

var (
	networkTypes = map[platform.NetworkType]string{platform.NetworkWiFi: "wifi", platform.NetworkCellular: "cellular", platform.NetworkEthernet: "ethernet", platform.NetworkOther: "other", platform.NetworkUnknown: "unknown"}
	vpnStates    = map[platform.VPNState]string{platform.VPNDetected: "detected", platform.VPNNotDetected: "not_detected", platform.VPNUnknown: "unknown"}
	availability = map[platform.Availability]string{platform.AvailabilityYes: "yes", platform.AvailabilityNo: "no", platform.AvailabilityUnknown: "unknown"}
)

// encodedResult builds a completed synthetic TCP-01 result carrying nc and
// returns its encoding.
func encodedResult(t *testing.T, d *networkcontext.Source, snap networkcontext.Snapshot, nc observation.NetworkContext) []byte {
	t.Helper()
	dec, err := result.New(result.Binding{TaskID: "campaign-1", Output: result.Output{Test: model.Test{ID: "TCP-01", Version: result.Contract, ResultVersion: result.Contract}},
		MaxAttempts: 1, MaxDurationMS: 30000, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"},
		ContextDecoder: d.Decoder(), Path: snap.Path(), AccessGeneration: snap.Generation()})
	if err != nil {
		t.Fatal(err)
	}
	o, err := observation.DecodeProtocol(observation.Contract, []byte(`{"type":"tcp_connect","status":"succeeded","duration_ms":0,"details":{"address_family":"ipv4"}}`), observation.ProtocolLimits{})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	timing, err := result.Timing(start, start)
	if err != nil {
		t.Fatal(err)
	}
	// A changed snapshot belongs to an attempt the change ended.
	r := model.Result{SchemaVersion: result.Contract, TaskID: "campaign-1", RunID: result.NewRunID(), Attempt: 1,
		Test: model.Test{ID: "TCP-01", Version: result.Contract, ResultVersion: result.Contract}, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"},
		Time: timing, Execution: model.Execution{State: "incomplete", Reason: "network_changed"}, NetworkContext: &nc}
	if !nc.Access().NetworkChanged {
		r.Execution = model.Execution{State: "completed"}
		r.Measurement = &model.Measurement{Outcome: "connected", Observations: []observation.Observation{o}}
	}
	v, err := dec.Build(r)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func assertNoCanary(t *testing.T, where string, b []byte) {
	t.Helper()
	for _, c := range capability.Canaries {
		if bytes.Contains(b, []byte(c)) {
			t.Fatalf("%s carries forbidden value %q", where, c)
		}
	}
}

// CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT: the core takes access context only
// through platform.AccessState, which has no field for a forbidden item, and
// the serialized access object holds exactly the six registered fields with
// the reported normalized values, for every value of each field. Neither the
// access object nor the Measurement Result carries a canary. An access input
// with any additional field is rejected.
// requirement-ref: REQ-CONTEXT-0002
// requirement-ref: REQ-CONTEXT-0003
func TestCcContextContextLayersAccessContext(t *testing.T) {
	want := []string{"ipv4_available", "ipv6_available", "network_changed", "network_type", "validated", "vpn_state"}
	// The port's reading has only the coarse fields and the generation.
	if got := stateFields(); !slices.Equal(got, []string{"Generation", "IPv4", "IPv6", "Type", "VPN", "Validated"}) {
		t.Fatalf("platform.AccessState fields = %v", got)
	}
	n := capability.NewNetwork(platform.AccessState{Generation: 1})
	if len(n.Withheld) != len(capability.Canaries) {
		t.Fatal("fake platform holds no canaries")
	}
	d, err := networkcontext.NewDirect(n)
	if err != nil {
		t.Fatal(err)
	}
	check := func(s platform.AccessState) {
		t.Helper()
		n.Set(s)
		snap, err := d.Capture(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for _, changed := range []bool{false, true} {
			nc, err := snap.Context(changed)
			if err != nil {
				t.Fatal(err)
			}
			b, err := json.Marshal(nc)
			if err != nil {
				t.Fatal(err)
			}
			var wire struct {
				Access map[string]any `json:"access"`
			}
			if err := json.Unmarshal(b, &wire); err != nil {
				t.Fatal(err)
			}
			keys := make([]string, 0, len(wire.Access))
			for k := range wire.Access {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			a := wire.Access
			if !slices.Equal(keys, want) || a["network_type"] != networkTypes[s.Type] || a["vpn_state"] != vpnStates[s.VPN] ||
				a["ipv4_available"] != availability[s.IPv4] || a["ipv6_available"] != availability[s.IPv6] || a["validated"] != availability[s.Validated] || a["network_changed"] != changed {
				t.Fatalf("access object %s for %+v", b, s)
			}
			assertNoCanary(t, "access object", b)
			assertNoCanary(t, "Measurement Result", encodedResult(t, d, snap, nc))
		}
	}
	for typ := range networkTypes {
		check(platform.AccessState{Generation: 1, Type: typ})
	}
	for vpn := range vpnStates {
		check(platform.AccessState{Generation: 1, VPN: vpn})
	}
	for v := range availability {
		check(platform.AccessState{Generation: 1, IPv4: v})
		check(platform.AccessState{Generation: 1, IPv6: v})
		check(platform.AccessState{Generation: 1, Validated: v})
	}
	for _, extra := range []string{`"interface":"canary-if-wlan7"`, `"local_ipv4":"198.51.100.77"`, `"ssid":"canary-ssid-7f3a"`, `"location":"55.7777,37.7777"`} {
		in := `{"network_type":"wifi","vpn_state":"unknown","ipv4_available":"yes","ipv6_available":"unknown","validated":"unknown","network_changed":false,` + extra + `}`
		if _, err := observation.DecodeAccess(observation.Contract, []byte(in), 1); err == nil {
			t.Fatalf("access input with %s accepted", extra)
		}
	}
}

// CC-CONTEXT-FORMAT-ACCESS-OBJECT: every registered access value is accepted
// and produced; case variants, unregistered strings, a non-boolean
// network_changed and a missing field are rejected; values the build does not
// recognize from the platform are normalized to unknown, never guessed; and
// no forbidden item appears in an encoded access object or result.
// requirement-ref: REQ-CONTEXT-FORMAT-0004
// requirement-ref: REQ-CONTEXT-FORMAT-0005
func TestCcContextFormatAccessObject(t *testing.T) {
	base := map[string]any{"network_type": "wifi", "vpn_state": "unknown", "ipv4_available": "yes", "ipv6_available": "unknown", "validated": "unknown", "network_changed": false}
	decode := func(edit func(map[string]any)) error {
		v := map[string]any{}
		for k, x := range base {
			v[k] = x
		}
		edit(v)
		b, _ := json.Marshal(v)
		_, err := observation.DecodeAccess(observation.Contract, b, 1)
		return err
	}
	accept := map[string][]any{
		"network_type": {"wifi", "cellular", "ethernet", "other", "unknown"}, "vpn_state": {"detected", "not_detected", "unknown"},
		"ipv4_available": {"yes", "no", "unknown"}, "ipv6_available": {"yes", "no", "unknown"}, "validated": {"yes", "no", "unknown"},
		"network_changed": {true, false},
	}
	for field, values := range accept {
		for _, v := range values {
			if err := decode(func(m map[string]any) { m[field] = v }); err != nil {
				t.Fatalf("%s %v rejected", field, v)
			}
		}
	}
	reject := map[string][]any{
		"network_type": {"WIFI", "Wifi", "bluetooth", "", nil}, "vpn_state": {"Detected", "maybe", nil},
		"ipv4_available": {"Yes", "true", true}, "ipv6_available": {"NO", "partial"}, "validated": {"Yes", "captive", false},
		"network_changed": {"false", nil, 0},
	}
	for field, values := range reject {
		for _, v := range values {
			if err := decode(func(m map[string]any) { m[field] = v }); err == nil {
				t.Fatalf("%s %v accepted", field, v)
			}
		}
		if err := decode(func(m map[string]any) { delete(m, field) }); err == nil {
			t.Fatalf("missing %s accepted", field)
		}
	}
	for _, extra := range []string{"interface_name", "address", "ssid", "bssid", "dns_server", "cell_id", "sim", "carrier", "subscriber_id", "location"} {
		if err := decode(func(m map[string]any) { m[extra] = "canary-" + extra }); err == nil {
			t.Fatalf("extra %s accepted", extra)
		}
	}
	// Values outside the registered platform enumerations stay unknown.
	a := networkcontext.Normalize(platform.AccessState{Type: 99, VPN: 7, IPv4: 9, IPv6: 9, Validated: 9})
	if a != (observation.Access{NetworkType: "unknown", VPNState: "unknown", IPv4Available: "unknown", IPv6Available: "unknown", Validated: "unknown"}) {
		t.Fatalf("unrecognized values normalized to %+v", a)
	}
	n := capability.NewNetwork(platform.AccessState{Generation: 3, Type: platform.NetworkCellular, VPN: platform.VPNDetected, IPv4: platform.AvailabilityYes, IPv6: platform.AvailabilityYes, Validated: platform.AvailabilityYes})
	d, err := networkcontext.NewDirect(n)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := d.Capture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	nc, _ := snap.Context(false)
	b, _ := json.Marshal(nc)
	assertNoCanary(t, "access object", b)
	r := encodedResult(t, d, snap, nc)
	assertNoCanary(t, "Measurement Result", r)
	// The generation is local metadata and never serialized.
	if strings.Contains(string(r), "generation") || strings.Contains(string(b), "generation") {
		t.Fatal("generation serialized")
	}
}

// stateFields lists the fields of platform.AccessState in sorted order.
func stateFields() []string {
	typ := reflect.TypeOf(platform.AccessState{})
	out := make([]string, 0, typ.NumField())
	for i := range typ.NumField() {
		out = append(out, typ.Field(i).Name)
	}
	slices.Sort(out)
	return out
}
