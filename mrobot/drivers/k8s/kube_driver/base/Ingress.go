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

package base

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

// CreateIngress create ingress
func (c *Client) CreateIngress(ingress *v1beta1.Ingress) error {
	ingressClient := c.k.ExtensionsV1beta1().Ingresses(ingress.Namespace)
	_, err := ingressClient.Create(c.ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "create ingress")
	}
	return nil
}

//GetIngressByName .
func (c *Client) GetIngressByName(name, namespace string, ops metav1.GetOptions) (*v1beta1.Ingress, error) {
	ingressClient := c.k.ExtensionsV1beta1().Ingresses(namespace)
	redep, err := ingressClient.Get(c.ctx, name, ops)
	if err != nil {
		return nil, err
	}
	return redep, nil
}

func (c *Client) DeleteIngress(ingress *v1beta1.Ingress, ops metav1.DeleteOptions) error {
	ingressClient := c.k.CoreV1().PersistentVolumeClaims(ingress.Namespace)
	err := ingressClient.Delete(c.ctx, ingress.Name, ops)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return errors.Wrapf(err, "delete ingress [%s]", ingress.Name)
	}

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			_, err := ingressClient.Get(ctx, ingress.GetObjectMeta().GetName(), metav1.GetOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					result <- true
					return
				}
				errChan <- err
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				log.Infof(ctx, "ingress deleting...")
			}
			time.Sleep(time.Second * 5)
		}
	}(ctx)
	select {
	case err := <-errChan:
		return err
	case <-result:
		return nil
	case <-time.After(time.Duration(300) * time.Second):
		return errors.New("ingress check delete after 240 second timeout")
	case <-c.ctx.Done():
		return errors.New("ingress delete state unknown with context deadline")
	}
}
