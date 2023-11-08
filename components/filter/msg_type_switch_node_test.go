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

package filter

import (
	"testing"

	"github.com/xyzbit/rulego/api/types"
	"github.com/xyzbit/rulego/test"
	"github.com/xyzbit/rulego/test/assert"
)

func TestMsgTypeSwitchNodeOnMsg(t *testing.T) {
	var node MsgTypeSwitchNode
	config := types.NewConfig()
	err := node.Init(config, nil)
	if err != nil {
		t.Errorf("err=%s", err)
	}
	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string) {
		if msg.Type == "ACTIVITY_EVENT" {
			assert.Equal(t, "ACTIVITY_EVENT", relationType)
		} else if msg.Type == "INACTIVITY_EVENT" {
			assert.Equal(t, "INACTIVITY_EVENT", relationType)
		}
	})
	metaData := types.BuildMetadata(make(map[string]string))
	msg := ctx.NewMsg("ACTIVITY_EVENT", metaData, "AA")
	err = node.OnMsg(ctx, msg)
	if err != nil {
		t.Errorf("err=%s", err)
	}

	msg2 := ctx.NewMsg("INACTIVITY_EVENT", metaData, "BB")
	err = node.OnMsg(ctx, msg2)
	if err != nil {
		t.Errorf("err=%s", err)
	}
}
