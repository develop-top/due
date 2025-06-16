package protocol_test

import (
	"testing"

	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/utils/xuuid"
)

func TestEncodeHandshakeReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeHandshakeReq(cluster.Gate, xuuid.UUID()))
	t.Log(buffer.Bytes())
}

func TestDecodeHandshakeReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeHandshakeReq(cluster.Gate, xuuid.UUID()))
	seq, insKind, insID, err := protocol.DecodeHandshakeReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("seq: %v", seq)
	t.Logf("kind: %v", insKind)
	t.Logf("id: %v", insID)
}

func TestEncodeHandshakeRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeHandshakeRes(codes.OK))
	t.Log(buffer.Bytes())
}

func TestDecodeHandshakeRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeHandshakeRes(codes.InternalError))
	code, err := protocol.DecodeHandshakeRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("code: %v", code)
}
