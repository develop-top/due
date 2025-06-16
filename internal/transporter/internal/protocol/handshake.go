package protocol

import (
	"encoding/binary"
	"io"

	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
)

const (
	handshakeReqBytes = b8
	handshakeResBytes = defaultCodeBytes
)

// EncodeHandshakeReq 编码握手请求
// 协议：ins kind + ins id
func EncodeHandshakeReq(insKind cluster.Kind, insID string) buffer.Buffer {
	size := handshakeReqBytes + len(insID)
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(size)
	writer.WriteUint8s(uint8(insKind))
	writer.WriteString(insID)
	return buf
}

// DecodeHandshakeReq 解码握手请求
// 协议：size + header + route + seq + ins kind + ins id
func DecodeHandshakeReq(data []byte) (seq uint64, insKind cluster.Kind, insID string, err error) {
	reader := buffer.NewReader(data)

	if _, err = reader.Seek(defaultSizeBytes+defaultHeaderBytes+defaultRouteBytes, io.SeekStart); err != nil {
		return
	}

	if seq, err = reader.ReadUint64(binary.BigEndian); err != nil {
		return
	}

	var k uint8
	if k, err = reader.ReadUint8(); err != nil {
		return
	} else {
		insKind = cluster.Kind(k)
	}

	if insID, err = reader.ReadString(len(data) - SizeHeadRouteSeqBytes - handshakeReqBytes); err != nil {
		return
	}

	return
}

// EncodeHandshakeRes 编码握手响应
// 协议：code
func EncodeHandshakeRes(code uint16) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(defaultCodeBytes)
	writer.WriteUint16s(binary.BigEndian, code)
	return buf
}

// DecodeHandshakeRes 解码握手响应
// 协议：size + header + route + seq + code
func DecodeHandshakeRes(data []byte) (code uint16, err error) {
	if len(data) != SizeHeadRouteSeqBytes+handshakeResBytes {
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
