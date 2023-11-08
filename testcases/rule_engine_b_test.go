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

package testcases

import (
	"testing"

	"github.com/xyzbit/rulego"
	"github.com/xyzbit/rulego/api/types"
	"github.com/xyzbit/rulego/utils/str"
)

func BenchmarkChainNotChangeMetadata(b *testing.B) {
	b.ResetTimer()
	config := rulego.NewConfig()
	ruleEngine, err := rulego.New(str.RandomStr(10), []byte(ruleChainFile), rulego.WithConfig(config))
	if err != nil {
		b.Error(err)
	}
	for i := 0; i < b.N; i++ {
		metaData := types.NewMetadata()
		metaData.PutValue("productType", "test01")
		msg := types.NewMsg(0, "TEST_MSG_TYPE", types.JSON, metaData, "{\"temperature\":35}")
		ruleEngine.OnMsg(msg)
	}
}

func BenchmarkChainChangeMetadataAndMsg(b *testing.B) {
	config := rulego.NewConfig()

	ruleEngine, err := rulego.New(str.RandomStr(10), []byte(ruleChainFile), rulego.WithConfig(config))
	if err != nil {
		b.Error(err)
	}
	// modify s1 node content
	_ = ruleEngine.ReloadChild("s2", []byte(modifyMetadataAndMsgNode))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metaData := types.NewMetadata()
		metaData.PutValue("productType", "test01")
		msg := types.NewMsg(0, "TEST_MSG_TYPE", types.JSON, metaData, "{\"temperature\":35}")
		ruleEngine.OnMsg(msg)
	}
}

func BenchmarkCallRestApiNodeGo(b *testing.B) {
	// 不使用协程池
	config := rulego.NewConfig()
	ruleEngine, _ := rulego.New(str.RandomStr(10), loadFile("./chain_call_rest_api.json"), rulego.WithConfig(config))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callRestApiNode(ruleEngine)
	}
}

func BenchmarkCallRestApiNodeWorkerPool(b *testing.B) {
	// 使用协程池
	config := rulego.NewConfig(types.WithDefaultPool())
	ruleEngine, _ := rulego.New(str.RandomStr(10), loadFile("./chain_call_rest_api.json"), rulego.WithConfig(config))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callRestApiNode(ruleEngine)
	}
}

//	func BenchmarkCallRestApiNodeAnts(b *testing.B) {
//		defaultAntsPool, _ := ants.NewPool(200000)
//		//使用协程池
//		config := rulego.NewConfig(types.WithPool(defaultAntsPool))
//		ruleEngine, _ := rulego.New(str.RandomStr(10), loadFile("./chain_call_rest_api.json"), rulego.WithConfig(config))
//		b.ResetTimer()
//		for i := 0; i < b.N; i++ {
//			callRestApiNode(ruleEngine)
//		}
//	}
func callRestApiNode(ruleEngine *rulego.RuleEngine) {
	metaData := types.NewMetadata()
	metaData.PutValue("productType", "test01")
	msg := types.NewMsg(0, "TEST_MSG_TYPE", types.JSON, metaData, "{\"aa\":\"aaaaaaaaaaaaaa\"}")
	ruleEngine.OnMsg(msg)
}
