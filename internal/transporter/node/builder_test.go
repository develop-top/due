package node_test

import (
	"context"
	"github.com/develop-top/due/v2/cluster"
	"github.com/develop-top/due/v2/internal/transporter/node"
	"github.com/develop-top/due/v2/utils/xuuid"
	"testing"
)

func TestBuilder(t *testing.T) {
	builder := node.NewBuilder(&node.Options{
		InsID:   xuuid.UUID(),
		InsKind: cluster.Gate,
	})

	client, err := builder.Build("127.0.0.1:49898")
	if err != nil {
		t.Fatal(err)
	}

	err = client.Deliver(context.Background(), 1, 2, []byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}
}
