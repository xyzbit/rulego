package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xyzbit/rulego"
	"github.com/xyzbit/rulego/api/types"
	_ "github.com/xyzbit/rulego/testdata/pb" // 注册pb信息
)

// 注意：需要启动一个 grpc 服务，实现 ./testdata/pb 目录下的proto接口
func main() {
	config := rulego.NewConfig(
		types.WithDefaultPool(),
	)

	// 案例一：一个 grpc 的 http 代理的案例
	// 1.js请求数据转化 2.调用grpc服务 3.js响应数据转化
	fmt.Println("案例一：一个 grpc 的 http 代理的案例")
	ruleEngine, err := rulego.New("grpcHttpProxy", []byte(grpcHttpProxyFile), rulego.WithConfig(config))
	if err != nil {
		panic(err)
	}

	metaData := types.NewMetadata()
	// 将 http 请求的数据转化为 json 传入
	msg := types.NewMsg(time.Now().Unix(), "TEST_MSG_TYPE1", types.JSON, metaData, `{"name": "RULEGO","is_login": "true"}`)

	ruleEngine.OnMsgWithOptions(msg, types.WithEndFunc(func(msg types.RuleMsg, err error) {
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(msg)
	}))
	time.Sleep(time.Second * 5)

	fmt.Println("\n\n案例二：编排多个 grpc 的案例")
	// 初始化子规则链实例
	_, err = rulego.New("callMultipleGrpc", []byte(callMultipleGrpc), rulego.WithConfig(config))
	if err != nil {
		panic(err)
	}
	ruleEngine2, err := rulego.New("orchestratingMultipleGrpc", []byte(orchestratingMultipleGrpc), rulego.WithConfig(config))
	if err != nil {
		panic(err)
	}

	metaData1 := types.NewMetadata()
	// 将 http 请求的数据转化为 json 传入
	msg1 := types.NewMsg(time.Now().Unix(), "TEST_MSG_TYPE1", types.JSON, metaData1, `{"name": "RULEGO"}`)

	ret := make(map[string]interface{})
	ruleEngine2.OnMsgAndWait(msg1, types.WithEndFunc(func(msg types.RuleMsg, err error) {
		if err != nil {
			fmt.Println(err)
			return
		}
		if err := json.Unmarshal([]byte(msg.Data), &ret); err != nil {
			fmt.Println(err)
			return
		}
		for k, v := range msg.Metadata.Values() {
			ret[k] = v
		}
		fmt.Printf("rule run end, msg: %+v\n\n", msg)
	}))

	fmt.Println("组合接口返回值：", ret)

	time.Sleep(time.Second * 5)
}

var grpcHttpProxyFile = `
{
  "ruleChain": {
	"id":"grpcHttpProxy",
    "name": "http代理",
    "root": true
  },
  "metadata": {
    "nodes": [
       {
        "id": "s1",
        "type": "jsTransform",
        "name": "转换 helloworld.SayHello 入参",
        "configuration": {
          "jsScript": "let bool = !!msg['is_login']; msg['is_login']=bool; return {'msg':msg,'metadata':metadata,'msgType':msgType};"
        }
      },
      {
        "id": "s2",
        "type": "grpcCall",
        "name": "调用 helloworld.SayHello 接口",
        "configuration": {
          "reqType":  "helloworld.HelloRequest",
          "respType": "helloworld.HelloReply",
          "method":      "helloworld.Greeter/SayHello",
          "target":      "127.0.0.1:8088"
        }
      },
		{
        "id": "s3",
        "type": "jsTransform",
        "name": "转化出参",
        "configuration": {
          "jsScript": "let now = new Date(); msg['timestamp']=now.getTime(); return {'msg':msg,'metadata':metadata,'msgType':msgType};"
        }
      }
    ],
    "connections": [
      {
        "fromId": "s1",
        "toId": "s2",
        "type": "Success"
      },
      {
        "fromId": "s2",
        "toId": "s3",
        "type": "Success"
      }
    ],
    "ruleChainConnections": null
  }
}
`

var orchestratingMultipleGrpc = `
{
  "ruleChain": {
	"id":"orchestratingMultipleGrpc",
    "name": "编排多个 grpc",
    "root": true
  },
  "metadata": {
    "nodes": [
      {
        "id": "root_flow_node_01",
        "type": "flow",
        "name": "请求多个 grpc",
        "debugMode": true,
        "configuration": {
          "targetId": "callMultipleGrpc"
        }
      },
		  {
        "id": "root_s2",
        "type": "jsTransform",
        "name": "转化出参",
        "configuration": {
          "jsScript": "let now = new Date();let retrunMsg = {}; retrunMsg['timestamp'] = now.getTime();return {'msg':retrunMsg,'metadata':metadata,'msgType':msgType};"
        }
      }
    ],
    "connections": [
      {
        "fromId": "root_flow_node_01",
        "toId": "root_s2",
        "type": "Success"
      }
    ],
    "ruleChainConnections": null
  }
}
`

var callMultipleGrpc = `
{
  "ruleChain": {
	"id":"callMultipleGrpc",
    "name": "请求多个 grpc",
    "root": true
  },
  "metadata": {
    "nodes": [
      {
        "id": "s1",
        "type": "jsTransform",
        "name": "转换 helloworld.GetUserInfo 入参",
        "configuration": {
          "jsScript": "let now = new Date(); msg['start_time']=now.getTime();return {'msg':msg,'metadata':metadata,'msgType':msgType};"
        }
      },
      {
        "id": "s2",
        "type": "grpcCall",
        "name": "调用 helloworld.GetUserInfo 接口",
        "configuration": {
          "reqType":  "helloworld.GetUserInfoRequest",
          "respType": "helloworld.GetUserInfoReply",
          "method":      "helloworld.Greeter/GetUserInfo",
          "target":      "127.0.0.1:8088"
        }
      },
      {
        "id": "s3",
        "type": "grpcCall",
        "name": "调用 helloworld.GetWalletInfo 接口",
        "configuration": {
          "reqType":  "helloworld.GetWalletInfoRequest",
          "respType": "helloworld.GetWalletInfoReply",
          "method":      "helloworld.Greeter/GetWalletInfo",
          "target":      "127.0.0.1:8088"
        }
      }
    ],
    "connections": [
      {
        "fromId": "s1",
        "toId": "s2",
        "type": "Success"
      },
      {
        "fromId": "s1",
        "toId": "s3",
        "type": "Success"
      }
    ],
    "ruleChainConnections": null
  }
}
`
