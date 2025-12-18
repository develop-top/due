package xcall

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/develop-top/due/v2/log"
)

type Counter interface {
	Inc(labels ...string)
}

var metricAgent Counter

func RegisterMetricsAgent(agent Counter) {
	metricAgent = agent
}

// Call 安全地调用函数
func Call(fn func()) {
	if fn == nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			var skip int
			switch err.(type) {
			case runtime.Error:
				skip = 3
				log.Panic(err)
			default:
				skip = 2
				log.Panicf("panic error: %v", err)
			}
			if metricAgent != nil {
				_, file, line, _ := runtime.Caller(skip)
				location := fmt.Sprintf("%s:%d", file, line)
				metricAgent.Inc(location)
			}
		}
	}()

	fn()
}

// Go 执行单个协程
func Go(fn func()) {
	go Call(fn)
}

// GoWithTimeout 执行多个协程（附带超时时间）
func GoWithTimeout(timeout time.Duration, fns ...func()) {
	NewGoroutines().Add(fns...).Run(context.Background(), timeout)
}

// GoWithDeadline 执行多个协程（附带最后期限）
func GoWithDeadline(deadline time.Time, fns ...func()) {
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	NewGoroutines().Add(fns...).Run(ctx)
}
