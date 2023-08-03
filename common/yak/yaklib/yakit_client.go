package yaklib

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/antlr4yak"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"net/http"
	"time"
)

type YakitMessageHandleFunc func(i interface{}) error

func NewVirtualYakitClient(h YakitMessageHandleFunc) *YakitClient {
	remoteClient := NewYakitClient("")
	remoteClient.send = func(i interface{}) error {
		return h(i)
	}
	return remoteClient
}

func NewVirtualYakitClientWithExecResult(h func(result *ypb.ExecResult) error) *YakitClient {
	return NewVirtualYakitClient(func(i interface{}) error {
		switch ret := i.(type) {
		case *ypb.ExecResult:
			return h(ret)
		case *YakitLog:
			return h(NewYakitLogExecResult(ret.Level, ret.Data))
		default:
			log.Warnf("unhandled yakit client message: %v", spew.Sdump(ret))
		}
		return nil
	})
}

func RawHandlerToExecOutput(h func(any) error) func(result *ypb.ExecResult) error {
	return func(result *ypb.ExecResult) error {
		return h(result)
	}
}

type YakitClient struct {
	addr      string
	client    *http.Client
	yakLogger *YakLogger
	send      func(i interface{}) error
}

func NewYakitClient(addr string) *YakitClient {
	logger := CreateYakLogger()
	client := &YakitClient{
		addr:      addr,
		yakLogger: logger,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionSSL30, // nolint[:staticcheck]
					MaxVersion:         tls.VersionTLS13,
				},
				TLSHandshakeTimeout:   10 * time.Second,
				DisableCompression:    true,
				MaxIdleConns:          1,
				MaxIdleConnsPerHost:   1,
				MaxConnsPerHost:       1,
				IdleConnTimeout:       5 * time.Minute,
				ResponseHeaderTimeout: 30 * time.Second,
				ExpectContinueTimeout: 30 * time.Second,
			},
			Timeout: 15 * time.Second,
		},
	}

	client.send = func(i interface{}) error {
		if client == nil {
			return utils.Errorf("no client set")
		}

		if client.addr == "" {
			return nil
		}

		msgRaw, err := YakitMessageGenerator(i)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("GET", client.addr, bytes.NewBuffer(msgRaw))
		if err != nil {
			return utils.Errorf("build http request failed: %s", err)
		}
		_, err = client.client.Do(req)
		if err != nil {
			log.Errorf("client failed: %s", err)
			return err
		}
		return nil
	}
	client.client.Timeout = 15 * time.Second
	return client
}
func (c *YakitClient) SetYakLog(logger *YakLogger) {
	c.yakLogger = logger
}

// 输入
func (c *YakitClient) YakitLog(level string, tmp string, items ...interface{}) error {
	data := fmt.Sprintf(tmp, items...)
	return c.send(&YakitLog{
		Level:     level,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func (c *YakitClient) YakitDraw(level string, data interface{}) {
	err := c.send(&YakitLog{
		Level:     level,
		Data:      utils.InterfaceToString(data),
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		log.Error(err)
	}
}
func (c *YakitClient) Output(i interface{}) error {
	level, msg := MarshalYakitOutput(i)
	return c.YakitLog(level, msg)
}
func (c *YakitClient) SendRaw(y *YakitLog) error {
	if c == nil {
		return utils.Error("no client")
	}
	return c.send(y)
}

func SetEngineClient(e *antlr4yak.Engine, client *YakitClient) {
	//修改yakit库的客户端
	e.ImportSubLibs("yakit", GetExtYakitLibByClient(client))

	//修改全局默认客户端
	InitYakit(client)
}
