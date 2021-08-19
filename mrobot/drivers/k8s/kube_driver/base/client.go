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

package base

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	k   *kubernetes.Clientset
	ctx context.Context
}

// NewClientInCluster new client in cluster
func NewClientInCluster(ctx context.Context) (*Client, error)  {
	c, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "get in cluster kube config")
	}
	kubeClient, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, errors.Wrap(err, "new kube client")
	}
	return &Client{k: kubeClient, ctx: ctx}, nil
}

// NewClientByConfig new client by config
func NewClientByConfig(ctx context.Context,kubeConfig []byte) (*Client, error)  {
	if len(kubeConfig) == 0 {
		return nil, errors.New("Error Get empty kube config ")
	}
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new config by kube config")
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "new kube client")
	}
	return &Client{k: kubeClient, ctx: ctx}, nil
}

// NewClientByAuth new client by auth
func NewClientByAuth(ctx context.Context,apiURL, token, cert string) (*Client, error)  {
	if len(apiURL) == 0 || len(token) == 0 || len(cert) == 0 {
		return nil, errors.New("Error Get empty param config ")
	}
	config, err := newConfig(apiURL, token, cert)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "new client for config")
	}
	return &Client{k: kubeClient, ctx: ctx}, nil
}

func newConfig(apiURL string, token string, caCert string) (*rest.Config, error) {

	tlsClientConfig := rest.TLSClientConfig{}

	if _, err := newCertPool(caCert); err != nil {
		return nil, err
	} else {
		tlsClientConfig.CAData = []byte(caCert)
	}
	return &rest.Config{
		Host:            apiURL,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     token,
	}, nil
}

func newCertPool(caCert string) (*x509.CertPool, error) {
	certs, err := ParseCertsPEM([]byte(caCert))
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool, nil
}

// ParseCertsPEM returns the x509.Certificates contained in the given PEM-encoded byte array
// Returns an error if a certificate could not be parsed, or if the data does not contain any certificates
func ParseCertsPEM(pemCerts []byte) ([]*x509.Certificate, error) {
	ok := false
	certs := []*x509.Certificate{}
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		// Only use PEM "CERTIFICATE" blocks without extra headers
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, err
		}

		certs = append(certs, cert)
		ok = true
	}

	if !ok {
		return certs, errors.New("data does not contain any valid RSA or ECDSA certificates")
	}
	return certs, nil
}
