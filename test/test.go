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

package test

import (
	"context"
	"time"

	"github.com/xyzbit/rulego/api/types"
)

// NodeTestRuleContext
// 只为测试单节点，临时创建的上下文
// 无法把多个节点组成链式
// callback 回调处理结果
type NodeTestRuleContext struct {
	config   types.Config
	context  context.Context
	callback func(msg types.RuleMsg, relationType string)
	self     types.Node
}

func NewRuleContext(config types.Config, callback func(msg types.RuleMsg, relationType string)) types.RuleContext {
	return &NodeTestRuleContext{
		config:   config,
		callback: callback,
		context:  context.TODO(),
	}
}

func NewRuleContextFull(config types.Config, self types.Node, callback func(msg types.RuleMsg, relationType string)) types.RuleContext {
	return &NodeTestRuleContext{
		config:   config,
		self:     self,
		callback: callback,
		context:  context.TODO(),
	}
}

func (ctx *NodeTestRuleContext) TellSuccess(msg types.RuleMsg) {
	ctx.callback(msg, types.Success)
}

func (ctx *NodeTestRuleContext) TellFailure(msg types.RuleMsg, err error) {
	ctx.callback(msg, types.Failure)
}

func (ctx *NodeTestRuleContext) TellNext(msg types.RuleMsg, relationTypes ...string) {
	for _, relationType := range relationTypes {
		ctx.callback(msg, relationType)
	}
}

func (ctx *NodeTestRuleContext) TellSelf(msg types.RuleMsg, delayMs int64) {
	time.AfterFunc(time.Millisecond*time.Duration(delayMs), func() {
		if ctx.self != nil {
			ctx.self.OnMsg(ctx, msg)
		}
	})
}

func (ctx *NodeTestRuleContext) NewMsg(msgType string, metaData types.Metadata, data string) types.RuleMsg {
	return types.NewMsg(0, msgType, types.JSON, metaData, data)
}

func (ctx *NodeTestRuleContext) GetSelfId() string {
	return ""
}

func (ctx *NodeTestRuleContext) Config() types.Config {
	return ctx.config
}

func (ctx *NodeTestRuleContext) SubmitTack(task func()) {
	go task()
}

func (ctx *NodeTestRuleContext) SetEndFunc(onEndFunc func(msg types.RuleMsg, err error)) types.RuleContext {
	return ctx
}

func (ctx *NodeTestRuleContext) GetEndFunc() func(msg types.RuleMsg, err error) {
	return nil
}

func (ctx *NodeTestRuleContext) SetContext(c context.Context) types.RuleContext {
	ctx.context = c
	return ctx
}

func (ctx *NodeTestRuleContext) GetContext() context.Context {
	return ctx.context
}
