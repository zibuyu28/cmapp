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

//GetDeployment .
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
