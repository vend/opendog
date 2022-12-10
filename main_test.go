package main

import (
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestDeserialiseMsgPackTrace(t *testing.T) {
	// Setup
	testSpan := ddspan{
		Service:  "clowncar.consumer",
		Name:     "kafka.consume",
		Resource: "Consume Topic kumamon-pics",
		TraceID:  2621704535220764017,
		SpanID:   5480210989507178917,
		ParentID: 2621704535220764017,
		Start:    1670560555960703334,
		Duration: 1025388084,
		Error:    0,
		Meta: map[string]string{
			"system.pid": "1",
			"runtime.id": "da557be3-da3a-4f3d-8aa5-86c4339fcfbc",
			"env":        "test",
			"_dd.p.dm":   "-3",
		},
		Metrics: map[string]float64{
			"_sampling_priority_v1": 2,
			"partition":             0,
			"offset":                57,
			"_dd.top_level":         1,
		},
		Type: "synthetics",
	}
	traceBundle := ddtraces{ddtrace{testSpan}}
	payload, err := msgpack.Marshal(traceBundle)
	if err != nil {
		t.Fatal(err)
	}

	// Actual test
	unpackedTraces, err := deserialiseMsgpackTraces(payload)
	if err != nil {
		t.Fatalf("Received an error: %v", err)
	}
	if len(unpackedTraces) == 0 {
		t.Fatal("Received 0 traces in deserialised trace bundle")
	}
}
