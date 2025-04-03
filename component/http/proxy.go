package http

import (
	"github.com/develop-top/due/v2/errors"
	"github.com/develop-top/due/v2/transport"
)

type Proxy struct {
	server *Server
}

func newProxy(s *Server) *Proxy {
	return &Proxy{server: s}
}

// Router 获取路由器
func (p *Proxy) Router() Router {
	return &router{app: p.server.app, proxy: p}
}

// NewMeshClient 新建微服务客户端
// target参数可分为三种模式:
// 服务直连模式: 	direct://127.0.0.1:8011
// 服务直连模式: 	direct://711baf8d-8a06-11ef-b7df-f4f19e1f0070
// 服务发现模式: 	discovery://service_name
func (p *Proxy) NewMeshClient(target string) (transport.Client, error) {
	if p.server.opts.transporter == nil {
		return nil, errors.ErrMissTransporter
	}

	return p.server.opts.transporter.NewClient(target)
}
