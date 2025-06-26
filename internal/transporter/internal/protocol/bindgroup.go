package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"io"
)

const (
	bindGroupsReqBytes = b64
	bindGroupsResBytes = defaultCodeBytes
)

// EncodeBindGroupsReq 编码绑定组请求
// 协议：cid + groups
func EncodeBindGroupsReq(cid int64, groups []int64) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(bindGroupsReqBytes + len(groups)*b64)
	writer.WriteInt64s(binary.BigEndian, cid)
	writer.WriteInt64s(binary.BigEndian, groups...)
	return buf
}

// DecodeBindGroupsReq 解码绑定组请求
// 协议：size + header + route + seq + [trace] + cid + groups
func DecodeBindGroupsReq(data []byte) (seq uint64, cid int64, groups []int64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+bindGroupsReqBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	var header uint8
	if header, err = DecodeHeader(reader); err != nil {
		return
	}
	index := SizeHeadRouteSeqBytes
	if header&TraceBit == TraceBit {
		index += defaultTraceBytes
	}

	if _, err = reader.Seek(int64(index), io.SeekStart); err != nil {
		return
	}

	if cid, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	count := (len(data) - index - bindGroupsReqBytes + 1) / b64

	if groups, err = reader.ReadInt64s(binary.BigEndian, count); err != nil {
		return
	}

	return
}

// EncodeBindGroupsRes 编码绑定组响应
// 协议：code
func EncodeBindGroupsRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(bindGroupsResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodeBindGroupsRes 解码绑定组响应
// 协议：size + header + route + seq + [trace] + code
func DecodeBindGroupsRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+bindGroupsResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-defaultCodeBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	return
}
