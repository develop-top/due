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
	broadcastReqBytes = b8
	broadcastResBytes = defaultCodeBytes + b64
)

// EncodeBroadcastReq 编码广播请求
// 协议：session kind + <message packet>
func EncodeBroadcastReq(kind session.Kind, message buffer.Buffer) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(broadcastReqBytes)
	writer.WriteUint8s(uint8(kind))
	buf.Mount(message)
	return buf
}

// DecodeBroadcastReq 解码广播请求
// 协议：size + header + route + seq + [trace] + session kind + <message packet>
func DecodeBroadcastReq(data []byte) (seq uint64, kind session.Kind, message []byte, err error) {
	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	var header uint8
	header, err = DecodeHeader(reader)
	if err != nil {
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

	message = data[index+broadcastReqBytes:]

	return
}

// EncodeBroadcastRes 编码广播响应
// 协议：code + [total]
func EncodeBroadcastRes(code uint16, total ...uint64) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(broadcastResBytes)
	writer.WriteUint16s(binary.BigEndian, code)

	if code == codes.OK && len(total) > 0 && total[0] != 0 {
		writer.WriteUint64s(binary.BigEndian, total[0])
	}

	return buf
}

// DecodeBroadcastRes 解码广播响应
// 协议：size + header + route + seq + [trace] + code + [total]
func DecodeBroadcastRes(data []byte) (code uint16, total uint64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+broadcastResBytes-b64 {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	var header uint8
	header, err = DecodeHeader(reader)
	if err != nil {
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

	if code == codes.OK && len(data) == index+broadcastResBytes {
		total, err = reader.ReadUint64(binary.BigEndian)
	}

	return
}
