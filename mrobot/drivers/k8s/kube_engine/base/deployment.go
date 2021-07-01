package kubeclient

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreateDeployment .
func (c *Client) CreateDeployment(dep *appsv1.Deployment) error {
	if dep.Namespace == "" {
		dep.Namespace = corev1.NamespaceDefault
	}
	deploymentsClient := c.k.AppsV1().Deployments(dep.Namespace)

	var newDep *appsv1.Deployment
	old, err := deploymentsClient.Get(c.ctx, dep.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			n, err := deploymentsClient.Create(c.ctx, dep, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrap(err, "create deployment")
			}
			newDep = n
		} else {
			return errors.Wrap(err, "get deployment")
		}
	} else if old != nil {
		newDep, err = deploymentsClient.Update(c.ctx, dep, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrap(err, "update deployment")
		}
	}
	// need to check pod
	_, err = c.CheckDeployment(newDep)
	if err != nil {
		return errors.Wrap(err, "check deployment state finish")
	}
	return nil
}

// CheckDeployment check deployment exist
func (c *Client) CheckDeployment(dep *appsv1.Deployment) (r bool, err error) {
	deploymentsClient := c.k.AppsV1().Deployments(dep.GetObjectMeta().GetNamespace())
	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(c.ctx)
	defer cancelFunc()
	go func(ctx context.Context) {
		for {
			d, err := deploymentsClient.Get(ctx, dep.GetObjectMeta().GetName(), metav1.GetOptions{})
			if err != nil {
				errChan <- err
				return
			}
			if d.Status.AvailableReplicas == *d.Spec.Replicas {
				result <- true
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				log.Infof(ctx,"deployment creating...")
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
		return false, errors.New("deployment create failed after 300 second timeout")
	case <-c.ctx.Done():
		return false, errors.New("deployment create failed with context deadline")
	}
}

//GetDeployment .
func (c *Client) GetDeploymentByName(name, namespace string, ops metav1.GetOptions) (*appsv1.Deployment, error) {
	deploymentsClient := c.KubeClient.AppsV1().Deployments(namespace)
	redep, err := deploymentsClient.Get(name, ops)
	if err != nil {
		return nil, err
	}
	return redep, nil
}

//GetDeploymentList .
func (c *Client) GetDeploymentList(dep *appsv1.Deployment, ops metav1.ListOptions) *appsv1.DeploymentList {
	deploymentsClient := c.KubeClient.AppsV1().Deployments(dep.Namespace)
	list, err := deploymentsClient.List(ops)
	if err != nil {
		log.Errorf(err.Error())
	}
	for _, d := range list.Items {
		log.Infof("Deployment ：", d.Name, d.Spec.Replicas)
	}
	return list
}

//DeleteDeployment .
func (c *Client) DeleteDeployment(dep *appsv1.Deployment, ops *metav1.DeleteOptions) error {
	deploymentsClient := c.KubeClient.AppsV1().Deployments(dep.Namespace)
	err := deploymentsClient.Delete(dep.Name, ops)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	log.Infof("Delete deployment %q", dep.Name)

	var errChan = make(chan error, 1)
	var result = make(chan bool, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			_, err := deploymentsClient.Get(dep.GetObjectMeta().GetName(), metav1.GetOptions{})
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
				log.Infof("Check deployment deleted...")
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
		return errors.New("deployment deleted check failed after 240 second timeout")
	}
}

//UpdateDeployment .
func (c *Client) UpdateDeployment(dep *appsv1.Deployment) error {
	deploymentsClient := c.KubeClient.AppsV1().Deployments(dep.Namespace)
	old, err := deploymentsClient.Get(dep.GetObjectMeta().GetName(), metav1.GetOptions{})
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	if old == nil || old.ResourceVersion == "" {
		log.Errorf("cant get present ResourceVersion")
		return errors.New("cant get present ResourceVersion")
	}
	log.Infof("old.ResourceVersion is: %s", old.ResourceVersion)
	dep.ResourceVersion = old.ResourceVersion
	newDep, err := deploymentsClient.Update(dep)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	log.Infof("Updated deployment %q \n", newDep.GetObjectMeta().GetName())
	return nil
}

func (c *Client) ExecDeploymentPod(dep *appsv1.Deployment, podName string, command string) (err error, sout, serr string) {
	req := c.KubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(dep.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			TypeMeta: metav1.TypeMeta{},
			Stdin:    true,
			Stdout:   true,
			Stderr:   true,
			TTY:      false,
			Command:  []string{"bash", "-c", command},
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
	if err != nil {
		return err, sout, serr
	}

	// 使用bytes.Buffer变量接收标准输出和标准错误
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return err, sout, serr
	}

	sout = stdout.String()
	serr = stderr.String()
	return err, sout, serr
}
