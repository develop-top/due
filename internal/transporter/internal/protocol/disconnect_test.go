package protocol_test

import (
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/session"
	"testing"
)

func TestEncodeDisconnectReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDisconnectReq(session.User, 3, true))

	t.Log(buffer.Bytes())
}

func TestDecodeDisconnectReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDisconnectReq(session.User, 3, false))

	seq, kind, target, force, err := protocol.DecodeDisconnectReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("kind: %v", kind)
	t.Logf("target: %v", target)
	t.Logf("force: %v", force)
}

func TestEncodeDisconnectRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDisconnectRes(codes.OK))

	t.Log(buffer.Bytes())
}

func TestDecodeDisconnectRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDisconnectRes(codes.InternalError))

	code, err := protocol.DecodeDisconnectRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
}
