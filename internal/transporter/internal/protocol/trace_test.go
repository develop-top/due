package protocol_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/tracer"
	"github.com/develop-top/due/v2/utils/xtrace"
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

func TestEncodeTraceBuffer(t *testing.T) {
	buf := protocol.EncodeTraceBuffer(context.Background(), buffer.NewNocopyBuffer([]byte("hello")))
	if binary.BigEndian.Uint32(buf.Bytes()) != 25+5 {
		t.Fail()
	}
	for _, v := range buf.Bytes()[4:29] {
		if v != uint8(0) {
			t.Fail()
		}
	}
	if string(buf.Bytes()[29:]) != "hello" {
		t.Fail()
	}
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
	ctx, _ = xtrace.StartRPCClientSpan(ctx, "test")

	b := protocol.EncodeTraceBuffer(ctx, protocol.EncodeDeliverReq(1, 2, 3, []byte("hello")))

	isHeartbeat, route, seq, data, traceCtx, err := protocol.ReadTraceMessage(bytes.NewBuffer(b.Bytes()))
	if err != nil {
		t.Fatal()
	}
	if isHeartbeat {
		t.Fatal()
	}
	if route != 12 {
		t.Fatal()
	}
	if seq != 1 {
		t.Fatal()
	}

	seq, cid, uid, message, err := protocol.DecodeDeliverReq(data)
	if err != nil {
		t.Fatal()
	}
	if seq != 1 {
		t.Fatal()
	}
	if cid != 2 {
		t.Fatal()
	}
	if uid != 3 {
		t.Fatal()
	}
	if string(message) != "hello" {
		t.Fatal()
	}

	t.Log(traceCtx)
	sc := protocol.UnmarshalSpanContext(traceCtx)
	t.Log(sc)

	traceID := sc.TraceID()
	traceIDEmpty := [16]byte{}
	if slices.Equal(traceID[:], traceIDEmpty[:]) {
		t.Fatal()
	}

	spanID := sc.SpanID()
	spanIDEmpty := [8]byte{}
	if slices.Equal(spanID[:], spanIDEmpty[:]) {
		t.Fatal()
	}

	if !slices.Equal([]byte{byte(sc.TraceFlags())}, []byte{0}) {
		t.Fatal()
	}
}
