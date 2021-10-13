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
	"fmt"
	"github.com/zibuyu28/cmapp/common/log"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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

var AGDefaultHealthPort = 9009

func healthFunc(ctx context.Context, muxs []*http.ServeMux) {
	var mux *http.ServeMux
	if len(muxs) == 0 {
		mux = http.NewServeMux()
	} else {
		mux = muxs[0]
	}
	mux.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(writer, "ok")
	})
	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", AGDefaultHealthPort),
		Handler: mux,
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf(ctx, "Currently fail to listen on port [9009]. Err: [%v]", err)
	}
}

func Start(ctx context.Context, muxs ...*http.ServeMux) {
	go healthFunc(ctx, muxs)
	//wscli, err := wsClientIns(ctx)
	//if err != nil {
	//	log.Fatalf(ctx, "Currently fail to new ws client. Err: [%v]", err)
	//}

	_, err := pluginIns(ctx)
	if err != nil {
		log.Fatalf(ctx, "Currently fail to new plugin. Err: [%v]", err)
	}

	// block
	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		// TODO app reload
		default:
			return
		}
	}

	//b := broker{
	//	wsFont: wscli,
	//	plg:    plg,
	//}
	//b.Execute(ctx)
}
