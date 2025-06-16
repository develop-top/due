package protocol_test

import (
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/packet"
	"github.com/develop-top/due/v2/session"
	"testing"
)

func TestEncodeMulticastReq(t *testing.T) {
	message, err := packet.PackMessage(&packet.Message{
		Route:  1,
		Seq:    2,
		Buffer: []byte("hello world"),
	})
	if err != nil {
		t.Fatal(err)
	}

	buf := protocol.EncodeBuffer(protocol.DataBit, 0, 1, nil, protocol.EncodeMulticastReq(session.User, []int64{1, 2, 3}, buffer.NewNocopyBuffer(message)))

	t.Log(buf.Bytes())
}

func TestDecodeMulticastReq(t *testing.T) {
	message, err := packet.PackMessage(&packet.Message{
		Route:  1,
		Seq:    2,
		Buffer: []byte("hello world"),
	})
	if err != nil {
		t.Fatal(err)
	}

	buf := protocol.EncodeBuffer(protocol.DataBit, 0, 1, nil, protocol.EncodeMulticastReq(session.User, []int64{1, 2, 3}, buffer.NewNocopyBuffer(message)))

	seq, kind, targets, message, err := protocol.DecodeMulticastReq(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("kind: %v", kind)
	t.Logf("targets: %v", targets)
	t.Logf("message: %v", string(message))
}

func TestEncodeMulticastRes(t *testing.T) {
	buf := protocol.EncodeBuffer(protocol.DataBit, 0, 1, nil, protocol.EncodeMulticastRes(codes.OK, 20))

	t.Log(buf.Bytes())
}

func TestDecodeMulticastRes(t *testing.T) {
	buf := protocol.EncodeBuffer(protocol.DataBit, 0, 1, nil, protocol.EncodeMulticastRes(codes.OK, 20))

	code, total, err := protocol.DecodeMulticastRes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
	t.Logf("total: %v", total)
}
