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
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreateDeployment .
func (c *Client) CreateDeployment(dep *appsv1.Deployment) error {
	if dep.Namespace == "" {
		dep.Namespace = corev1.NamespaceDefault
	}
	deploymentsClient := c.k.AppsV1().Deployments(dep.Namespace)

	var newDep *appsv1.Deployment
	old, err := deploymentsClient.Get(c.ctx, dep.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			n, err := deploymentsClient.Create(c.ctx, dep, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrap(err, "create deployment")
			}
			newDep = n
		} else {
			return errors.Wrap(err, "get deployment")
		}
	} else if old != nil {
		newDep, err = deploymentsClient.Update(c.ctx, dep, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrap(err, "update deployment")
		}
	}
	// need to check pod
	_, err = c.CheckDeployment(newDep)
	if err != nil {
		return errors.Wrap(err, "check deployment state finish")
	}
	return nil
}

// CheckDeployment check deployment exist
func (c *Client) CheckDeployment(dep *appsv1.Deployment) (r bool, err error) {
	deploymentsClient := c.k.AppsV1().Deployments(dep.GetObjectMeta().GetNamespace())
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			d, err := deploymentsClient.Get(ctx, dep.GetObjectMeta().GetName(), metav1.GetOptions{})
			if err != nil {
				errChan <- err
				return
			}
			if d.Status.AvailableReplicas == *d.Spec.Replicas {
				result <- true
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				log.Infof(ctx, "deployment creating...")
			}
			time.Sleep(time.Second * 5)
		}
	}(ctx)
	select {
	case err := <-errChan:
		return false, err
	case r = <-result:
		return r, nil
	case <-time.After(time.Duration(300) * time.Second):
		return false, errors.New("deployment create failed after 300 second timeout")
	case <-c.ctx.Done():
		return false, errors.New("deployment create failed with context deadline")
	}
}

//DeleteDeployment .
func (c *Client) DeleteDeployment(dep *appsv1.Deployment, ops metav1.DeleteOptions) error {
	deploymentsClient := c.k.AppsV1().Deployments(dep.Namespace)
	err := deploymentsClient.Delete(c.ctx, dep.Name, ops)
	if err != nil {
		return errors.Wrapf(err, "delete deployment [%s]", dep.Name)
	}
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			_, err := deploymentsClient.Get(ctx, dep.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof(ctx, "deployment deleting...")
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
		return errors.New("deployment deleted check failed after 300 second timeout")
	case <-c.ctx.Done():
		return errors.New("deployment delete state unknown with context deadline")
	}
}

//UpdateDeployment .
func (c *Client) UpdateDeployment(dep *appsv1.Deployment) error {
	deploymentsClient := c.k.AppsV1().Deployments(dep.Namespace)
	old, err := deploymentsClient.Get(c.ctx, dep.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "get deployment [%s]", dep.Name)
	}
	if old == nil || old.ResourceVersion == "" {
		return errors.Errorf("cant get deployment [%s] present ResourceVersion", dep.Name)
	}
	dep.ResourceVersion = old.ResourceVersion
	_, err = deploymentsClient.Update(c.ctx, dep, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "update deployment")
	}
	return nil
}
