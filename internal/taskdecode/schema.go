package taskdecode

import (
	"strconv"
	"strings"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

// Registered generic object caps (REQ-LIMIT-0004). They apply even if a test
// descriptor omits them.
const (
	maxInputBytes   = 49152
	maxOptionsBytes = 8192
)

// The two common envelopes share one shape. shapeSchema checks only the closed
// structure and value kinds (steps 3 and 4): integer ranges, counts and byte
// caps are relaxed and input and options are opaque objects. fieldSchema then
// applies every value range and registered limit (step 6), with input and
// options still opaque apart from their registered byte caps.
var (
	shapeSchema = envelope(false)
	fieldSchema = envelope(true)
)

// relaxed integer endpoints admit every canonical integer token the parser
// can represent within the task byte limit.
var relaxed = strings.Repeat("9", 4096)

// requirement-ref: REQ-TASK-0001
// Every fixed object is closed. Input/options are checked per test in Parameters.
func envelope(strict bool) *jsondecode.Schema {
	// bounded marks a registered maximum with limit; a plain range check is not.
	bounded := func(min, max int, limit bool) *jsondecode.Schema {
		if !strict {
			return &jsondecode.Schema{Kind: jsondecode.Integer, Min: "-" + relaxed, Max: relaxed, Limit: limit}
		}
		return &jsondecode.Schema{Kind: jsondecode.Integer, Min: strconv.Itoa(min), Max: strconv.Itoa(max), Limit: limit}
	}
	list := func(element *jsondecode.Schema, max int, limit bool) *jsondecode.Schema {
		if !strict {
			max = 0
		}
		return &jsondecode.Schema{Kind: jsondecode.Array, Element: element, MaxCount: max, Limit: limit}
	}
	opaque := func(maxBytes int) *jsondecode.Schema {
		if !strict {
			maxBytes = 0
		}
		return &jsondecode.Schema{Kind: jsondecode.Object, Additional: &jsondecode.Schema{Kind: jsondecode.Any}, MaxBytes: maxBytes, Limit: true}
	}
	text := &jsondecode.Schema{Kind: jsondecode.String}
	boolean := &jsondecode.Schema{Kind: jsondecode.Boolean}
	nullableTime := &jsondecode.Schema{Kind: jsondecode.String, Nullable: true}
	limits := object(map[string]jsondecode.Field{
		"max_duration_ms":    field(bounded(1, 60000, true), false),
		"max_bytes_sent":     field(bounded(0, 65536, true), false),
		"max_bytes_received": field(bounded(0, 262144, true), false),
		"max_connections":    field(bounded(0, 12, true), false),
	})
	backoff := object(map[string]jsondecode.Field{
		"mode": field(text, true), "initial_delay_ms": field(bounded(30000, 300000, true), true),
		"max_delay_ms": field(bounded(30000, 600000, true), true), "jitter_ms": field(bounded(0, 60000, true), false),
		"multiplier": field(bounded(2, 2, true), false),
	})
	schedule := object(map[string]jsondecode.Field{
		"starts_at": field(nullableTime, false), "not_after": field(text, false),
		"time_window": field(object(map[string]jsondecode.Field{
			"timezone": field(text, false),
			"ranges":   field(list(object(map[string]jsondecode.Field{"start": field(text, true), "end": field(text, true)}), 4, true), true),
		}), false),
		"repeat": field(object(map[string]jsondecode.Field{
			"mode": field(text, false), "interval": field(bounded(21600, 604800, true), false), "jitter": field(bounded(300, 1800, true), false),
		}), false),
		"retry": field(object(map[string]jsondecode.Field{
			"max_attempts": field(bounded(1, 2, true), true), "retry_on": field(list(text, 4, false), false), "backoff": field(backoff, false),
		}), false),
		"limits": field(object(map[string]jsondecode.Field{"max_runs": field(bounded(1, 28, true), false)}), false),
	})
	execution := object(map[string]jsondecode.Field{
		"network_path": field(text, false), "limits": field(limits, true),
		"network": field(object(map[string]jsondecode.Field{
			"allowed_transports": field(list(text, 0, false), false), "vpn_policy": field(text, false), "require_validated": field(boolean, false),
			"context": field(object(map[string]jsondecode.Field{
				"required_fields": field(list(text, 3, false), false), "allowed_country_codes": field(list(text, 249, false), false),
			}), false),
		}), false),
		"device": field(object(map[string]jsondecode.Field{
			"minimum_battery_percent": field(bounded(0, 100, false), false), "allow_charging_only": field(boolean, false),
		}), false),
		"concurrency": field(object(map[string]jsondecode.Field{"mode": field(text, false)}), false),
	})
	labelBytes, labelCount := 2048, 16
	if !strict {
		labelBytes, labelCount = 0, 0
	}
	return object(map[string]jsondecode.Field{
		"schema_version": field(text, true), "task_id": field(text, true), "test_id": field(text, true), "test_version": field(text, true),
		"input": field(opaque(maxInputBytes), true), "options": field(opaque(maxOptionsBytes), true), "schedule": field(schedule, true), "execution": field(execution, true),
		"safety": field(object(map[string]jsondecode.Field{
			"risk_level": field(text, true), "requires_user_consent": field(boolean, true),
			"risk_factors": field(list(object(map[string]jsondecode.Field{"type": field(text, true), "level": field(text, true)}), 0, false), false),
		}), true),
		// requirement-ref: REQ-LIMIT-0004
		"labels": field(&jsondecode.Schema{Kind: jsondecode.Object, Additional: text, MaxBytes: labelBytes, MaxCount: labelCount, Limit: true}, false),
	})
}

// inputSchema and optionsSchema are the selected test's descriptors with the
// registered byte caps applied when the descriptor omits or exceeds them.
func inputSchema(d Definition) *jsondecode.Schema   { return capped(*d.Input, maxInputBytes) }
func optionsSchema(d Definition) *jsondecode.Schema { return capped(*d.Options, maxOptionsBytes) }

func capped(s jsondecode.Schema, max int) *jsondecode.Schema {
	if s.MaxBytes == 0 || s.MaxBytes > max {
		s.MaxBytes, s.Limit = max, true
	}
	return &s
}

func object(fields map[string]jsondecode.Field) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Object, Fields: fields}
}
func field(s *jsondecode.Schema, required bool) jsondecode.Field {
	return jsondecode.Field{Schema: s, Required: required}
}
func integer(min, max int) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Integer, Min: strconv.Itoa(min), Max: strconv.Itoa(max)}
}
func array(element *jsondecode.Schema, max int) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Array, Element: element, MaxCount: max}
}
