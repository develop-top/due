package protocol

import (
	"encoding/binary"
	"github.com/develop-top/due/v2/errors"
	"io"

	"github.com/develop-top/due/v2/core/buffer"
)

const (
	publishReqBytes = b8
	publishResBytes = b64
)

// EncodePublishReq 编码发布频道消息请求
// 协议：channel len + channel + <message packet>
func EncodePublishReq(channel string, message buffer.Buffer) buffer.Buffer {
	channelBytes := len([]byte(channel))
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(publishReqBytes + channelBytes)
	writer.WriteUint8s(uint8(channelBytes))
	writer.WriteString(channel)
	buf.Mount(message)

	return buf
}

// DecodePublishReq 解码发布频道消息请求
// 协议：size + header + route + seq + [trace] + channel len + channel + <message packet>
func DecodePublishReq(data []byte) (seq uint64, channel string, message []byte, err error) {
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

	var channelBytes uint8

	if channelBytes, err = reader.ReadUint8(); err != nil {
		return
	}

	if channel, err = reader.ReadString(int(channelBytes)); err != nil {
		return
	}

	message = data[int(index)+b8+int(channelBytes):]

	return
}

// EncodePublishRes 编码发布频道消息响应
// 协议：total
func EncodePublishRes(total uint64) buffer.Buffer {
	buf := buffer.NewNocopyBuffer()
	writer := buf.Malloc(publishResBytes)
	writer.WriteUint64s(binary.BigEndian, total)

	return buf
}

// DecodePublishRes 解码组播响应
// 协议：size + header + route + seq + [trace] + total
func DecodePublishRes(data []byte) (total uint64, err error) {
	if len(data) < SizeHeadRouteSeqBytes+publishResBytes {
		err = errors.ErrInvalidMessage
		return
	}

	reader := buffer.NewReader(data)

	if _, err = reader.Seek(-publishResBytes, io.SeekEnd); err != nil {
		return
	}

	if total, err = reader.ReadUint64(binary.BigEndian); err != nil {
		return
	}

	return
}
