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

package external

import (
	"testing"

	"github.com/xyzbit/rulego/api/types"
	"github.com/xyzbit/rulego/test"
	"github.com/xyzbit/rulego/testdata/pb"
)

func TestRPCApiCallNodeOnMsg(t *testing.T) {
	// 需要自动生成的代码
	_ = &pb.GetTaskListReq{}

	var node RPCCallNode
	configuration := make(types.Configuration)
	configuration["serviceName"] = "welfare"
	configuration["method"] = "welfare.v1.Welfare/GetTaskList"
	configuration["target"] = "127.0.0.1:9098"

	config := types.NewConfig()
	// config.OnDebug = func(flowType, nodeId string, msg types.RuleMsg, relationType string, err error) {
	// 	t.Logf("flowType=%s, nodeId=%s, msg=%+v, relationType=%s, err=%s", flowType, nodeId, msg, relationType, err)
	// }
	// config.OnEnd = func(msg types.RuleMsg, err error) {
	// 	t.Logf("end msg=%+v, err=%s", msg, err)
	// }

	err := node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}

	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string) {
		t.Logf("rule context %+v, relationType: %s", msg, relationType)
	})
	metaData := types.BuildMetadata(make(map[string]string))
	metaData.PutValue("is_login", "true")
	metaData.PutValue("_req_type", "welfare.v1.GetTaskListReq")
	metaData.PutValue("_reply_type", "welfare.v1.GetTaskListResp")

	msg := ctx.NewMsg("PB_MSG", metaData, `{
		"project": "newmedia_rapidapp",
		"user_status": {
			"uid": "481739124807512917",
			"is_login": ${is_login}
		}
	}`)
	err = node.OnMsg(ctx, msg)
	if err != nil {
		t.Errorf("err=%s", err)
	}
}
