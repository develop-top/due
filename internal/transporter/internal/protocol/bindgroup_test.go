package protocol_test

import (
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/internal/transporter/internal/protocol"
	"testing"
)

func TestEncodeBindGroupsReq(t *testing.T) {
	buffer := protocol.EncodeBindGroupsReq(2, []int64{3, 4, 5})

	t.Log(buffer.Bytes())
}

func TestDecodeBindGroupsReq(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeBindGroupsReq(2, []int64{3, 4, 5}))

	seq, cid, groups, err := protocol.DecodeBindGroupsReq(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("seq: %v", seq)
	t.Logf("cid: %v", cid)
	t.Logf("groups: %v", groups)
}

func TestEncodeBindGroupsRes(t *testing.T) {
	buffer := protocol.EncodeBindGroupsRes(codes.OK)

	t.Log(buffer.Bytes())
}

func TestDecodeBindGroupsRes(t *testing.T) {
	buffer := protocol.EncodeBuffer(0, 0, 1, nil, protocol.EncodeBindGroupsRes(codes.InternalError))

	code, err := protocol.DecodeBindGroupsRes(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %v", code)
}
