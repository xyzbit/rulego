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

package transform

import (
	"testing"

	"github.com/xyzbit/rulego/api/types"
	"github.com/xyzbit/rulego/test"
	"github.com/xyzbit/rulego/test/assert"
)

func TestJsTransformNodeOnMsg(t *testing.T) {
	var node JsTransformNode
	configuration := make(types.Configuration)
	configuration["jsScript"] = `
		metadata['test']='test02';
		metadata['index']=52;
		msgType='TEST_MSG_TYPE2';
		var msg2={};
		msg2['bb']=22
		return {'msg':msg2,'metadata':metadata,'msgType':msgType};
  	`
	config := types.NewConfig()
	err := node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}
	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string) {
		assert.Equal(t, "{\"bb\":22}", msg.Data)
		assert.Equal(t, "TEST_MSG_TYPE2", msg.Type)
		assert.Equal(t, types.Success, relationType)
	})
	metaData := types.BuildMetadata(make(map[string]string))
	msg := ctx.NewMsg("TEST_MSG_TYPE_AA", metaData, "AA")
	err = node.OnMsg(ctx, msg)
	if err != nil {
		t.Errorf("err=%s", err)
	}
}
