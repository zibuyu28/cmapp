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

//// Param keys from env
//const (
//	coreWSAddr  string = "CORE_WS_ADDR"
//	agentUnique string = "AGENT_UNIQUE"
//)
//
//type wsFront struct {
//	cli         *ws.Client
//	coreWSAddr  string
//	agentUnique string
//}
//
//func wsClientIns(ctx context.Context) (*wsFront, error) {
//	// 获取到core的地址，链接
//	wsFlag, ok := Flags[coreWSAddr]
//	if !ok {
//		return nil, errors.New("fail to get [CORE_WS_ADDR] from env")
//	}
//	wsaddr := wsFlag.Value
//	if len(wsaddr) == 0 {
//		return nil, errors.New("env value of [CORE_WS_ADDR] is empty")
//	}
//
//	agFlag, ok := Flags[agentUnique]
//
//	if !ok {
//		return nil, errors.New("fail to get [AGENT_UNIQUE] from env")
//	}
//	unique := agFlag.Value
//	if len(unique) == 0 {
//		return nil, errors.New("env value of [AGENT_UNIQUE] is empty")
//	}
//
//	client := ws.NewClient(unique, wsaddr)
//	err := client.Connect(ctx)
//	if err != nil {
//		return nil, errors.Wrapf(err, "connect to core, addr is [%s]", wsaddr)
//	}
//	log.Infof(ctx, "connect to core [%s] success", wsaddr)
//
//	return &wsFront{
//		cli:         client,
//		coreWSAddr:  wsaddr,
//		agentUnique: unique,
//	}, nil
//}
//
//func (wf *wsFront) run(ctx context.Context) {
//	go wf.cli.ReadPump()
//	go wf.cli.WritePump()
//}
//
//func (wf *wsFront) ReceiveAction(req *ActionREQ) error {
//
//	messageStr := <-wf.cli.Receive
//
//	err := json.Unmarshal(messageStr, req)
//	if err != nil {
//		return errors.Wrapf(err, "unmarshal message [%s] to action request", string(messageStr))
//	}
//	return nil
//}
//
//func (wf *wsFront) ActionResp(ar ActionRESP) error {
//	marshal, err := json.Marshal(ar)
//	if err != nil {
//		return errors.Wrapf(err, "marshal aciton resp [%+v] to json", ar)
//	}
//	wf.cli.Send <- marshal
//	return nil
//}
