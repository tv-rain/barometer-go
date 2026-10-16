package bundle_test

import (
	"errors"
	"testing"

	"github.com/tv-rain/barometer-go/internal/bundle"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/result"
)

// requirement-ref: REQ-JSON-0010
// requirement-ref: REQ-JSON-0011
// requirement-ref: REQ-BUNDLE-0003
func TestBundleValidationPrecedence(t *testing.T) {
	b := builder(t, true)
	invalidLink := &model.RotatingLinkage{Scope: "unapproved", Epoch: -1}
	for _, tc := range []struct {
		name    string
		builder *bundle.Builder
		items   []result.Value
		link    *model.RotatingLinkage
		want    error
	}{
		{"binding before count", &bundle.Builder{}, nil, invalidLink, bundle.ErrBinding},
		{"count before linkage", b, make([]result.Value, 17), invalidLink, jsondecode.ErrBounds},
		{"linkage before member", b, []result.Value{{}}, invalidLink, bundle.ErrValue},
		{"invalid member rejects all", b, []result.Value{notRun(t), {}}, nil, bundle.ErrValue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for range 20 {
				v, err := tc.builder.Build(tc.items, tc.link)
				rejected(t, v, err)
				if !errors.Is(err, tc.want) || len(err.Error()) > 128 {
					t.Fatalf("precedence: %v", err)
				}
			}
		})
	}
}
