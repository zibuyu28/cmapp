package kubeclient

import (
	"context"
	"errors"
	log "git.hyperchain.cn/blocface/golog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

//CreatePersistentVolume .
func (c *Clients) CreatePersistentVolume(pv *corev1.PersistentVolume) error {

	newpv, err := c.KubeClient.CoreV1().PersistentVolumes().Create(pv)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	log.Infof("Created PersistentVolume %q \n", newpv.GetObjectMeta().GetName())
	return nil
}

//DeletePersistentVolume .
func (c *Clients) DeletePersistentVolume(pv *corev1.PersistentVolume, ops *metav1.DeleteOptions) error {

	err := c.KubeClient.CoreV1().PersistentVolumes().Delete(pv.Name, ops)
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("Delete PersistentVolume %q \n", pv.GetObjectMeta().GetName())
	return err
}

//GetPersistentVolumeByName .
func (c *Clients) GetPersistentVolumeByName(name string, options metav1.GetOptions) (*corev1.PersistentVolume, error) {
	newpv, err := c.KubeClient.CoreV1().PersistentVolumes().Get(name,options)
	if err != nil {
		return nil, err
	}
	return newpv, nil
}


// CheckPV check pv exist
func (c *Clients) CheckPV(pv *corev1.PersistentVolume) (r bool, err error) {
	claims := c.KubeClient.CoreV1().PersistentVolumes()
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			pvc, err := claims.Get(pv.GetObjectMeta().GetName(), metav1.GetOptions{})
			if err != nil {
				errChan <- err
				return
			}
			if pvc.Status.Phase == corev1.VolumeBound {
				result <- true
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				log.Infof("Check pvc status...")
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
		return false, errors.New("pv create after 100 second timeout")
	}
}
