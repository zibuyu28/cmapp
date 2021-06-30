package kubeclient

import (
	"context"
	"errors"
	log "git.hyperchain.cn/blocface/golog"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

//GetDeployment .
func (c *Client) GetIngressByName(name, namespace string, ops metav1.GetOptions) (*v1beta1.Ingress, error) {
	ingressClient := c.KubeClient.ExtensionsV1beta1().Ingresses(namespace)
	redep, err := ingressClient.Get(name, ops)
	if err != nil {
		return nil, err
	}
	return redep, nil
}


func(c *Client) DeleteIngress(ingress *v1beta1.Ingress ,ops *metav1.DeleteOptions)error{
	ingressClient := c.KubeClient.CoreV1().PersistentVolumeClaims(ingress.Namespace)
	err := ingressClient.Delete(ingress.Name, ops)
	if err != nil {
		log.Errorf(err.Error())
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	log.Infof("Delete Ingress %q", ingress.GetObjectMeta().GetName())

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			_, err := ingressClient.Get(ingress.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof("Check ingress status...")
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
		return errors.New("ingress check delete after 240 second timeout")
	}
}
