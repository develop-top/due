package server

import "context"

type RouteHandler func(ctx context.Context, conn *Conn, data []byte) error

type chData struct {
	isHeartbeat bool   // 是否心跳
	route       uint8  // 路由
	data        []byte // 数据
	trace       []byte // 链路追踪上下文
}
