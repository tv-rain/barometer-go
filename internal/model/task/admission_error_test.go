package task_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// REQ-TRUST-0033..0035: finite, payload-free operational errors, separate
// from measurement results and execution reasons.
func TestAdmissionErrorVocabulary(t *testing.T) {
	for _, tc := range []struct {
		code task.AdmissionCode
		text string
	}{
		{task.SourceUnauthenticated, "task_source_unauthenticated"},
		{task.IssuerNotAuthorized, "task_issuer_not_authorized"},
		{task.PolicyUnavailable, "task_policy_unavailable"},
		{task.TrustUnavailable, "task_trust_unavailable"},
		{task.Expired, "task_expired"}, {task.IDConflict, "task_id_conflict"},
		{task.PolicyRollback, "policy_rollback"}, {task.TargetNotRegistered, "target_not_registered"},
		{task.LimitExceeded, "task_limit_exceeded"}, {task.SchemaInvalid, "task_schema_invalid"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			e, err := task.NewAdmissionError(tc.code)
			if err != nil || e.Code() != tc.code || e.Error() != tc.text {
				t.Fatal("wrong operational code")
			}
			if task.ExecutionReason(e.Error()).Valid() {
				t.Fatal("admission error became execution reason")
			}
			if b, err := json.Marshal(e); !errors.Is(err, task.ErrAdmissionEncoding) || len(b) != 0 {
				t.Fatal("error serialized")
			}
			before := e
			canary := "secret credentials participant target.example raw-platform-message"
			if err := json.Unmarshal([]byte(`{"message":"`+canary+`"}`), &e); !errors.Is(err, task.ErrAdmissionEncoding) || e != before {
				t.Fatal("error imported arbitrary data")
			}
			for _, format := range []string{"%s", "%v", "%+v", "%#v"} {
				if strings.Contains(fmt.Sprintf(format, e), canary) {
					t.Fatal("sensitive input retained")
				}
			}
		})
	}
	// The entire uint8 domain is checked, including the invalid zero value.
	for n := 0; n < 256; n++ {
		e, err := task.NewAdmissionError(task.AdmissionCode(n))
		if n >= 1 && n <= 10 {
			continue
		}
		if !errors.Is(err, task.ErrAdmissionCode) || e.Code() != 0 || e.Error() != "task: invalid admission error" {
			t.Fatal("invalid code accepted")
		}
	}
}
