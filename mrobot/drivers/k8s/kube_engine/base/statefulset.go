package kubeclient

import (
	"context"
	"errors"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetStatefulSetByName .
func (c *Clients) GetStatefulSetByName(name, namespace string, ops metav1.GetOptions) (*appsv1.StatefulSet, error) {
	sfsClient := c.KubeClient.AppsV1().StatefulSets(namespace)
	redep, err := sfsClient.Get(name, ops)
	if err != nil {
		return nil, err
	}
	return redep, nil
}

// CreateStatefulSet .
func (c *Clients) CreateStatefulSet(sfs *appsv1.StatefulSet) error {
	if sfs.Namespace == "" {
		sfs.Namespace = corev1.NamespaceDefault
	}
	sfsClient := c.KubeClient.AppsV1().StatefulSets(sfs.Namespace)
	var newSfs *appsv1.StatefulSet
	// Create Service
	old, err := sfsClient.Get(sfs.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			n, err := sfsClient.Create(sfs)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			newSfs = n
		} else {
			log.Error(err.Error())
			return err
		}
	} else if old != nil {
		newSfs, err = sfsClient.Update(sfs)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	_, err = c.CheckStatefulSet(sfs)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	log.Infof("Created deployment %q \n", newSfs.GetObjectMeta().GetName())
	return nil
}

// checkStatefulSet check stateful set
func (c *Clients) CheckStatefulSet(sfs *appsv1.StatefulSet) (r bool, err error) {
	sets := c.KubeClient.AppsV1().StatefulSets(sfs.GetObjectMeta().GetNamespace())
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			s, err := sets.Get(sfs.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof("Check stateful set status...")
			}
			time.Sleep(time.Second * 1)
		}
	}(ctx)
	select {
	case err := <-errChan:
		return false, err
	case r = <-result:
		return r, nil
	case <-time.After(time.Duration(180) * time.Second):
		cancelFunc()
		return false, errors.New("stateful set create failed after 100 second timeout")
	}
}

// DeleteStatefulSet .
func (c *Clients) DeleteStatefulSet(sfs *appsv1.StatefulSet, ops *metav1.DeleteOptions) error {
	if sfs.Namespace == "" {
		sfs.Namespace = corev1.NamespaceDefault
	}
	sfsClient := c.KubeClient.AppsV1().StatefulSets(sfs.Namespace)
	err := sfsClient.Delete(sfs.Name, ops)
	if err != nil {
		log.Errorf(err.Error())
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	log.Infof("Created deployment %q ", sfs.GetObjectMeta().GetName())

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			_, err := sfsClient.Get(sfs.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof("Check stateful set status...")
			}
			time.Sleep(time.Second * 1)
		}
	}(ctx)
	select {
	case err := <-errChan:
		return err
	case <-result:
		return nil
	case <-time.After(time.Duration(240) * time.Second):
		cancelFunc()
		return errors.New("stateful set delete failed after 24o second timeout")
	}
}
