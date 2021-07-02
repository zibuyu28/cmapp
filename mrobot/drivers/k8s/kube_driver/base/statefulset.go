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

//GetStatefulSetByName .
func (c *Client) GetStatefulSetByName(name, namespace string, ops metav1.GetOptions) (*appsv1.StatefulSet, error) {
	sfsClient := c.k.AppsV1().StatefulSets(namespace)
	redep, err := sfsClient.Get(c.ctx, name, ops)
	if err != nil {
		return nil, err
	}
	return redep, nil
}

// CreateStatefulSet .
func (c *Client) CreateStatefulSet(sfs *appsv1.StatefulSet) error {
	if sfs.Namespace == "" {
		sfs.Namespace = corev1.NamespaceDefault
	}
	sfsClient := c.k.AppsV1().StatefulSets(sfs.Namespace)
	// Create Service
	old, err := sfsClient.Get(c.ctx, sfs.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_, err := sfsClient.Create(c.ctx, sfs, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrapf(err, "create stateful set [%s]", sfs.Name)
			}
		} else {
			return errors.Wrapf(err, "get stateful set [%s]", sfs.Name)
		}
	} else if old != nil {
		_, err = sfsClient.Update(c.ctx, sfs, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "update stateful set [%s]", sfs.Name)
		}
	}
	_, err = c.CheckStatefulSet(sfs)
	if err != nil {
		return errors.Wrapf(err, "check stateful set [%s]", sfs.Name)

	}
	return nil
}

// CheckStatefulSet check stateful set
func (c *Client) CheckStatefulSet(sfs *appsv1.StatefulSet) (r bool, err error) {
	sets := c.k.AppsV1().StatefulSets(sfs.GetObjectMeta().GetNamespace())
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			s, err := sets.Get(c.ctx, sfs.GetObjectMeta().GetName(), metav1.GetOptions{})
			if err != nil {
				errChan <- err
				return
			}
			if s.Status.CurrentReplicas == *s.Spec.Replicas {
				result <- true
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				log.Infof(ctx, "stateful set creating...")
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
		return false, errors.New("stateful set create failed after 300 second timeout")
	case <-c.ctx.Done():
		return false, errors.New("stateful set create state unknown with context deadline")
	}
}

// DeleteStatefulSet .
func (c *Client) DeleteStatefulSet(sfs *appsv1.StatefulSet, ops metav1.DeleteOptions) error {
	if sfs.Namespace == "" {
		sfs.Namespace = corev1.NamespaceDefault
	}
	sfsClient := c.k.AppsV1().StatefulSets(sfs.Namespace)
	err := sfsClient.Delete(c.ctx, sfs.Name, ops)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return errors.Wrapf(err, "delete stateful set [%s]", sfs.Name)
	}

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			_, err := sfsClient.Get(ctx, sfs.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof(ctx, "stateful set deleting...")
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
		return errors.New("stateful set delete failed after 300 second timeout")
	case <-c.ctx.Done():
		return errors.New("stateful set delete state unknown with context deadline")
	}
}
