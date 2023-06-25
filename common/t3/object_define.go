package t3

import (
	"errors"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yserx"
)

var objectsHex = map[string]string{
	"jvmid": "aced0005737200137765626c6f6769632e726a766d2e4a564d4944dc49c23ede121e2a0c00007870771c01000000000000000100093132372e302e302e3183b579520000000078",
}

func GetObject(name string) ([]yserx.JavaSerializable, error) {
	if v, ok := objectsHex[name]; ok {
		vb, err := codec.DecodeHex(v)
		if err != nil {
			return nil, err
		}
		return yserx.ParseJavaSerialized(vb)
	}
	return nil, errors.New("not found object")
}
