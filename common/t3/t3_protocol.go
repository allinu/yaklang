package t3

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type T3Client struct {
	timeout  time.Duration
	protocol string
	conn     net.Conn
	debug    bool
	Version  string
	As       int // stack容量（AbbrevSize）
	Hl       int // t3协议头长度（HeaderLength）
}

func (t3 *T3Client) SetDebug(d bool) {
	t3.debug = d
}
func (t3 *T3Client) ToString() string {
	return fmt.Sprintf("t3 %s\nAS:%d\nHL:%d\n\n", t3.Version, t3.As, t3.Hl)
}
func (t3 *T3Client) ConnectServer(addr string, proxys ...string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return utils.Errorf("parse url error: %v", err)
	}
	helloReq := fmt.Sprintf("%s %s\nAS:%d\nHL:%d\n\n", u.Scheme, t3.Version, t3.As, t3.Hl)
	t3.conn, err = utils.TCPConnect(u.Host, t3.timeout, proxys...)
	if err != nil {
		return utils.Errorf("connect error: %v", err)
	}
	n, err := t3.conn.Write([]byte(helloReq)) // send hello
	if err != nil {
		return utils.Errorf("send header error: %v", err)
	}
	log.Debugf("send hello request successful, length: %d", n)

	// read hello response
	buf := bufio.NewReader(t3.conn)
	line, _, err := buf.ReadLine() // check login success
	if err != nil {
		return utils.Errorf("read login info failed: %v", err)
	}
	splits := strings.Split(string(line), ":")
	if !(len(splits) > 0 && utils.StringArrayContains([]string{"HELO", "LGIN", "SERV", "UNAV", "LICN", "RESC", "VERS", "CATA", "CMND"}, splits[0])) {
		return utils.Error("login failed")
	}
	readHeaderVar := func(line []byte) string {
		n = bytes.IndexByte(line, ':') // check version
		if n != -1 && len(line) > n+1 {
			return string(line[n+1:])
		}
		return ""
	}
	version := readHeaderVar(line) // check version
	if version != "" {
		log.Debugf("remote version: %v", version)
	} else {
		return utils.Error("check version failed")
	}

	// read bootstrap params
	for {
		line, _, err = buf.ReadLine()
		if err != nil {
			return utils.Errorf("read bootstrap params failed: %v", err)
		}
		if strings.HasPrefix(string(line), "AS") {
			as := readHeaderVar(line)
			if as != "" {
				n, err := strconv.Atoi(as)
				if err != nil {
					return utils.Errorf("parse AS error: %v", err)
				}
				if n < t3.As {
					t3.As = n
				}
			}
		}
		if strings.HasPrefix(string(line), "HL") {
			hl := readHeaderVar(line)
			if hl != "" {
				n, err := strconv.Atoi(hl)
				if err != nil {
					return utils.Errorf("parse AS error: %v", err)
				}
				t3.Hl = n
			}
		}
		if len(line) == 0 {
			break
		}
	}
	helloRes := ""
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			return utils.Errorf("read hello response error: %v", err)
		}
		helloRes += string(append(line, '\n'))
		if len(line) == 0 {
			break
		}
	}

	res := utils.StableReader(t3.conn, 1*time.Second, 1024)
	spew.Dump(string(res))
	return nil
}
func NewT3Client() *T3Client {
	return &T3Client{
		Version: "10.3.1",
		As:      255,
		Hl:      19,
		timeout: 3 * time.Second,
	}
}
