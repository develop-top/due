package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/session"
	"io"
)

const (
	statReqBytes = b8
	statResBytes = defaultCodeBytes + b64
)

// EncodeStatReq 编码统计在线人数请求
// 协议：session kind
func EncodeStatReq(kind session.Kind) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(statReqBytes)
	writer.WriteUint8s(uint8(kind))
	return buf
}

// DecodeStatReq 解码统计在线人数请求
// 协议：size + header + route + seq + [trace] + session kind
func DecodeStatReq(data []byte) (seq uint64, kind session.Kind, err error) {
	if len(data) < SizeHeadRouteSeqBytes+statReqBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	if _, err = reader.Seek(-statReqBytes, io.SeekEnd); err != nil {
		return
	}

	var k uint8

	if k, err = reader.ReadUint8(); err == nil {
		kind = session.Kind(k)
	}

	return
}

// EncodeStatRes 编码统计在线人数响应
// 协议：code + [total]
func EncodeStatRes(code uint16, total ...uint64) buffer.Buffer {
	size := statResBytes
	if code != codes.OK || len(total) == 0 || total[0] == 0 {
		size -= b64
	}

	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(size)
	writer.WriteUint16s(binary.BigEndian, code)

	if code == codes.OK && len(total) > 0 && total[0] != 0 {
		writer.WriteUint64s(binary.BigEndian, total[0])
	}

	return buf
}

// DecodeStatRes 解码统计在线人数响应
// 协议：size + header + route + seq + [trace] + code + [total]
func DecodeStatRes(data []byte) (code uint16, total uint64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+statResBytes-b64 {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

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

	code, err = reader.ReadUint16(binary.BigEndian)
	if err != nil {
		return
	}

	if code == codes.OK && len(data) == index+statResBytes {
		total, err = reader.ReadUint64(binary.BigEndian)
	}

	return
}
