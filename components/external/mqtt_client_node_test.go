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
	"github.com/xyzbit/rulego/test/assert"
)

func TestMqttClientNodeOnMsg(t *testing.T) {
	var node MqttClientNode
	configuration := make(types.Configuration)
	configuration["Server"] = "127.0.0.1:1883"
	configuration["Topic"] = "/device/msg"
	config := types.NewConfig()
	err := node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}
	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string) {
		assert.Equal(t, types.Success, relationType)
	})
	metaData := types.BuildMetadata(make(map[string]string))
	msg := ctx.NewMsg("TEST_MSG_TYPE_AA", metaData, "{\"test\":\"AA\"}")
	err = node.OnMsg(ctx, msg)
	if err != nil {
		t.Errorf("err=%s", err)
	}
}
