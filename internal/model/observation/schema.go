package observation

import (
	"strconv"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

func object(fields map[string]jsondecode.Field) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Object, Fields: fields}
}
func textSchema() *jsondecode.Schema { return &jsondecode.Schema{Kind: jsondecode.String} }
func boolean() *jsondecode.Schema    { return &jsondecode.Schema{Kind: jsondecode.Boolean} }
func integer(max string) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Integer, Min: "0", Max: max}
}
func required(s *jsondecode.Schema) jsondecode.Field {
	return jsondecode.Field{Schema: s, Required: true}
}
func optional(s *jsondecode.Schema) jsondecode.Field { return jsondecode.Field{Schema: s} }
func number(v *jsondecode.Value, name string) int    { n, _ := strconv.Atoi(field(v, name)); return n }

var tcpSchema = object(map[string]jsondecode.Field{"address_family": required(textSchema())})
var httpSchema = object(map[string]jsondecode.Field{
	"address_family": required(textSchema()), "protocol": required(textSchema()),
	"status_code":      optional(&jsondecode.Schema{Kind: jsondecode.Integer, Min: "100", Max: "599"}),
	"content_encoding": optional(textSchema()),
	"redirect":         optional(object(map[string]jsondecode.Field{"followed_count": required(integer("2")), "terminal_location_class": required(textSchema())})),
	"body":             optional(object(map[string]jsondecode.Field{"bytes_processed": required(integer("65536")), "truncated": required(boolean())})),
	"checks":           optional(&jsondecode.Schema{Kind: jsondecode.Array, MaxCount: 17, Element: object(map[string]jsondecode.Field{"kind": required(textSchema()), "index": required(integer("7")), "matched": required(boolean())})}),
})

func envelope(details *jsondecode.Schema) *jsondecode.Schema {
	return object(map[string]jsondecode.Field{
		"type": required(textSchema()), "status": required(textSchema()), "duration_ms": required(integer("60000")), "failure": optional(textSchema()), "details": required(details),
	})
}

var protocolSelection = jsondecode.Selection{Fields: []string{"type"}, Cases: []jsondecode.Case{
	{Values: []string{"tcp_connect"}, Schema: envelope(tcpSchema)},
	{Values: []string{"http_exchange"}, Schema: envelope(httpSchema)},
	{Values: []string{"dns_query"}, Schema: envelope(dnsSchema)},
	{Values: []string{"tls_handshake"}, Schema: envelope(tlsSchema)},
}}

var dnsSchema = object(map[string]jsondecode.Field{
	"engine": required(textSchema()), "address_family": optional(textSchema()), "query_type": required(textSchema()), "response_limit_reached": optional(boolean()),
	"responses": required(&jsondecode.Schema{Kind: jsondecode.Array, MaxCount: 3, Element: object(map[string]jsondecode.Field{
		"arrival_index": required(&jsondecode.Schema{Kind: jsondecode.Integer, Min: "1", Max: "3"}), "arrival_offset_ms": required(integer("60000")), "rcode": required(textSchema()),
		"answers": required(&jsondecode.Schema{Kind: jsondecode.Array, MaxCount: 16, Element: object(map[string]jsondecode.Field{"type": required(textSchema()), "value": required(textSchema())})}),
	})}),
})
var tlsSchema = object(map[string]jsondecode.Field{
	"address_family": required(textSchema()), "tls_version": optional(textSchema()), "alpn": optional(textSchema()), "certificate_validation": required(textSchema()), "alert": optional(textSchema()),
})
