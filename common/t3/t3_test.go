package t3

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yserx"
	"github.com/yaklang/yaklang/common/yso"
	"testing"
)

func TestSendHello(t *testing.T) {
	t3 := NewT3Client()
	err := t3.ConnectServer("t3://34.228.52.53:7001")
	if err != nil {
		t.Fatal(err)
	}
}
func TestGenerateJvmIdObject(t *testing.T) {
	ser, err := GetObject("jvmid")
	if err != nil {
		t.Fatal(err)
	}
	byt := yserx.MarshalJavaObjects(ser...)
	spew.Dump(codec.EncodeToHex(byt))
	res, err := yso.Dump(ser)
	spew.Dump(res)
}
