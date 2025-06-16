package protocol_test

import (
	"testing"

	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
)

func TestEncodeBindReq(t *testing.T) {
	buffer := protocol.EncodeBindReq(2, 3)

	t.Log(buffer.Bytes())
}

func TestDecodeBindReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeBindReq(2, 3))

	seq, cid, uid, err := protocol.DecodeBindReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("cid: %v", cid)
	t.Logf("uid: %v", uid)
}

func TestEncodeBindRes(t *testing.T) {
	buffer := protocol.EncodeBindRes(codes.OK)

	t.Log(buffer.Bytes())
}

func TestDecodeBindRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeBindRes(codes.OK))

	code, err := protocol.DecodeBindRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
}
