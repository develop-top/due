package protocol_test

import (
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/session"
	"testing"
)

func TestDecodeStatReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeStatReq(session.User))

	seq, kind, err := protocol.DecodeStatReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("kind: %v", kind)
}

func TestDecodeStatRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeStatRes(codes.OK, 3000))

	code, total, err := protocol.DecodeStatRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
	t.Logf("total: %v", total)
}
