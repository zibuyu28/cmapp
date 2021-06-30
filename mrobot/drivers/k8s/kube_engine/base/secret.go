package kubeclient

import (
	"context"
	"errors"
	log "git.hyperchain.cn/blocface/golog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

//GetSecretByName .
func (c *Clients) GetSecretByName(namespace, name string, ops metav1.GetOptions) (*corev1.Secret, error) {
	return c.KubeClient.CoreV1().Secrets(namespace).Get(name, ops)
}

func (c *Clients)DeleteSecret(se *corev1.Secret,ops *metav1.DeleteOptions)error{
	serectsClient := c.KubeClient.CoreV1().Secrets(se.Namespace)
	err := serectsClient.Delete(se.Name, ops)
	if err != nil {
		log.Errorf(err.Error())
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	log.Infof("Delete secret %q", se.Name)

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			_, err := serectsClient.Get(se.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof("Check secret deleted...")
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
		return errors.New("secret deleted check failed after 240 second timeout")
	}
}