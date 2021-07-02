package base

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreatePersistentVolumeClaim .
func (c *Client) CreatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) error {
	pvcs := c.k.CoreV1().PersistentVolumeClaims(pvc.Namespace)
	//check pvc exist
	existPVC, err := pvcs.Get(c.ctx, pvc.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_, err := pvcs.Create(c.ctx, pvc, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrapf(err, "create pvc [%s]", pvc.Name)
			}
		} else {
			return errors.Wrapf(err, "get pvc [%s]", pvc.Name)
		}
	} else if existPVC != nil {
		// 如果已经存在 pending 的pvc ,先删除，后创建
		if existPVC.Status.Phase == apiv1.ClaimPending {
			err := pvcs.Delete(c.ctx, pvc.GetObjectMeta().GetName(), metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrapf(err, "delete pvc [%s]", pvc.Name)
			}
			time.Sleep(time.Second * 5)
			_, err = pvcs.Create(c.ctx, pvc, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrapf(err, "recreate pvc [%s]", pvc.Name)
			}
		} else if existPVC.Status.Phase == apiv1.ClaimBound {
			log.Infof(c.ctx, "Currently pvc [%s] status is bound", pvc.GetObjectMeta().GetName())
			if existPVC.Spec.VolumeName != pvc.Spec.VolumeName {
				log.Errorf(c.ctx, "Currently pvc [%s] got error when bound pv [%s]", pvc.GetObjectMeta().GetName(), existPVC.Spec.VolumeName)
				log.Errorf(c.ctx, "Currently pvc [%s] need to bound pv [%s], but this pvc [%s] is bound with pv %[s]",
					pvc.GetObjectMeta().GetName(), pvc.Spec.VolumeName, pvc.GetObjectMeta().GetName(), existPVC.Spec.VolumeName)
				return fmt.Errorf("fail to bond pv, because pvc [%s] need to bound pv [%s], but this pvc [%s] is bound with pv [%s]",
					pvc.GetObjectMeta().GetName(), pvc.Spec.VolumeName, pvc.GetObjectMeta().GetName(), existPVC.Spec.VolumeName)
			}
			return nil
		}
		_, err = pvcs.Update(c.ctx, pvc, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "update pvc [%s]", pvc.Name)
		}
	}
	_, err = c.CheckPVC(pvc)
	if err != nil {
		return errors.Wrapf(err, "check pvc [%s] status", pvc.Name)
	}
	return nil
}

// checkPVC check pvc exist
func (c *Client) CheckPVC(pvc *corev1.PersistentVolumeClaim) (r bool, err error) {
	claims := c.k.CoreV1().PersistentVolumeClaims(pvc.GetObjectMeta().GetNamespace())
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			pvc, err := claims.Get(ctx, pvc.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof(ctx, "pvc creating...")
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
		return false, errors.New("pvc create after 100 second timeout")
	case <-c.ctx.Done():
		return false, errors.New("pvc create state unknown with context deadline")
	}
}

// DeletePersistentVolumeClaim .
func (c *Client) DeletePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, ops metav1.DeleteOptions) error {
	volumeClaims := c.k.CoreV1().PersistentVolumeClaims(pvc.Namespace)
	err := volumeClaims.Delete(c.ctx, pvc.Name, ops)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return errors.Wrapf(err, "delete pvc [%s]", pvc.Name)
	}

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			_, err := volumeClaims.Get(c.ctx, pvc.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof(ctx, "pvc deleting...")
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
		return errors.New("pvc check delete after 300 second timeout")
	case <-c.ctx.Done():
		return errors.New("pvc delete state unknown with context deadline")
	}
}
