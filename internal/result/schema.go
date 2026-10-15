package result

import (
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"strconv"
)

func obj(f map[string]jsondecode.Field) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Object, Fields: f}
}
func str() *jsondecode.Schema     { return &jsondecode.Schema{Kind: jsondecode.String} }
func boolean() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.Boolean} }
func num(min, max int) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Integer, Min: strconv.Itoa(min), Max: strconv.Itoa(max)}
}
func req(s *jsondecode.Schema) jsondecode.Field { return jsondecode.Field{Schema: s, Required: true} }
func opt(s *jsondecode.Schema) jsondecode.Field { return jsondecode.Field{Schema: s} }
func arr(s *jsondecode.Schema, max int) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Array, Element: s, MaxCount: max}
}

var testSchema = obj(map[string]jsondecode.Field{"id": req(str()), "version": req(str()), "result_version": req(str())})
var accessSchema = obj(map[string]jsondecode.Field{"network_type": req(str()), "vpn_state": req(str()), "ipv4_available": req(str()), "ipv6_available": req(str()), "validated": req(str()), "network_changed": req(boolean())})
var contextSchema = obj(map[string]jsondecode.Field{"access": req(accessSchema), "egress": req(obj(map[string]jsondecode.Field{
	"status": req(str()), "lookup_source": req(str()), "network_path": req(str()), "connection_code": opt(str()), "country_code": opt(str()), "region_code": opt(str()), "asn": opt(num(1, 4294967295)), "classification_version": opt(str()),
}))})

// The framing pass establishes finite representation bounds. DecodeProtocol
// subsequently selects and closes the exact observation/details type atomically.
var observationSchema = obj(map[string]jsondecode.Field{
	"type": req(str()), "status": req(str()), "failure": opt(str()), "duration_ms": req(num(0, 60000)),
	"details": req(obj(map[string]jsondecode.Field{
		"address_family": opt(str()), "engine": opt(str()), "query_type": opt(str()), "response_limit_reached": opt(boolean()),
		"responses":   opt(arr(obj(map[string]jsondecode.Field{"arrival_index": req(num(1, 3)), "arrival_offset_ms": req(num(0, 60000)), "rcode": req(str()), "answers": req(arr(obj(map[string]jsondecode.Field{"type": req(str()), "value": req(str())}), 16))}), 3)),
		"tls_version": opt(str()), "alpn": opt(str()), "certificate_validation": opt(str()), "alert": opt(str()),
		"protocol": opt(str()), "status_code": opt(num(100, 599)), "content_encoding": opt(str()),
		"redirect": opt(obj(map[string]jsondecode.Field{"followed_count": req(num(0, 2)), "terminal_location_class": req(str())})),
		"body":     opt(obj(map[string]jsondecode.Field{"bytes_processed": req(num(0, 65536)), "truncated": req(boolean())})),
		"checks":   opt(arr(obj(map[string]jsondecode.Field{"kind": req(str()), "index": req(num(0, 7)), "matched": req(boolean())}), 17)),
	})),
})

func measurementSchema(id string) *jsondecode.Schema {
	f := map[string]jsondecode.Field{"outcome": req(str()), "observations": req(arr(observationSchema, 16))}
	switch id {
	case "NET-01":
		f["access"] = opt(accessSchema)
	case "NET-03":
		f["platform_validation"] = req(str())
		f["response_class"] = opt(str())
	case "NET-05":
		f["profile"] = req(obj(map[string]jsondecode.Field{"profile_id": req(str()), "profile_version": req(&jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "9223372036854775807"})}))
		f["network_validation"] = req(obj(map[string]jsondecode.Field{"state": req(str()), "source": req(str())}))
		// Step payload schemas are selected from the trusted ordered plan in a second
		// bounded pass, before any model escapes. This union is closed per primitive.
		f["steps"] = req(arr(obj(map[string]jsondecode.Field{"step_id": req(str()), "test": req(testSchema), "state": req(str()), "measurement": opt(measurementSchema("primitive"))}), 8))
		f["stopped_before_step"] = opt(str())
	}
	return obj(f)
}
func envelope(id string) *jsondecode.Schema {
	return obj(map[string]jsondecode.Field{
		"schema_version": req(str()), "task_id": req(str()), "run_id": req(str()), "attempt": req(num(1, 2)), "test": req(testSchema),
		"implementation":  req(obj(map[string]jsondecode.Field{"library_version": req(str()), "platform": req(str())})),
		"time":            req(obj(map[string]jsondecode.Field{"started_at": req(str()), "precision": req(str()), "duration_ms": req(num(0, 60000)), "duration_capped": opt(boolean())})),
		"execution":       req(obj(map[string]jsondecode.Field{"state": req(str()), "reason": opt(str())})),
		"network_context": opt(contextSchema), "measurement": opt(measurementSchema(id)),
	})
}
