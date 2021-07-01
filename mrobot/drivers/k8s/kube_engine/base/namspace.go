package kubeclient

import (
	log "git.hyperchain.cn/blocface/golog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// GetNamespaceList .
func (c *Client) GetNamespaceList(ops metav1.ListOptions) *corev1.NamespaceList {

	nss, err := c.KubeClient.CoreV1().Namespaces().List(ops)
	if err != nil {
		log.Errorf(err.Error())
	}
	for _, ns := range nss.Items {
		log.Infof("Namespace：", ns.Name, ns.Status.Phase)
	}
	return nss
}

// CreateNameSpace .
func (c *Client) CreateNameSpace(ns *corev1.Namespace) error {
	nameSpace, err := c.KubeClient.CoreV1().Namespaces().Create(ns)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	log.Infof("Created namesapce %q \n", nameSpace.GetObjectMeta().GetName())
	return nil
}

// DeleteNameSpace .
func (c *Client) DeleteNameSpace(ns *corev1.Namespace, ops *metav1.DeleteOptions) error {
	err := c.KubeClient.CoreV1().Namespaces().Delete(ns.Name, ops)
	if err != nil {
		log.Errorf(err.Error())
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	log.Infof("Delete namesapce %q \n", ns.GetObjectMeta().GetName())
	return err
}
