package testkit_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/tv-rain/barometer-go/internal/testkit"
)

// This deliberately small synthetic model is not a Barometer schema or decoder.
// It demonstrates consumer-owned classification and positive controls using the
// same runner that the protocol example below uses.
func TestReuseForSyntheticModel(t *testing.T) {
	invalid := errors.New("synthetic model rejection")
	decode := func(input []byte) error {
		var object map[string]json.RawMessage
		if json.Unmarshal(input, &object) != nil || len(object) != 1 || string(object["version"]) != `"1"` {
			return invalid
		}
		return nil
	}
	if err := decode([]byte(`{"version":"1"}`)); err != nil {
		t.Fatal("positive control failed")
	}
	testkit.RunRejections(t, []testkit.RejectionCase{
		{Name: "unknown-field", Requirement: "synthetic model closed schema", Expected: "model", Input: []byte(`{"version":"1","extra":0}`)},
		{Name: "unsupported-version", Requirement: "synthetic model exact version", Expected: "model", Input: []byte(`{"version":"2"}`)},
		{Name: "wrong-type", Requirement: "synthetic model string version", Expected: "model", Input: []byte(`{"version":1}`)},
	}, decode, func(class string, err error) bool { return class == "model" && errors.Is(err, invalid) })
}

func TestReuseForSyntheticProtocol(t *testing.T) {
	invalid := errors.New("synthetic protocol rejection")
	decode := func(input []byte) error {
		if len(input) != 2 || input[0] != 1 {
			return invalid
		}
		return nil
	}
	if err := decode([]byte{1, 0}); err != nil {
		t.Fatal("positive control failed")
	}
	boundary, err := testkit.ByteBoundaries(2, 1)
	if err != nil {
		t.Fatal(err)
	}
	testkit.RunRejections(t, []testkit.RejectionCase{
		{Name: "truncated", Requirement: "synthetic protocol frame length", Expected: "frame", Input: boundary.Below},
		{Name: "oversized", Requirement: "synthetic protocol frame length", Expected: "frame", Input: boundary.Above},
		{Name: "unsupported-version", Requirement: "synthetic protocol version", Expected: "frame", Input: []byte{2, 0}},
	}, decode, func(class string, err error) bool { return class == "frame" && errors.Is(err, invalid) })
}

func TestPaddedJSONIsExactAndIndependent(t *testing.T) {
	value := []byte(`{"x":0}`)
	for _, size := range []int{len(value), 65535, 65536, 65537} {
		fixture, err := testkit.PaddedJSON(value, size)
		if err != nil || len(fixture) != size || !json.Valid(fixture) {
			t.Fatal("padding changed syntax or size")
		}
		fixture[0] = 'x'
		if value[0] != '{' {
			t.Fatal("padding aliases source")
		}
	}
	for _, size := range []int{-1, len(value) - 1, testkit.MaxFixtureBytes + 1} {
		if _, err := testkit.PaddedJSON(value, size); !errors.Is(err, testkit.ErrFixtureBounds) {
			t.Fatal("invalid padding size accepted")
		}
	}
}

func TestTransportCancellationAfterResponse(t *testing.T) {
	transport, _ := testkit.NewTransport(testkit.ResponseStep{StatusCode: 200, Body: []byte("synthetic")})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response, err := transport.RoundTrip(request(t, ctx))
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := io.ReadAll(response.Body); !errors.Is(err, context.Canceled) {
		t.Fatal("body read ignored cancellation")
	}
	// io.Copy must not bypass cancellation through a promoted WriterTo method.
	if n, err := io.Copy(io.Discard, response.Body); n != 0 || !errors.Is(err, context.Canceled) {
		t.Fatal("copy bypassed cancellation")
	}
	_ = response.Body.Close()
	if n, err := io.Copy(io.Discard, response.Body); n != 0 || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("copy bypassed close")
	}
	if _, open := transport.Counts(); open != 0 {
		t.Fatal("cancelled body leaked")
	}
	wait, _ := testkit.NewTransport(testkit.ResponseStep{WaitForCancel: true})
	if _, err := wait.RoundTrip(request(t, context.Background())); err == nil {
		t.Fatal("uncancellable wait accepted")
	}
}
