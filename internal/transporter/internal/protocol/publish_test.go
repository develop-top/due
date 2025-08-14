package protocol_test

import (
	"github.com/develop-top/due/v2/internal/transporter/internal/route"
	"testing"

	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"github.com/develop-top/due/v2/packet"
)

func TestEncodePublishReq(t *testing.T) {
	message, err := packet.PackMessage(&packet.Message{
		Route:  1,
		Seq:    2,
		Buffer: []byte("hello world"),
	})
	if err != nil {
		t.Fatal(err)
	}

	buf := protocol.EncodeBuffer(protocol.DataBit, route.Publish, 1, nil, protocol.EncodePublishReq("channel", buffer.NewNocopyBuffer(message)))

	t.Log(buf.Bytes())
}

func TestDecodePublishReq(t *testing.T) {
	message, err := packet.PackMessage(&packet.Message{
		Route:  1,
		Seq:    2,
		Buffer: []byte("hello world"),
	})
	if err != nil {
		t.Fatal(err)
	}

	buf := protocol.EncodeBuffer(protocol.DataBit, route.Publish, 1, nil, protocol.EncodePublishReq("channel", buffer.NewNocopyBuffer(message)))

	seq, channel, message, err := protocol.DecodePublishReq(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("channel: %v", channel)
	t.Logf("message: %v", string(message))
}

func TestEncodePublishRes(t *testing.T) {
	buf := protocol.EncodeBuffer(protocol.DataBit, route.Publish, 1, nil, protocol.EncodePublishRes(10))

	t.Log(buf.Bytes())
}

func TestDecodePublishRes(t *testing.T) {
	buf := protocol.EncodeBuffer(protocol.DataBit, route.Publish, 1, nil, protocol.EncodePublishRes(10))

	total, err := protocol.DecodePublishRes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("total: %v", total)
}
