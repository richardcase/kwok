/*
Copyright 2022 The Kubernetes Authors.

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

package k8s

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// BuildKubeconfig builds a kubeconfig file from the given parameters.
func BuildKubeconfig(conf BuildKubeconfigConfig) (string, error) {

	cfg := &v1.Config{
		APIVersion:     v1.SchemeGroupVersion.Version,
		Kind:           "Config",
		CurrentContext: conf.ProjectName,
		Clusters: []v1.NamedCluster{
			{
				Name: conf.ProjectName,
				Cluster: v1.Cluster{
					Server: conf.Address,
				},
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: conf.ProjectName,
				Context: v1.Context{
					Cluster: conf.ProjectName,
				},
			},
		},
	}

	if conf.SecurePort {
		cfg.Clusters[0].Cluster.InsecureSkipTLSVerify = true
		cfg.Contexts[0].Context.AuthInfo = conf.ProjectName

		user := v1.AuthInfo{}
		if !conf.EmbedCerts {
			user.ClientCertificate = conf.AdminCrtPath
			user.ClientKey = conf.AdminKeyPath
		} else {
			data, err := os.ReadFile(conf.AdminCrtPath)
			if err != nil {
				return "", fmt.Errorf("reading certificate file %s: %w", conf.AdminCrtPath, err)
			}
			user.ClientCertificateData = data

			data, err = os.ReadFile(conf.AdminKeyPath)
			if err != nil {
				return "", fmt.Errorf("reading certificate file %s: %w", conf.AdminKeyPath, err)
			}
			user.ClientKeyData = data
		}

		cfg.AuthInfos = []v1.NamedAuthInfo{
			{
				Name:     conf.ProjectName,
				AuthInfo: user,
			},
		}
	}

	out, err := runtime.Encode(clientcmdlatest.Codec, cfg)

	//out, err := clientcmd.Write(*cfg)
	if err != nil {
		return "", errors.Wrap(err, "failed to serialize config to yaml")
	}

	return string(out), nil
}

// BuildKubeconfigConfig is the configuration for BuildKubeconfig.
type BuildKubeconfigConfig struct {
	ProjectName  string
	SecurePort   bool
	Address      string
	AdminCrtPath string
	AdminKeyPath string
	EmbedCerts   bool
}
