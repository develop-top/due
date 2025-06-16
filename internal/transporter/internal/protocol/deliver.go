package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"io"
)

const (
	deliverReqBytes = b64 + b64
	deliverResBytes = defaultCodeBytes
)

// EncodeDeliverReq 编码投递消息请求
// 协议：cid + uid + <message packet>
func EncodeDeliverReq(cid int64, uid int64, message []byte) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(deliverReqBytes)
	writer.WriteInt64s(binary.BigEndian, cid, uid)
	buf.Mount(message)
	return buf
}

// DecodeDeliverReq 解码投递消息请求
// 协议：size + header + route + seq + [trace] + cid + uid + <message packet>
func DecodeDeliverReq(data []byte) (seq uint64, cid int64, uid int64, message []byte, err error) {
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

	if cid, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	if uid, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	message = data[index+deliverReqBytes:]

	return
}

// EncodeDeliverRes 编码投递消息响应
// 协议：code
func EncodeDeliverRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(deliverResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodeDeliverRes 解码投递消息响应
// 协议：size + header + route + seq + [trace] + code
func DecodeDeliverRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+deliverResBytes {
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
