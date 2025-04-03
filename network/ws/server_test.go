/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/5/29 10:59 上午
 * @Desc: TODO
 */

package ws_test

import (
	"github.com/develop-top/due/network/ws/v2"
	"github.com/develop-top/due/v2/log"
	"github.com/develop-top/due/v2/network"
	"github.com/develop-top/due/v2/packet"
	"github.com/develop-top/due/v2/utils/xcall"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	server := ws.NewServer()
	server.OnStart(func() {
		t.Logf("server is started")
	})
	server.OnConnect(func(conn network.Conn) {
		t.Logf("connection is opened, connection id: %d", conn.ID())
	})
	server.OnDisconnect(func(conn network.Conn) {
		t.Logf("connection is closed, connection id: %d", conn.ID())
	})
	server.OnReceive(func(conn network.Conn, msg []byte) {
		message, err := packet.UnpackMessage(msg)
		if err != nil {
			t.Error(err)
			return
		}

		t.Logf("receive msg from client, connection id: %d, seq: %d, route: %d, msg: %s", conn.ID(), message.Seq, message.Route, string(message.Buffer))

		msg, err = packet.PackMessage(&packet.Message{
			Seq:    1,
			Route:  1,
			Buffer: []byte("I'm fine~~"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if err = conn.Push(msg); err != nil {
			t.Error(err)
		}
	})
	server.OnUpgrade(func(w http.ResponseWriter, r *http.Request) (allowed bool) {
		return true
	})

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	xcall.Go(func() {
		err := http.ListenAndServe(":8089", nil)
		if err != nil {
			log.Errorf("pprof server start failed: %v", err)
		}
	})

	select {}
}

func TestServer_Benchmark(t *testing.T) {
	server := ws.NewServer()
	server.OnStart(func() {
		t.Logf("server is started")
	})
	server.OnReceive(func(conn network.Conn, msg []byte) {
		_, err := packet.UnpackMessage(msg)
		if err != nil {
			t.Error(err)
			return
		}

		msg, err = packet.PackMessage(&packet.Message{
			Seq:    1,
			Route:  1,
			Buffer: []byte("I'm fine~~"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if err = conn.Push(msg); err != nil {
			t.Error(err)
		}
	})
	server.OnUpgrade(func(w http.ResponseWriter, r *http.Request) (allowed bool) {
		return true
	})

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	xcall.Go(func() {
		err := http.ListenAndServe(":8089", nil)
		if err != nil {
			log.Errorf("pprof server start failed: %v", err)
		}
	})

	select {}
}
