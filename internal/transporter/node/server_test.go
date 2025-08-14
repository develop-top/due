package node_test

import (
	"context"
	"testing"
	"time"

	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/internal/transporter/internal/server"
	"github.com/develop-top/due/v2/internal/transporter/node"
	"github.com/develop-top/due/v2/log"
)

func TestServer(t *testing.T) {
	server, err := node.NewServer(&provider{}, &server.Options{
		Addr: ":49898",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("server listen on: %s", server.ListenAddr())

	go server.Start()

	<-time.After(20 * time.Second)
}

type provider struct {
}

// Trigger 触发事件
func (p *provider) Trigger(ctx context.Context, gid string, cid, uid int64, event cluster.Event) error {
	return nil
}

// Deliver 投递消息
func (p *provider) Deliver(ctx context.Context, gid, nid string, cid, uid int64, message []byte) error {
	log.Infof("gid: %s, nid: %s, cid: %d, uid: %d message: %s", gid, nid, cid, uid, string(message))
	return nil
}

// GetState 获取状态
func (p *provider) GetState(ctx context.Context) (cluster.State, error) {
	return cluster.Work, nil
}

// SetState 设置状态
func (p *provider) SetState(ctx context.Context, state cluster.State) error {
	return nil
}
