package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	_ "github.com/vmihailenco/msgpack/v5"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	collectorpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

func deserialiseMsgpackTraces(payload []byte) (ddtraces, error) {
	var bundle ddtraces
	if err := msgpack.Unmarshal(payload, &bundle); err != nil {
		return bundle, err
	}
	return bundle, nil
}

func main() {
	traceHandler := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		if len(body) == 1 {
			log.Printf("Received initial connection from %s on %s", r.UserAgent(), r.RemoteAddr)
			return
		}
		traces, err := deserialiseMsgpackTraces(body)
		err = msgpack.Unmarshal(body, &traces)
		if err != nil {
			log.Println("Failed to unpack traces")
		}
		if len(traces) > 0 && len(traces[0]) > 0 {
			resourceSpans := []*tracepb.ResourceSpans{}
			for _, trace := range traces {
				for _, span := range trace {
					pbspan := tracepb.Span{
						Name:              span.Resource,
						Kind:              tracepb.Span_SPAN_KIND_CLIENT,
						StartTimeUnixNano: uint64(time.Unix(0, span.Start).UnixNano()),
						EndTimeUnixNano:   uint64(time.Unix(0, span.Start+span.Duration).UnixNano()),
						Attributes:        []*v1.KeyValue{},
						Status: &tracepb.Status{
							Code: tracepb.Status_STATUS_CODE_UNSET,
						},
					}
					if span.TraceID != 0 {
						// Trace ID is 32 chars while Datadog's default is 19 so we just pad it out
						traceID, err := hex.DecodeString(fmt.Sprintf("%032d", span.TraceID))
						if err != nil {
							log.Print(err)
							log.Print("Failed to convert trace ID")
						}
						pbspan.TraceId = traceID
					}
					if span.SpanID != 0 {
						// Span ID is 16 chars while Datadog's default is ~19 so we just trim it down
						// In the event that it is shorter, we will pad it out to be 16 chars
						spanID, err := hex.DecodeString(fmt.Sprintf("%016d", span.SpanID)[:16])
						if err != nil {
							log.Print("Failed to convert spanID ID")
						}
						pbspan.SpanId = spanID
					}
					if span.ParentID != 0 {
						// Span ID is 16 chars while Datadog's default is ~19 so we just trim it down
						// In the event that it is shorter, we will pad it out to be 16 chars
						parentID, err := hex.DecodeString(fmt.Sprintf("%016d", span.ParentID)[:16])
						if err != nil {
							log.Print("Failed to convert parent ID")
						}
						pbspan.ParentSpanId = parentID
					}
					for k, v := range span.Meta {
						pbspan.Attributes = append(pbspan.Attributes, &v1.KeyValue{
							Key: k,
							Value: &v1.AnyValue{
								Value: &v1.AnyValue_StringValue{
									StringValue: v,
								},
							},
						})
					}
					for k, v := range span.Metrics {
						pbspan.Attributes = append(pbspan.Attributes, &v1.KeyValue{
							Key: k,
							Value: &v1.AnyValue{
								Value: &v1.AnyValue_DoubleValue{
									DoubleValue: v,
								},
							},
						})
					}
					pbspan.Attributes = append(pbspan.Attributes, &v1.KeyValue{
						Key: string(semconv.ServiceNameKey),
						Value: &v1.AnyValue{
							Value: &v1.AnyValue_StringValue{
								StringValue: span.Service,
							},
						},
					})
					scopeSpans := tracepb.ScopeSpans{
						Scope: &v1.InstrumentationScope{
							Name:    "opendog",
							Version: "1.0.0",
						},
						Spans:     []*tracepb.Span{&pbspan},
						SchemaUrl: semconv.SchemaURL,
					}
					resourceSpan := tracepb.ResourceSpans{
						Resource: &resourcepb.Resource{
							Attributes: []*v1.KeyValue{
								{
									Key: string(semconv.ServiceNameKey),
									Value: &v1.AnyValue{
										Value: &v1.AnyValue_StringValue{
											StringValue: span.Service,
										},
									},
								},
							},
						},
						ScopeSpans: []*tracepb.ScopeSpans{&scopeSpans},
						SchemaUrl:  semconv.SchemaURL,
					}
					resourceSpans = append(resourceSpans, &resourceSpan)
				}
			}
			request := collectorpb.ExportTraceServiceRequest{
				ResourceSpans: resourceSpans,
			}
			collectorRequest, err := proto.Marshal(&request)
			if err != nil {
				log.Fatal(err)
			}
			res, err := http.Post("http://localhost:4318/v1/traces", "application/x-protobuf", bytes.NewBuffer(collectorRequest))
			if err != nil {
				log.Fatal(err)
			}
			if res.StatusCode == 200 {
				log.Printf("Succesfully converted and forwarded %d spans", len(resourceSpans))
			} else {
				log.Printf("Failed to convert spans. Received status code %d", res.StatusCode)
			}
			fmt.Fprintf(w, "OK")
		}
	}

	http.HandleFunc("/", traceHandler)
	log.Print("Opendog is listening on 127.0.0.1:8126")
	log.Fatal(http.ListenAndServe(":8126", nil))
}
