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
	"strings"
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
	assert.Equal(i.t, "helloworld.Greeter/SayHello", method)

	msg, ok := args.(protoiface.MessageV1)
	assert.True(i.t, ok)
	assert.True(i.t, strings.Contains(msg.String(), `name:"RULEGO"`))
	assert.True(i.t, strings.Contains(msg.String(), `is_login:true`))

	resp, ok := reply.(*pb.HelloReply)
	assert.True(i.t, ok)
	if ok {
		resp.Code = 200
		resp.Message = "Hello RULEGO, your login status: true"
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
	node := &RPCCallNode{
		Config: RPCCallNodeConfiguration{
			ReqType:  "helloworld.HelloRequest",
			RespType: "helloworld.HelloReply",
			Method:   "helloworld.Greeter/SayHello",
			Target:   "127.0.0.1:8088",
		},
		gconn: &MockGRPCConn{t: t},
	}

	config := types.NewConfig()
	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string) {
		assert.Equal(t, msg.Data, `{"code":200,"message":"Hello RULEGO, your login status: true"}`)
	})

	metaData := types.BuildMetadata(make(map[string]string))
	metaData.PutValue("is_login", "true")

	msg := ctx.NewMsg("PB_MSG", metaData, `{
		 "name": "RULEGO",
		 "is_login": ${is_login}
	 }`)
	node.OnMsg(ctx, msg)
}
