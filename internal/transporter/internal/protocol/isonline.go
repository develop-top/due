package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/session"
	"io"
)

const (
	isOnlineReqBytes = b8 + b64
	isOnlineResBytes = defaultCodeBytes + b8
)

// EncodeIsOnlineReq 编码检测用户是否在线请求
// 协议：session kind + target
func EncodeIsOnlineReq(kind session.Kind, target int64) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(isOnlineReqBytes)
	writer.WriteUint8s(uint8(kind))
	writer.WriteInt64s(binary.BigEndian, target)
	return buf
}

// DecodeIsOnlineReq 解码检测用户是否在线请求
// 协议：size + header + route + seq + [trace] + session kind + target
func DecodeIsOnlineReq(data []byte) (seq uint64, kind session.Kind, target int64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+isOnlineReqBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	if _, err = reader.Seek(-isOnlineReqBytes, io.SeekEnd); err != nil {
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

	return
}

// EncodeIsOnlineRes 编码检测用户是否在线响应
// 协议：code + online state
func EncodeIsOnlineRes(code uint16, isOnline bool) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(isOnlineResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	writer.WriteBools(isOnline)
	return buf
}

// DecodeIsOnlineRes 解码检测用户是否在线响应
// 协议：size + header + route + seq + [trace] + code + online state
func DecodeIsOnlineRes(data []byte) (code uint16, isOnline bool, err error) {
	if len(data) < SizeHeadRouteSeqBytes+isOnlineResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-isOnlineResBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	if isOnline, err = reader.ReadBool(); err != nil {
		return
	}

	return
}
