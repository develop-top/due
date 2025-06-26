package protocol_test

import (
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"testing"
)

func TestEncodeUnbindGroupsReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeUnbindGroupsReq(2, []int64{3, 4, 5}))

	t.Log(buffer.Bytes())
}

func TestDecodeUnbindGroupsReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeUnbindGroupsReq(2, []int64{3, 4, 5}))

	seq, cid, groups, err := protocol.DecodeUnbindGroupsReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("cid: %v", cid)
	t.Logf("groups: %v", groups)
}

func TestEncodeUnbindGroupsRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeUnbindGroupsRes(2))

	t.Log(buffer.Bytes())
}

func TestDecodeUnbindGroupsRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeUnbindGroupsRes(2))

	code, err := protocol.DecodeUnbindGroupsRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
}
