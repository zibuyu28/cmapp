/*
 * Copyright © 2021 zibuyu28
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package worker

import (
	"context"
	"github.com/zibuyu28/cmapp/common/log"
	"os"
	"strings"
)

const driverPrefix string = "DRIAGENT_"

type Flag struct {
	Name  string
	Value string
}

var Flags = make(map[string]Flag)

func init() {
	environ := os.Environ()
	for _, s := range environ {
		if strings.HasPrefix(s, driverPrefix) {
			kvs := strings.SplitN(strings.TrimPrefix(s, driverPrefix), "=", 2)
			if len(kvs) != 2 {
				continue
			}
			Flags[kvs[0]] = Flag{
				Name:  kvs[0],
				Value: kvs[1],
			}
		}
	}
}

//// agCmd represents the ag command
//var AgCmd = &cobra.Command{
//	Use:   "startag",
//	Short: "ag machine agent",
//	Long: `A longer description that spans multiple lines and likely contains examples
//and usage of using your command. For example:
//
//Cobra is a CLI library for Go that empowers applications.
//This application is a tool to generate the needed files
//to quickly create a Cobra application.`,
//	Run: func(cmd *cobra.Command, args []string) {
//		timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
//		defer cancelFunc()
//		// 有很多信息在环境变量中
//		// 启动 wsclient 链接到 core，开始接收信息
//		// 调用自身二进制，启动ag
//		start(timeout)
//	},
//}

func Start(ctx context.Context) {
	wscli, err := wsClientIns(ctx)
	if err != nil {
		log.Fatalf(ctx, "Currently fail to new ws client. Err: [%v]", err)
	}

	plg, err := pluginIns(ctx)
	if err != nil {
		log.Fatalf(ctx, "Currently fail to new plugin. Err: [%v]", err)
	}

	b := broker{
		wsFont: wscli,
		plg:    plg,
	}
	b.Execute(ctx)
}
