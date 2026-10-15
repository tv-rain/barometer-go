package taskdecode

import (
	"strconv"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
)

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

// requirement-ref: REQ-TASK-0001
// Every fixed object is closed. Input/options are chosen from the exact tuple.
func envelope(d Definition) *jsondecode.Schema {
	text := &jsondecode.Schema{Kind: jsondecode.String}
	boolean := &jsondecode.Schema{Kind: jsondecode.Boolean}
	nullableTime := &jsondecode.Schema{Kind: jsondecode.String, Nullable: true}
	limits := object(map[string]jsondecode.Field{
		"max_duration_ms":    field(integer(1, 60000), false),
		"max_bytes_sent":     field(integer(0, 65536), false),
		"max_bytes_received": field(integer(0, 262144), false),
		"max_connections":    field(integer(0, 12), false),
	})
	backoff := object(map[string]jsondecode.Field{
		"mode": field(text, true), "initial_delay_ms": field(integer(30000, 300000), true),
		"max_delay_ms": field(integer(30000, 600000), true), "jitter_ms": field(integer(0, 60000), false),
		"multiplier": field(integer(2, 2), false),
	})
	schedule := object(map[string]jsondecode.Field{
		"starts_at": field(nullableTime, false), "not_after": field(text, false),
		"time_window": field(object(map[string]jsondecode.Field{
			"timezone": field(text, false),
			"ranges":   field(array(object(map[string]jsondecode.Field{"start": field(text, true), "end": field(text, true)}), 4), true),
		}), false),
		"repeat": field(object(map[string]jsondecode.Field{
			"mode": field(text, false), "interval": field(integer(21600, 604800), false), "jitter": field(integer(300, 1800), false),
		}), false),
		"retry": field(object(map[string]jsondecode.Field{
			"max_attempts": field(integer(1, 2), true), "retry_on": field(array(text, 4), false), "backoff": field(backoff, false),
		}), false),
		"limits": field(object(map[string]jsondecode.Field{"max_runs": field(integer(1, 28), false)}), false),
	})
	execution := object(map[string]jsondecode.Field{
		"network_path": field(text, false), "limits": field(limits, true),
		"network": field(object(map[string]jsondecode.Field{
			"allowed_transports": field(array(text, 0), false), "vpn_policy": field(text, false), "require_validated": field(boolean, false),
			"context": field(object(map[string]jsondecode.Field{
				"required_fields": field(array(text, 3), false), "allowed_country_codes": field(array(text, 249), false),
			}), false),
		}), false),
		"device": field(object(map[string]jsondecode.Field{
			"minimum_battery_percent": field(integer(0, 100), false), "allow_charging_only": field(boolean, false),
		}), false),
		"concurrency": field(object(map[string]jsondecode.Field{"mode": field(text, false)}), false),
	})
	input, options := *d.Input, *d.Options
	// Registered generic object caps apply even if a test descriptor omits them.
	if input.MaxBytes == 0 || input.MaxBytes > 49152 {
		input.MaxBytes = 49152
	}
	if options.MaxBytes == 0 || options.MaxBytes > 8192 {
		options.MaxBytes = 8192
	}
	return object(map[string]jsondecode.Field{
		"schema_version": field(text, true), "task_id": field(text, true), "test_id": field(text, true), "test_version": field(text, true),
		"input": field(&input, true), "options": field(&options, true), "schedule": field(schedule, true), "execution": field(execution, true),
		"safety": field(object(map[string]jsondecode.Field{
			"risk_level": field(text, true), "requires_user_consent": field(boolean, true),
			"risk_factors": field(array(object(map[string]jsondecode.Field{"type": field(text, true), "level": field(text, true)}), 0), false),
		}), true),
		"labels": field(&jsondecode.Schema{Kind: jsondecode.Object, Additional: text, MaxBytes: 2048, MaxCount: 16}, false),
	})
}
