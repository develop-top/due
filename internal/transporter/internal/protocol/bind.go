package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"io"
)

const (
	bindReqBytes = b64 + b64
	bindResBytes = defaultCodeBytes
)

// EncodeBindReq 编码绑定请求
// 协议：cid + uid
func EncodeBindReq(cid, uid int64) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(bindReqBytes)
	writer.WriteInt64s(binary.BigEndian, cid, uid)
	return buf
}

// DecodeBindReq 解码绑定请求
// 协议：size + header + route + seq + [trace] + cid + uid
func DecodeBindReq(data []byte) (seq uint64, cid, uid int64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+bindReqBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	if _, err = reader.Seek(-bindReqBytes, io.SeekEnd); err != nil {
		return
	}

	if cid, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	if uid, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	return
}

// EncodeBindRes 编码绑定响应
// 协议：code
func EncodeBindRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(bindResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodeBindRes 解码绑定响应
// 协议：size + header + route + seq + [trace] + code
func DecodeBindRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+bindResBytes {
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
