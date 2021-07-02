package base

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

//GetSecretByName .
func (c *Client) GetSecretByName(namespace, name string, ops metav1.GetOptions) (*corev1.Secret, error) {
	return c.k.CoreV1().Secrets(namespace).Get(c.ctx, name, ops)
}

func (c *Client) DeleteSecret(se *corev1.Secret, ops metav1.DeleteOptions) error {
	serectsClient := c.k.CoreV1().Secrets(se.Namespace)
	err := serectsClient.Delete(c.ctx, se.Name, ops)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return errors.Wrapf(err, "delete secret [%s]", se.Name)
	}
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			_, err := serectsClient.Get(ctx, se.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof(ctx, "secret deleting...")
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
		return errors.New("secret deleted check failed after 300 second timeout")
	case <-c.ctx.Done():
		return errors.New("secret deleted state unknown with context deadline")
	}
}
