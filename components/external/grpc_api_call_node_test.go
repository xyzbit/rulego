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
	"context"
	"testing"

	"github.com/xyzbit/rulego/api/types"
	"github.com/xyzbit/rulego/test"
	"github.com/xyzbit/rulego/test/assert"
	"github.com/xyzbit/rulego/testdata/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"
)

type MockGRPCConn struct {
	t *testing.T
}

func (i *MockGRPCConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	assert.Equal(i.t, "welfare.v1.Welfare/GetTaskList", method)

	msg, ok := args.(protoiface.MessageV1)
	assert.True(i.t, ok)
	assert.Equal(i.t, msg.String(), `project:"test" user_status:{uid:"481739124807512917" is_login:true}`)

	resp := reply.(*pb.GetTaskListResp)
	resp.Tasks = []*pb.Task{
		{Id: 1, Key: "key1", Name: "task1"},
	}
	return nil
}

func (i *MockGRPCConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	i.t.Logf("method: %s", method)
	return nil, nil
}

func (i *MockGRPCConn) Close() error {
	return nil
}

func TestRPCApiCallNodeOnMsg(t *testing.T) {
	// 需要自动生成的代码
	_ = &pb.GetTaskListReq{}

	node := &RPCCallNode{
		Config: RPCCallNodeConfiguration{
			ServiceName: "welfare",
			Method:      "welfare.v1.Welfare/GetTaskList",
			Target:      "127.0.0.1:9098",
		},
		gconn: &MockGRPCConn{t: t},
	}

	config := types.NewConfig()
	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string) {
		assert.Equal(t, msg.Data, `{"tasks":[{"id":1,"key":"key1","name":"task1"}]}`)
	})

	metaData := types.BuildMetadata(make(map[string]string))
	metaData.PutValue("is_login", "true")
	metaData.PutValue("_req_type", "welfare.v1.GetTaskListReq")
	metaData.PutValue("_reply_type", "welfare.v1.GetTaskListResp")

	msg := ctx.NewMsg("PB_MSG", metaData, `{
		"project": "test",
		"user_status": {
			"uid": "481739124807512917",
			"is_login": ${is_login}
		}
	}`)
	err := node.OnMsg(ctx, msg)
	if err != nil {
		t.Errorf("err=%s", err)
	}
}
