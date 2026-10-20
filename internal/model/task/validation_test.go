package task_test

import (
	"reflect"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Trusted models deliberately have no JSON round trip. Reconstruction through
// copied typed views must preserve all restrictions without adding authority.
func TestTrustedModelReconstruction(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		for _, in := range []task.SourceInput{source(), {Origin: task.Embedded, IssuerID: "release", PolicyGeneration: 1}} {
			first, err := task.NewSourceContext(in)
			if err != nil {
				t.Fatal(err)
			}
			next := task.SourceInput{Origin: first.Origin(), IssuerID: first.IssuerID(), PolicyGeneration: first.PolicyGeneration()}
			second, err := task.NewSourceContext(next)
			if err != nil || !reflect.DeepEqual(first, second) {
				t.Fatal("source changed on reconstruction", err)
			}
		}
		pending, _ := task.NewSourceContext(source())
		first, err := pending.WithVerifiedPackage("reviewed-issuer", whole())
		if err != nil {
			t.Fatal(err)
		}
		issued, _ := first.IssuedAt()
		again, _ := task.NewSourceContext(task.SourceInput{Origin: first.Origin(), PolicyGeneration: first.PolicyGeneration()})
		second, err := again.WithVerifiedPackage(first.IssuerID(), issued)
		if err != nil || !reflect.DeepEqual(first, second) {
			t.Fatal("verified remote source changed on reconstruction", err)
		}
	})
	t.Run("policy", func(t *testing.T) {
		in := policy()
		first, err := task.NewLocalPolicy(in)
		if err != nil {
			t.Fatal(err)
		}
		next := task.PolicyInput{Generation: first.Generation(), Permissions: first.Permissions()}
		for _, original := range in.Issuers {
			view, ok := first.Issuer(original.IssuerID)
			if !ok {
				t.Fatal("missing issuer")
			}
			next.Issuers = append(next.Issuers, view)
		}
		for _, original := range in.Targets {
			view, ok := first.Target(original.ID)
			if !ok {
				t.Fatal("missing target")
			}
			next.Targets = append(next.Targets, view)
		}
		second, err := task.NewLocalPolicy(next)
		if err != nil || !reflect.DeepEqual(first, second) {
			t.Fatal("policy changed on reconstruction", err)
		}
		next.Permissions.Paths.Proxy = true
		if first.Permissions().Paths.Proxy || second.Permissions().Paths.Proxy {
			t.Fatal("policy authority expanded")
		}
	})
	t.Run("revocation", func(t *testing.T) {
		first, err := task.NewRevocations(task.RevocationInput{PolicyGeneration: 2, IssuerIDs: []string{"issuer"}, RegistrationIDs: []string{"endpoint"}})
		if err != nil {
			t.Fatal(err)
		}
		second, err := task.NewRevocations(first.Input())
		if err != nil || !reflect.DeepEqual(first, second) {
			t.Fatal("revocation changed on reconstruction", err)
		}
	})
}
