package protocol_test

import (
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/internal/transporter/internal/route"
	"github.com/develop-top/due/v2/session"
	"testing"
)

func TestDecodeIsOnlineReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(protocol.DataBit, route.IsOnline, 1, nil, protocol.EncodeIsOnlineReq(session.User, 1))

	seq, kind, target, err := protocol.DecodeIsOnlineReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("kind: %v", kind)
	t.Logf("target: %v", target)
}

func TestDecodeIsOnlineRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(protocol.DataBit, route.IsOnline, 1, nil, protocol.EncodeIsOnlineRes(codes.NotFoundSession, false))

	code, isOnline, err := protocol.DecodeIsOnlineRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
	t.Logf("isOnline: %v", isOnline)
}
