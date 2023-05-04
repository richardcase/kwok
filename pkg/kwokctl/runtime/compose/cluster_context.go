/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package compose

import (
	"context"
	"fmt"
	"os"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"sigs.k8s.io/kwok/pkg/kwokctl/runtime"
	"sigs.k8s.io/kwok/pkg/utils/format"
	"sigs.k8s.io/kwok/pkg/utils/kubeconfig"
	"sigs.k8s.io/kwok/pkg/utils/path"
)

// AddContext add the context of cluster to kubeconfig
func (c *Cluster) AddContext(ctx context.Context, kubeconfigPath string) error {
	config, err := c.Config(ctx)
	if err != nil {
		return err
	}
	conf := &config.Options

	scheme := "http"
	if conf.SecurePort {
		scheme = "https"
	}
	localAddress := conf.BindAddress

	pkiPath := c.GetWorkdirPath(runtime.PkiName)
	adminKeyPath := path.Join(pkiPath, "admin.key")
	adminCertPath := path.Join(pkiPath, "admin.crt")

	// set the context in default kubeconfig
	kubeConfig := &kubeconfig.Config{
		Cluster: &clientcmdapi.Cluster{
			Server: scheme + "://" + localAddress + ":" + format.String(conf.KubeApiserverPort),
		},
		Context: &clientcmdapi.Context{
			Cluster: c.Name(),
		},
	}
	if conf.SecurePort {
		kubeConfig.Cluster.InsecureSkipTLSVerify = true
		kubeConfig.Context.AuthInfo = c.Name()
		if !conf.KubeConfigEmbedCerts {
			kubeConfig.User = &clientcmdapi.AuthInfo{
				ClientCertificate: adminCertPath,
				ClientKey:         adminKeyPath,
			}
		} else {
			certData, err := os.ReadFile(adminCertPath)
			if err != nil {
				return fmt.Errorf("reading certificate file %s: %w", adminCertPath, err)
			}

			keyData, err := os.ReadFile(adminKeyPath)
			if err != nil {
				return fmt.Errorf("reading certificate file %s: %w", adminKeyPath, err)
			}

			kubeConfig.User = &clientcmdapi.AuthInfo{
				ClientCertificateData: certData,
				ClientKeyData:         keyData,
			}
		}

	}
	err = kubeconfig.AddContext(kubeconfigPath, c.Name(), kubeConfig)
	if err != nil {
		return err
	}
	return nil
}

// RemoveContext remove the context of cluster from kubeconfig
func (c *Cluster) RemoveContext(ctx context.Context, kubeconfigPath string) error {
	err := kubeconfig.RemoveContext(kubeconfigPath, c.Name())
	if err != nil {
		return err
	}
	return nil
}
