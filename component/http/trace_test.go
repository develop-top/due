package http

import (
	"bufio"
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestTraceResponse_SkipsSSEContentType(t *testing.T) {
	var resp fasthttp.Response
	resp.Header.SetContentType("text/event-stream")
	resp.SetBodyString("data: hello\n\n")

	if got := traceResponse(&resp); got != "<streaming response omitted>" {
		t.Fatalf("unexpected trace response: %q", got)
	}
}

func TestTraceResponse_SkipsStreamBody(t *testing.T) {
	var resp fasthttp.Response
	resp.SetBodyStreamWriter(func(w *bufio.Writer) {
		_, _ = w.WriteString("data: hello\n\n")
		_ = w.Flush()
	})

	if got := traceResponse(&resp); got != "<streaming response omitted>" {
		t.Fatalf("unexpected trace response: %q", got)
	}
}

func TestTraceResponse_FormatsRegularResponse(t *testing.T) {
	var resp fasthttp.Response
	resp.SetStatusCode(fasthttp.StatusAccepted)
	resp.Header.SetContentType("application/json")
	resp.SetBodyString(`{"ok":true}`)

	result := traceResponse(&resp)
	if !strings.Contains(result, "HTTP/1.1 202 Accepted") {
		t.Fatalf("missing status line in %q", result)
	}
	if !strings.Contains(result, "Content-Type: application/json") {
		t.Fatalf("missing content type in %q", result)
	}
	if !strings.Contains(result, `{"ok":true}`) {
		t.Fatalf("missing body in %q", result)
	}
}
