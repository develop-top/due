package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/core/buffer"
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/internal/transporter/internal/codes"
	"github.com/develop-top/due/v2/session"
	"github.com/develop-top/due/v2/utils/xnet"
	"io"
)

const (
	getIPReqBytes = b8 + b64
	getIPResBytes = defaultCodeBytes + b32
)

// EncodeGetIPReq 编码获取IP请求
// 协议：session kind + target
func EncodeGetIPReq(kind session.Kind, target int64) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(getIPReqBytes)
	writer.WriteUint8s(uint8(kind))
	writer.WriteInt64s(binary.BigEndian, target)
	return buf
}

// DecodeGetIPReq 解码获取IP请求
// 协议：size + header + route + seq + [trace] + session kind + target
func DecodeGetIPReq(data []byte) (seq uint64, kind session.Kind, target int64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+getIPReqBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	seq, err = DecodeSeq(reader)
	if err != nil {
		return
	}

	if _, err = reader.Seek(-getIPReqBytes, io.SeekEnd); err != nil {
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

// EncodeGetIPRes 编码获取IP响应
// 协议：code + ip
func EncodeGetIPRes(code uint16, ip ...string) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(getIPResBytes)
	writer.WriteUint16s(binary.BigEndian, code)

	if code == codes.OK && len(ip) > 0 && ip[0] != "" {
		writer.WriteUint32s(binary.BigEndian, xnet.IP2Long(ip[0]))
	}

	return buf
}

// DecodeGetIPRes 解码获取IP响应
// 协议：size + header + route + seq + [trace] + code + ip
func DecodeGetIPRes(data []byte) (code uint16, ip string, err error) {
	if len(data) < SizeHeadRouteSeqBytes+getIPResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-getIPReqBytes, io.SeekEnd); err != nil {
		return
	}

	if code, err = reader.ReadUint16(binary.BigEndian); err != nil {
		return
	}

	if code == codes.OK {
		var v uint32
		if v, err = reader.ReadUint32(binary.BigEndian); err != nil {
			return
		} else {
			ip = xnet.Long2IP(v)
		}
	}

	return
}
