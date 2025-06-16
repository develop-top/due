package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/session"
	"io"
)

const (
	pushReqBytes = b8 + b64
	pushResBytes = defaultCodeBytes
)

// EncodePushReq 编码推送请求
// 协议：session kind + target + <message packet>
func EncodePushReq(kind session.Kind, target int64, message buffer.Buffer) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(pushReqBytes)
	writer.WriteUint8s(uint8(kind))
	writer.WriteInt64s(binary.BigEndian, target)
	buf.Mount(message)
	return buf
}

// DecodePushReq 解码推送消息
// 协议：size + header + route + seq + [trace] + session kind + target + <message packet>
func DecodePushReq(data []byte) (seq uint64, kind session.Kind, target int64, message []byte, err error) {
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

	if target, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	message = data[index+pushReqBytes:]

	return
}

// EncodePushRes 编码推送响应
// 协议：code
func EncodePushRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(pushResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodePushRes 解码推送响应
// 协议：size + header + route + seq + [trace] + code
func DecodePushRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+pushResBytes {
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
