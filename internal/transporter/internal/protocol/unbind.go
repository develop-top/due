package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"io"
)

const (
	unbindReqBytes = b64
	unbindResBytes = defaultCodeBytes
)

// EncodeUnbindReq 编码解绑请求
// 协议：uid
func EncodeUnbindReq(uid int64) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(unbindReqBytes)
	writer.WriteInt64s(binary.BigEndian, uid)
	return buf
}

// DecodeUnbindReq 解码解绑请求
// 协议：size + header + route + seq + [trace] + uid
func DecodeUnbindReq(data []byte) (seq uint64, uid int64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+unbindReqBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	if _, err = reader.Seek(-unbindReqBytes, io.SeekEnd); err != nil {
		return
	}

	if uid, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	return
}

// EncodeUnbindRes 编码解绑响应
// 协议：code
func EncodeUnbindRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(unbindResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodeUnbindRes 解码解绑响应
// 协议：size + header + route + seq + [trace] + code
func DecodeUnbindRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+unbindResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-unbindResBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	return
}
