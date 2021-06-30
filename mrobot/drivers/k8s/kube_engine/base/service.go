package kubeclient

import (
	"errors"
	log "git.hyperchain.cn/blocface/golog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

//GetServiceList .
func (c *Clients) GetServiceList(ns string, ops metav1.ListOptions) *corev1.ServiceList {

	services, err := c.KubeClient.CoreV1().Services(ns).List(ops)
	if err != nil {
		log.Error(err.Error())
	}
	for _, service := range services.Items {
		log.Infof("Service：", service.Name, service.GetUID())
	}
	return services
}

//GetServiceList .
func (c *Clients) GetServiceByName(name,namespace string, ops metav1.GetOptions) (*corev1.Service, error) {
	service, err := c.KubeClient.CoreV1().Services(namespace).Get(name, ops)
	if err != nil {
		return nil, err
	}
	return service, nil
}

//CreateService .
func (c *Clients) CreateService(service *corev1.Service) error {
	newservice, err := c.KubeClient.CoreV1().Services(service.Namespace).Create(service)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		log.Error(err.Error())
		return err
	}
	log.Infof("Created Service %q", newservice.GetObjectMeta().GetName())
	return nil
}

// DeleteService .
func (c *Clients) DeleteService(service *corev1.Service, ops *metav1.DeleteOptions) error {
	err := c.KubeClient.CoreV1().Services(service.Namespace).Delete(service.Name, ops)
	if err != nil {
		log.Error(err.Error())
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	log.Infof("Delete Service %q", service.GetObjectMeta().GetName())
	return err
}

//CreateService .
func (c *Clients) UpdateService(service *corev1.Service) error {
	serviceClient := c.KubeClient.CoreV1().Services(service.Namespace) //.Update(service)
	if serviceClient == nil {
		log.Errorf("cant get serviceClient")
		return errors.New("cant get serviceClient")
	}

	oldservice, err := serviceClient.Get(service.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		log.Errorf("get present service fail: %s", err.Error())
		return err
	}
	if oldservice == nil || oldservice.ResourceVersion == "" {
		log.Errorf("cant get present ResourceVersion")
		return errors.New("cant get present ResourceVersion")
	}

	log.Infof("old.ResourceVersion is: %s ===  oldservice.Spec.ClusterIP is %s", oldservice.ResourceVersion, oldservice.Spec.ClusterIP)
	service.ResourceVersion = oldservice.ResourceVersion
	service.Spec.ClusterIP = oldservice.Spec.ClusterIP
	newS, err := serviceClient.Update(service)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	log.Infof("Updated deployment %q \n", newS.GetObjectMeta().GetName())

	return nil
}
