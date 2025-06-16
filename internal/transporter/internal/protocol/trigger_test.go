package protocol_test

import (
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"testing"
)

func TestEncodeTriggerReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeTriggerReq(cluster.Disconnect, 1))

	t.Log(buffer.Bytes())
}

func TestDecodeTriggerReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeTriggerReq(cluster.Disconnect, 1, 2))

	seq, evt, cid, uid, err := protocol.DecodeTriggerReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("evt: %v", evt)
	t.Logf("cid: %v", cid)
	t.Logf("uid: %v", uid)
}

func TestEncodeTriggerRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeTriggerRes(codes.InternalError))

	t.Log(buffer.Bytes())
}

func TestDecodeTriggerRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeTriggerRes(codes.InternalError))

	code, err := protocol.DecodeTriggerRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
}
