package main

/// Datadog

// spanContext contains a bunch of state that may or may not
// propagate across child spans within a single trace
// See https://github.com/DataDog/dd-trace-go/blob/main/ddtrace/tracer/spancontext.go#L25
type spanContext struct {
	trace  *ddtrace
	span   *ddspan
	errors int32

	traceID uint64
	spanID  uint64

	baggage    map[string]string
	hasBaggage uint32
	origin     string
}

// span represents information about a single HTTP request.
// It reflects a single unit of work and metadata about that
// work unit such as how long it took, whether it succeeded
// and other useful metrics and tags to report.
// See https://github.com/DataDog/dd-trace-go/blob/main/ddtrace/tracer/span.go#L62
type ddspan struct {
	Name     string             `msgpack:"name"`
	Service  string             `msgpack:"service"`
	Resource string             `msgpack:"resource"`
	Type     string             `msgpack:"type"`
	Start    int64              `msgpack:"start"`
	Duration int64              `msgpack:"duration"`
	Meta     map[string]string  `msgpack:"meta,omitempty"`
	Metrics  map[string]float64 `msgpack:"metrics,omitempty"`
	SpanID   uint64             `msgpack:"span_id"`
	TraceID  uint64             `msgpack:"trace_id"`
	ParentID uint64             `msgpack:"parent_id"`
	Error    int32              `msgpack:"error"`

	Context *spanContext `msgpack:"-"`
}

// trace is a group of spans that share the same trace ID.
// They may reflect multiple units of work within a single
// service or units of work across multiple services
// See https://github.com/DataDog/dd-trace-go/blob/main/ddtrace/tracer/tracer.go#L332
type ddtrace []ddspan

// traces are what we receive from the APM agents. Each
// APM agent will bundle up multiple traces together
// and send them off to an instance of the Datadog agent.
// This happens around every 10 seconds or so.
// Opendog does not deal with more than one trace at a time
// but msgpack traces come in a bundle we still deserialise
type ddtraces []ddtrace
