package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"io"
)

const (
	getStateResBytes = defaultCodeBytes + b8
	setStateReqBytes = b8
	setStateResBytes = defaultCodeBytes
)

// EncodeGetStateRes 编码获取状态响应
// 协议：code + cluster state
func EncodeGetStateRes(code uint16, state cluster.State) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(getStateResBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	writer.WriteUint8s(uint8(state))
	return buf
}

// DecodeGetStateRes 解码获取状态响应
// 协议：size + header + route + seq + [trace] + code + cluster state
func DecodeGetStateRes(data []byte) (code uint16, state cluster.State, err error) {
	if len(data) < SizeHeadRouteSeqBytes+getStateResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-getStateResBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	if status, e := reader.ReadUint8(); e != nil {
		err = e
	} else {
		state = cluster.State(status)
	}

	return
}

// EncodeSetStateReq 编码设置状态请求
// 协议：cluster state
func EncodeSetStateReq(state cluster.State) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(setStateReqBytes)
	writer.WriteUint8s(uint8(state))
	return buf
}

// DecodeSetStateReq 解码设置状态请求
// 协议：size + header + route + seq + [trace] + cluster state
func DecodeSetStateReq(data []byte) (seq uint64, state cluster.State, err error) {
	if len(data) < SizeHeadRouteSeqBytes+setStateReqBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	if _, err = reader.Seek(-setStateReqBytes, io.SeekEnd); err != nil {
		return
	}

	if status, e := reader.ReadUint8(); e != nil {
		err = e
	} else {
		state = cluster.State(status)
	}

	return
}

// EncodeSetStateRes 编码设置状态响应
// 协议：code
func EncodeSetStateRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(setStateReqBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodeSetStateRes 解码绑定响应
// 协议：size + header + route + seq + [trace] + code
func DecodeSetStateRes(data []byte) (code uint16, err error) {
	if len(data) < SizeHeadRouteSeqBytes+setStateResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-setStateResBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	return
}
