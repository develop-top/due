package protocol

const (
	defaultSizeBytes   = 4  // 包长度字节数
	defaultHeaderBytes = 1  // 头信息字节数
	defaultSeqBytes    = 8  // 序列号字节数
	defaultRouteBytes  = 1  // 路由号字节数
	defaultCodeBytes   = 2  // 错误码字节数
	defaultTraceBytes  = 25 // 链路追踪上下文字节数
)

const (
	DataBit      uint8 = 0 << 7 // 数据标识位
	HeartbeatBit uint8 = 1 << 7 // 心跳标识位
	TraceBit     uint8 = 1 << 6 // 链路追踪标识位
)

const (
	b8 = 1 << iota
	b16
	b32
	b64
)
