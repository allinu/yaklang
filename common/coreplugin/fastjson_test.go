package coreplugin

import (
	"testing"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yakgrpc"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

func TestGRPCMUSTPASS_Fastjson(t *testing.T) {
	client, err := yakgrpc.NewLocalClient(true)
	if err != nil {
		panic(err)
	}

	log.Infof("vulAddr: %v", vulAddr)
	//time.Sleep(5 * time.Hour)
	pluginName := "Fastjson 综合检测"
	//wg := sync.WaitGroup{}
	addFastjsonTestCase := func(vulInfo VulInfo, msg ...string) {
		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	Must(TestCoreMitmPlug(pluginName, server, vulInfo, client, t), msg...)
		//}()
		Must(CoreMitmPlugTest(pluginName, server, vulInfo, client, t), msg...)
	}
	//defer wg.Wait()
	vulInGet := VulInfo{
		Path: []string{
			"/fastjson/json-in-query?auth=" + codec.EncodeUrlCode(`{"user":"admin","password":"password"}`) + "&action=login",
		},
		ExpectedResult: map[string]int{
			"目标 fastjson 框架可能存在 RCE 漏洞 (DNSLog Check)": 1,
		},
		StrictMode: true,
		Id:         "json in query test",
	}

	vulInForm := VulInfo{
		Method: "POST",
		Path: []string{
			"/fastjson/json-in-form",
		},
		Headers: []*ypb.KVPair{
			{
				Key:   "Content-Type",
				Value: "application/x-www-form-urlencoded",
			},
		},
		Body: []byte(`auth={"user":"admin","password":"password"}`),
		ExpectedResult: map[string]int{
			"目标 fastjson 框架可能存在 RCE 漏洞 (DNSLog Check)": 1,
		},
		StrictMode: true,
		Id:         "json in form",
	}
	vulInBodyJson := VulInfo{
		Method: "POST",
		Path: []string{
			"/fastjson/json-in-body",
		},
		Body: []byte(`{"user":"admin","password":"password"}`),
		Headers: []*ypb.KVPair{
			{
				Key:   "Content-Type",
				Value: "application/json",
			},
		},
		ExpectedResult: map[string]int{
			"目标 fastjson 框架可能存在 RCE 漏洞 (DNSLog Check)": 1,
		},
		StrictMode: true,
		Id:         "json in body",
	}
	vulInGetServeByJackson := VulInfo{ // 这里不应该检出任何漏洞，并且发包数量应该为 1
		Method: "GET",
		Path: []string{
			"/fastjson/jackson-in-query?auth=" + codec.EncodeUrlCode(`{"user":"admin","password":"password"}`) + "&action=login",
		},
		ExpectedResult: map[string]int{},
		StrictMode:     true,
		Id:             "jackson in query",
	}
	vulInGetIntranet := VulInfo{
		Method: "GET",
		Path: []string{
			"/fastjson/get-in-query-intranet?auth=" + codec.EncodeUrlCode(`{"user":"admin","password":"password"}`) + "&action=login",
		},
		ExpectedResult: map[string]int{
			"目标 fastjson 框架可能存在 RCE 漏洞 (Delay Check)": 1,
		},
		StrictMode: true,
	}

	_ = vulInForm
	_ = vulInBodyJson
	_ = vulInGetServeByJackson
	_ = vulInGetIntranet
	addFastjsonTestCase(vulInGet, "Fastjson 综合检测插件对于 json in query 检测结果不符合预期")
	addFastjsonTestCase(vulInForm, "Fastjson 综合检测插件对于 json in form 检测结果不符合预期")
	addFastjsonTestCase(vulInBodyJson, "Fastjson 综合检测插件对于 json in body 检测结果不符合预期")
	addFastjsonTestCase(vulInGetServeByJackson, "Fastjson 综合检测插件对于 Jackson 检测结果不符合预期")
	addFastjsonTestCase(vulInGetIntranet, "Fastjson 综合检测插件对于 get in query intranet 检测结果不符合预期")
	// TODO: Cookie Fuzz 需要支持自动解码
	//vulInGet := VulInfo{
	//	Method: "GET",
	//	Path: []string{
	//		"/fastjson/json-in-cookie?action=login",
	//	},
	//	Headers: []*ypb.KVPair{
	//		{
	//			Key:   "Cookie",
	//			Value: `auth=` + codec.EncodeBase64Url(`{"id":"-1"}`),
	//		},
	//	},
	//	ExpectedResult: map[string]int{
	//		"目标 fastjson 框架可能存在 RCE 漏洞 (DNSLog Check)": 1,
	//	},
	//	StrictMode: true,
	//}
	//Must(TestCoreMitmPlug(pluginName, server, vulInGet, client, t), "Fastjson 综合检测插件检测结果不符合预期")
	// TODO: Authorization Fuzz 需要支持自动解码
	//vulInAuthorization := VulInfo{
	//	Method: "GET",
	//	Path: []string{
	//		"/fastjson/json-in-authorization?action=login",
	//	},
	//	Headers: []*ypb.KVPair{
	//		{
	//			Key:   "Authorization",
	//			Value: `Basic ` + codec.EncodeBase64Url(`{"user":"admin","password":"password"}`),
	//		},
	//	},
	//	ExpectedResult: map[string]int{
	//		"目标 fastjson 框架可能存在 RCE 漏洞 (DNSLog Check)": 1,
	//	},
	//	StrictMode: true,
	//}
	//addFastjsonTestCase(vulInAuthorization, "Fastjson 综合检测插件对于 Jackson 检测结果不符合预期")
}
func TestFastjson(t *testing.T) {
	client, err := yakgrpc.NewLocalClient()
	if err != nil {
		panic(err)
	}

	log.Infof("vulAddr: %v", vulAddr)
	//time.Sleep(5 * time.Hour)
	pluginName := "Fastjson 综合检测"
	vulInForm := VulInfo{
		Method: "POST",
		Path: []string{
			"/fastjson/unstable-network",
		},
		Headers: []*ypb.KVPair{
			{
				Key:   "Content-Type",
				Value: "application/x-www-form-urlencoded",
			},
		},
		Body:           []byte(`auth={"user":"admin","password":"password"}`),
		ExpectedResult: map[string]int{},
		StrictMode:     true,
		Id:             "json in form",
	}
	Must(CoreMitmPlugTest(pluginName, server, vulInForm, client, t), "Fastjson 综合检测插件检测结果不符合预期")
}
