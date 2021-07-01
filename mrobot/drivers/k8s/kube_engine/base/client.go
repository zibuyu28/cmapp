/*
 * Copyright © 2021 zibuyu28
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kubeclient

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
	"path/filepath"
)

type Client struct {
	k   *kubernetes.Clientset
	ctx context.Context
}

//NewClient new client by config
func NewClient(ctx context.Context, kubeconfig string) (*Client, error) {
	// uses the current context in kubeconfig
	config, err := loadConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{k: kubeClient, ctx: ctx}, nil
}

// loadConfig loads a REST Config as per the rules specified in GetConfig
func loadConfig(kubeconfig string) (*rest.Config, error) {

	var apiServerURL = ""
	// If a flag is specified with the kubeconfig location, use that
	if len(kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags(apiServerURL, kubeconfig)
	}
	// If an env variable is specified with the kubeconfig locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags(apiServerURL, os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster kubeconfig
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster kubeconfig, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "kubeconfig")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}
