package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"io"
)

const (
	triggerReqBytes = b8 + b64 + b64
	triggerResBytes = defaultCodeBytes
)

// EncodeTriggerReq 编码触发事件请求
// 协议：event + cid + [uid]
func EncodeTriggerReq(event cluster.Event, cid int64, uid ...int64) buffer.Buffer {
	size := triggerReqBytes
	if len(uid) == 0 || uid[0] == 0 {
		size -= b64
	}

	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(size)
	writer.WriteUint8s(uint8(event))
	writer.WriteInt64s(binary.BigEndian, cid)

	if len(uid) > 0 && uid[0] != 0 {
		writer.WriteInt64s(binary.BigEndian, uid[0])
	}

	return buf
}

// DecodeTriggerReq 解码触发事件请求
// 协议：size + header + route + seq + [trace] + event + cid + [uid]
func DecodeTriggerReq(data []byte) (seq uint64, event cluster.Event, cid int64, uid int64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+triggerReqBytes-b64 {
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

	var evt uint8
	if evt, err = reader.ReadUint8(); err != nil {
		return
	} else {
		event = cluster.Event(evt)
	}

	if cid, err = reader.ReadInt64(binary.BigEndian); err != nil {
		return
	}

	if len(data) == index+triggerReqBytes {
		uid, err = reader.ReadInt64(binary.BigEndian)
	}

	return
}

// EncodeTriggerRes 编码触发事件响应
// 协议：code
func EncodeTriggerRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(triggerResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodeTriggerRes 解码触发事件响应
// 协议：size + header + route + seq + [trace] + code
func DecodeTriggerRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+triggerResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-triggerResBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	return
}
