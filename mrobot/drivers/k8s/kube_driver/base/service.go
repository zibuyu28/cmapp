package base

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

//CreateService .
func (c *Client) CreateService(service *corev1.Service) error {
	_, err := c.k.CoreV1().Services(service.Namespace).Create(c.ctx, service, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return errors.Wrapf(err, "create service [%s]", service.Name)
	}
	return nil
}

// DeleteService .
func (c *Client) DeleteService(service *corev1.Service, ops metav1.DeleteOptions) error {
	err := c.k.CoreV1().Services(service.Namespace).Delete(c.ctx, service.Name, ops)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return errors.Wrapf(err, "delete")
	}
	return err
}

//CreateService .
func (c *Client) UpdateService(service *corev1.Service) error {
	serviceClient := c.k.CoreV1().Services(service.Namespace) //.Update(service)
	if serviceClient == nil {
		return errors.New("cant get service client")
	}

	oldservice, err := serviceClient.Get(c.ctx, service.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "get servcie [%v]", service.Name)
	}
	if oldservice == nil || oldservice.ResourceVersion == "" {
		return errors.New("cant get present ResourceVersion")
	}

	service.ResourceVersion = oldservice.ResourceVersion
	service.Spec.ClusterIP = oldservice.Spec.ClusterIP
	_, err = serviceClient.Update(c.ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "update servcie [%s]", service.Name)
	}

	return nil
}
