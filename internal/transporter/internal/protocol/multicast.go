package protocol

import (
	"encoding/binary"
	"io"

	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/session"
)

const (
	multicastReqBytes = b8 + b16
	multicastResBytes = defaultCodeBytes + b64
)

// EncodeMulticastReq 编码组播请求（最多组播65535个对象）
// 协议：session kind + count + targets + <message packet>
func EncodeMulticastReq(kind session.Kind, targets []int64, message buffer.Buffer) buffer.Buffer {
	size := multicastReqBytes + len(targets)*b64
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(size)
	writer.WriteUint8s(uint8(kind))
	writer.WriteUint16s(binary.BigEndian, uint16(len(targets)))
	writer.WriteInt64s(binary.BigEndian, targets...)
	buf.Mount(message)
	return buf
}

// DecodeMulticastReq 解码组播请求
// 协议：size + header + route + seq + [trace] + session kind + count + targets + <message packet>
func DecodeMulticastReq(data []byte) (seq uint64, kind session.Kind, targets []int64, message []byte, err error) {
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

	var k uint8
	if k, err = reader.ReadUint8(); err != nil {
		return
	} else {
		kind = session.Kind(k)
	}

	count, err := reader.ReadUint16(binary.BigEndian)
	if err != nil {
		return
	}

	if targets, err = reader.ReadInt64s(binary.BigEndian, int(count)); err != nil {
		return
	}

	message = data[uint16(index)+multicastReqBytes+8*count:]

	return
}

// EncodeMulticastRes 编码组播响应
// 协议：code + [total]
func EncodeMulticastRes(code uint16, total ...uint64) buffer.Buffer {
	size := multicastResBytes
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

// DecodeMulticastRes 解码组播响应
// 协议：size + header + route + seq + [trace] + code + [total]
func DecodeMulticastRes(data []byte) (code uint16, total uint64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+multicastResBytes-b64 {
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

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	if code == codes.OK && len(data) == index+multicastResBytes {
		total, err = reader.ReadUint64(binary.BigEndian)
	}

	return
}
