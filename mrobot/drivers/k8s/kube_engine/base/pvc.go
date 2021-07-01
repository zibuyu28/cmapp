package kubeclient

import (
	"context"
	"errors"
	"fmt"
	log "git.hyperchain.cn/blocface/golog"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreatePersistentVolumeClaim .
func (c *Clients) CreatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) error {
	pvcs := c.KubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace)
	var newPVC *corev1.PersistentVolumeClaim
	//check pvc exist
	existPVC, err := pvcs.Get(pvc.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			n, err := pvcs.Create(pvc)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			newPVC = n
		} else {
			log.Error(err.Error())
			return err
		}
	} else if existPVC != nil {
		// 如果已经存在 pending 的pvc ,先删除，后创建
		if existPVC.Status.Phase == apiv1.ClaimPending {
			err := pvcs.Delete(pvc.GetObjectMeta().GetName(), &metav1.DeleteOptions{})
			if err != nil {
				log.Error(err.Error())
				return err
			}
			time.Sleep(time.Second * 5)
			log.Info("time sleep 5 second ................")
			newPVC, err = pvcs.Create(pvc)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		} else if existPVC.Status.Phase == apiv1.ClaimBound {
			log.Infof("pvc %s status is bound", pvc.GetObjectMeta().GetName())
			if existPVC.Spec.VolumeName != pvc.Spec.VolumeName {
				log.Errorf("pvc %s got error bound pv : %s", pvc.GetObjectMeta().GetName(), existPVC.Spec.VolumeName)
				log.Errorf("pvc(%s) need to bound pv : %s, but this pvc(%s) is bound with pv : %s",
					pvc.GetObjectMeta().GetName(), pvc.Spec.VolumeName, pvc.GetObjectMeta().GetName(), existPVC.Spec.VolumeName)
				return fmt.Errorf("pvc(%s) need to bound pv : %s, but this pvc(%s) is bound with pv : %s",
					pvc.GetObjectMeta().GetName(), pvc.Spec.VolumeName, pvc.GetObjectMeta().GetName(), existPVC.Spec.VolumeName)
			}
			log.Warnf("pvc %s already bound to pv : %s", pvc.GetObjectMeta().GetName(), existPVC.Spec.VolumeName)
			return nil
		}
		_, err = pvcs.Update(pvc)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	_, err = c.CheckPVC(pvc)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Infof("Created PersistentVolumeClaim %q", newPVC.GetObjectMeta().GetName())
	return nil
}

// checkPVC check pvc exist
func (c *Clients) CheckPVC(pvc *corev1.PersistentVolumeClaim) (r bool, err error) {
	claims := c.KubeClient.CoreV1().PersistentVolumeClaims(pvc.GetObjectMeta().GetNamespace())
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			pvc, err := claims.Get(pvc.GetObjectMeta().GetName(), metav1.GetOptions{})
			if err != nil {
				errChan <- err
				return
			}
			if pvc.Status.Phase == apiv1.ClaimBound {
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
		return false, errors.New("pvc create after 100 second timeout")
	}
}

// DeletePersistentVolumeClaim .
func (c *Clients) DeletePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, ops *metav1.DeleteOptions) error {
	volumeClaims := c.KubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace)
	err := volumeClaims.Delete(pvc.Name, ops)
	if err != nil {
		log.Errorf(err.Error())
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	log.Infof("Delete PersistentVolumeClaim %q", pvc.GetObjectMeta().GetName())

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			_, err := volumeClaims.Get(pvc.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof("Check pvc status...")
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
		return errors.New("pvc check delete after 240 second timeout")
	}
}

// GetPersistentVolumeClaimByName .
func (c *Clients) GetPersistentVolumeClaimByName(name, namespace string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
	pvcClient := c.KubeClient.CoreV1().PersistentVolumeClaims(namespace)
	pvc, err := pvcClient.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return pvc, nil
}
