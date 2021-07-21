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

package k8s

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	"sync"
)

type workRepository struct {
	rep sync.Map
}

// TODO: add other useful param
type work struct {
	UID        string
	init       *initcmd
	deployment string
	net        map[int]network
}

type network struct {
	Name     string
	Protocol string
	Internal string
	External string
}

type initcmd struct {
	m    sync.Mutex
	init []string
}

func (i *initcmd) append(cmd string) {
	i.m.Lock()
	defer i.m.Unlock()

	i.init = append(i.init, cmd)
}

func (w *workRepository) load(ctx context.Context) (*work, error) {
	uid, err := guid(ctx)
	if err != nil {
		return nil, errors.New("load uid from context")
	}
	wsp, ok := w.rep.Load(uid)
	if !ok {
		return nil, errors.Errorf("fail to get workspace from rep by uid [%s]", uid)
	}
	return wsp.(*work), nil
}

func (w *workRepository) new(ctx context.Context, wsp *work) error {
	uid, err := guid(ctx)
	if err != nil {
		return errors.New("load uid from context")
	}
	wsp.UID = uid
	w.rep.Store(uid, wsp)
	return nil
}

func guid(ctx context.Context) (uid string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = errors.New("fail to get metadata from context")
		return
	}
	uuid := md.Get("MA_UUID")
	if len(uuid) == 0 {
		err = errors.New("fail to get uuid from metadata")
		return
	}
	uid = uuid[0]
	return
}
