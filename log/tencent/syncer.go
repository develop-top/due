package tencent

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/develop-top/due/v2/log"
	cls "github.com/tencentcloud/tencentcloud-cls-sdk-go"
)

const (
	fieldKeyLevel     = "level"
	fieldKeyTime      = "time"
	fieldKeyFile      = "file"
	fieldKeyMsg       = "msg"
	fieldKeyStack     = "stack"
	fieldKeyStackFunc = "func"
	fieldKeyStackFile = "file"
)

const Name = "tencent"

type Syncer struct {
	opts       *options
	producer   *cls.AsyncProducerClient
	rawPool    sync.Pool
	bufferPool sync.Pool
}

func NewSyncer(opts ...Option) *Syncer {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	config := cls.GetDefaultAsyncProducerClientConfig()
	config.Endpoint = o.endpoint
	config.AccessKeyID = o.accessKeyID
	config.AccessKeySecret = o.accessKeySecret

	producer, err := cls.NewAsyncProducerClient(config)
	if err != nil {
		panic(err)
	} else {
		producer.Start()
	}

	s := &Syncer{}
	s.opts = o
	s.producer = producer
	s.rawPool = sync.Pool{New: func() any { return make(map[string]string, 5) }}
	s.bufferPool = sync.Pool{New: func() any { return &bytes.Buffer{} }}

	return s
}

// Name 同步器名称
func (s *Syncer) Name() string {
	return Name
}

// Write 写入日志
func (s *Syncer) Write(entity *log.Entity) error {
	return s.producer.SendLog(s.opts.topicID, s.makeLog(entity), nil)
}

// Close 关闭同步器
func (s *Syncer) Close() error {
	return s.producer.Close(60000)
}

// 构建日志
func (s *Syncer) makeLog(entity *log.Entity) *cls.Log {
	raw := s.rawPool.Get().(map[string]string)
	defer func() {
		clear(raw)
		s.rawPool.Put(raw)
	}()

	raw[fieldKeyLevel] = string(entity.Level[:4])
	raw[fieldKeyTime] = entity.Time
	raw[fieldKeyFile] = entity.Caller
	raw[fieldKeyMsg] = entity.Message

	if len(entity.Frames) > 0 {
		b := s.bufferPool.Get().(*bytes.Buffer)
		defer func() {
			b.Reset()
			s.bufferPool.Put(b)
		}()

		fmt.Fprint(b, "[")
		for i, frame := range entity.Frames {
			if i == 0 {
				fmt.Fprintf(b, `{"%s":"%s"`, fieldKeyStackFunc, frame.Function)
			} else {
				fmt.Fprintf(b, `,{"%s":"%s"`, fieldKeyStackFunc, frame.Function)
			}
			fmt.Fprintf(b, `,"%s":"%s:%d"}`, fieldKeyStackFile, frame.File, frame.Line)
		}
		fmt.Fprint(b, "]")

		raw[fieldKeyStack] = b.String()
	}

	return cls.NewCLSLog(time.Now().Unix(), raw)
}
