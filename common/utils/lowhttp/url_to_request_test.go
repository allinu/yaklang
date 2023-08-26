package lowhttp

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUrlToGetRequestPacket(t *testing.T) {
	res := UrlToGetRequestPacket("https://baidu.com/asdfasdfasdf", []byte(`GET / HTTP/1.1
Host: baidu.com
Cookie: test=12;`), false)
	spew.Dump(res)
}

func TestUrlToHTTPRequest(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "raw path",
			args: args{text: "http://127.0.0.1:1231/abcdef%2f?a=1&b=2%2f"},
			want: []byte("GET /abcdef%2f?a=1&b=2%2f HTTP/1.1\r\nHost: 127.0.0.1:1231\r\n\r\n"),
		},
		{
			name: "raw fragment",
			args: args{text: "http://127.0.0.1:1231/abcdef/?a=1&b=2%2f#123%3E"},
			want: []byte("GET /abcdef/?a=1&b=2%2f#123%3E HTTP/1.1\r\nHost: 127.0.0.1:1231\r\n\r\n"),
		},
		{
			name: "raw fragment 2",
			args: args{text: "http://127.0.0.1:1231/abcdef/?a=1&b=2%2f#123%3E#"},
			want: []byte("GET /abcdef/?a=1&b=2%2f#123%3E# HTTP/1.1\r\nHost: 127.0.0.1:1231\r\n\r\n"),
		},
		{
			name: "end fragment",
			args: args{text: "http://127.0.0.1:1231/#"},
			want: []byte("GET /# HTTP/1.1\r\nHost: 127.0.0.1:1231\r\n\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UrlToHTTPRequest(tt.args.text)
			if err != nil {
				t.FailNow()
				return
			}
			assert.Equalf(t, tt.want, got, "UrlToHTTPRequest(%v)", tt.args.text)
		})
	}
}
