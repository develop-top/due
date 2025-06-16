package protocol_test

import (
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"testing"
)

func TestEncodeDeliverReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDeliverReq(2, 3, []byte("hello world")))

	t.Log(buffer.Bytes())
}

func TestDecodeDeliverReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDeliverReq(2, 3, []byte("hello world")))

	seq, cid, uid, message, err := protocol.DecodeDeliverReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("cid: %v", cid)
	t.Logf("uid: %v", uid)
	t.Logf("message: %v", string(message))
}

func TestEncodeDeliverRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDeliverRes(codes.InternalError))

	t.Log(buffer.Bytes())
}

func TestDecodeDeliverRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeDeliverRes(codes.InternalError))

	code, err := protocol.DecodeDeliverRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
}
