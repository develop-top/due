package protocol_test

import (
	"bytes"
	"context"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	route2 "github.com/develop-top/due/v2/internal/transporter/internal/route"
	"github.com/develop-top/due/v2/tracer"
	"go.opentelemetry.io/otel/trace"
	"slices"
	"testing"
)

func init() {
	tracer.StartAgent(&tracer.Options{
		TraceName:      "due",
		Name:           "trace_test",
		Endpoint:       "",
		Sampler:        0,
		Batcher:        "",
		OtlpHeaders:    nil,
		OtlpHttpPath:   "",
		OtlpHttpSecure: false,
		Disabled:       false,
	})
}

func TestUnmarshalSpanContext(t *testing.T) {
	var b = protocol.MarshalSpanContext(trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{2},
		TraceFlags: 1,
		TraceState: trace.TraceState{},
		Remote:     false,
	}))
	sc := protocol.UnmarshalSpanContext(b)
	if sc.TraceID() != [16]byte{1} {
		t.Fail()
	}
	if sc.SpanID() != [8]byte{2} {
		t.Fail()
	}
	if sc.TraceFlags() != trace.TraceFlags(1) {
		t.Fail()
	}
}

func TestUnmarshalSpanContext2(t *testing.T) {
	var b = protocol.MarshalSpanContext(trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{},
		SpanID:     trace.SpanID{},
		TraceFlags: 0,
		TraceState: trace.TraceState{},
		Remote:     false,
	}))
	sc := protocol.UnmarshalSpanContext(b)
	if sc.TraceID() != [16]byte{} {
		t.Fail()
	}
	if sc.SpanID() != [8]byte{} {
		t.Fail()
	}
	if sc.TraceFlags() != trace.TraceFlags(0) {
		t.Fail()
	}
}

func TestReadTraceMessage(t *testing.T) {
	ctx := context.Background()
	ctx, _ = tracer.NewSpan(ctx, "test")

	traceBytes := protocol.MarshalSpanContext(trace.SpanContextFromContext(ctx))

	buf := protocol.EncodeBuffer(0, route2.Deliver, 1, traceBytes, protocol.EncodeDeliverReq(2, 3, []byte("hello")))
	t.Logf("buf:%v", buf.Bytes())

	isHeartbeat, route, seq, data, traceCtx, err := protocol.ReadTraceMessage(bytes.NewBuffer(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if isHeartbeat {
		t.Fatal("isHeartbeat")
	}
	if route != route2.Deliver {
		t.Fatal(route)
	}
	if seq != 1 {
		t.Fatal(seq)
	}

	seq, cid, uid, message, err := protocol.DecodeDeliverReq(data)
	if err != nil {
		t.Fatal(err)
	}
	if seq != 1 {
		t.Fatal(seq)
	}
	if cid != 2 {
		t.Fatal(cid)
	}
	if uid != 3 {
		t.Fatal(uid)
	}
	if string(message) != "hello" {
		t.Fatal(string(message))
	}

	t.Log(traceCtx)
	sc := protocol.UnmarshalSpanContext(traceCtx)
	t.Log(sc)

	traceID := sc.TraceID()
	traceIDEmpty := [16]byte{}
	if slices.Equal(traceID[:], traceIDEmpty[:]) {
		t.Fatal(traceID)
	}

	spanID := sc.SpanID()
	spanIDEmpty := [8]byte{}
	if slices.Equal(spanID[:], spanIDEmpty[:]) {
		t.Fatal(spanID)
	}

	if !slices.Equal([]byte{byte(sc.TraceFlags())}, []byte{0}) {
		t.Fatal(byte(sc.TraceFlags()))
	}
}
