package protocol_test

import (
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/internal/transporter/internal/route"
	"testing"
)

func TestDecodeGetStateReq(t *testing.T) {
	buf := protocol.EncodeBuffer(0, route.GetState, 1, nil, nil)

	seq, err := protocol.DecodeSeq(buffer.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
}

func TestDecodeGetStateRes(t *testing.T) {
	buf := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeGetStateRes(codes.OK, cluster.Work))

	code, state, err := protocol.DecodeGetStateRes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
	t.Logf("state: %v", state)
}

func TestDecodeSetStateReq(t *testing.T) {
	buf := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeSetStateReq(cluster.Shut))

	seq, state, err := protocol.DecodeSetStateReq(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("state: %v", state)
}

func TestDecodeSetStateRes(t *testing.T) {
	buf := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeSetStateRes(codes.OK))

	code, err := protocol.DecodeSetStateRes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
}
