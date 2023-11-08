/*
 * Copyright 2023 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"time"

	"github.com/xyzbit/rulego"
	"github.com/xyzbit/rulego/api/types"
)

// 测试使用占位符替换配置
func main() {
	config := rulego.NewConfig()
	// 设置全局属性参数，通过${global.transformJs} 方式替换内容
	// 节点初始化时候替换,只替换一次
	config.Properties.PutValue("transformJs", `
		var value=global.globalValue;
		msg['addField2']=value;
		msg['addValue']=add(1,5); 
		msgType=handleMsg(msg,metadata,msgType);
		return {'msg':msg,'metadata':metadata,'msgType':msgType};
	`)
	// 在js脚本运行时获取全局变量：global.xx
	config.Properties.PutValue("globalValue", "addValueFromConfig")

	// 注册自定义函数
	config.RegisterUdf("add", func(a, b int) int {
		return a + b
	})
	config.RegisterUdf("handleMsg", func(msg map[string]interface{}, metadata map[string]string, msgType string) string {
		msg["returnFromGo"] = "returnFromGo"
		_, ok := rulego.Get("aa")
		msg["hasAaRuleChain"] = ok
		return "returnFromGoMsgType"
	})
	// 元数据
	metaData := types.NewMetadata()
	// 通过${url}替换内容
	// 运行时替换
	metaData.PutValue("postUrl", "http://127.0.0.1:8080/api/msg")

	// 处理数据
	ruleEngine, err := rulego.New("rule01", []byte(chainJsonFile), rulego.WithConfig(config))
	if err != nil {
		panic(err)
	}

	msg := types.NewMsg(0, "TEST_MSG_TYPE1", types.JSON, metaData, "{\"temperature\":41}")
	ruleEngine.OnMsgWithOptions(msg, types.WithEndFunc(func(msg types.RuleMsg, err error) {
		// 得到规则链处理结果
		fmt.Println("第一次执行", msg, err)
	}))

	time.Sleep(time.Second * 5)
	// 第二次执行
	// 元数据
	metaData = types.NewMetadata()
	// 通过${url}替换内容
	// 运行时替换
	metaData.PutValue("postUrl", "http://127.0.0.1:8080/api/msg2")
	msg = types.NewMsg(0, "TEST_MSG_TYPE1", types.JSON, metaData, "{\"temperature\":42}")
	ruleEngine.OnMsgWithOptions(msg, types.WithEndFunc(func(msg types.RuleMsg, err error) {
		// 得到规则链处理结果
		fmt.Println("第二次执行", msg, err)
	}))
	time.Sleep(time.Second * 30)
}

var chainJsonFile = `
{
  "ruleChain": {
	"id":"rule01",
    "name": "测试规则链",
    "root": true
  },
  "metadata": {
    "nodes": [
       {
        "id": "s1",
        "type": "jsTransform",
        "name": "转换",
        "configuration": {
          "jsScript": "${global.transformJs}"
        }
      },
      {
        "id": "s2",
        "type": "restApiCall",
        "name": "调用restApi增强数据",
        "configuration": {
          "restEndpointUrlPattern": "${postUrl}",
          "requestMethod": "POST",
          "maxParallelRequestsCount": 200
        }
      },
      {
        "id": "s4",
        "type": "log",
        "name": "记录响应日志",
        "configuration": {
          "jsScript": "return '响应\\n Incoming message:\\n' + JSON.stringify(msg) + '\\nIncoming metadata:\\n' + JSON.stringify(metadata);"
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
        "toId": "s4",
        "type": "Success"
      },
		{
        "fromId": "s2",
        "toId": "s4",
        "type": "Failure"
      }
    ],
    "ruleChainConnections": null
  }
}
`
