/*
Copyright (c) 2025 Wind River Systems, Inc.

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

package config

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "starlingx.io/ipsec-policy-manager-operator/api/v1"
	"starlingx.io/ipsec-policy-manager-operator/pkg/kubernetes"
	"starlingx.io/ipsec-policy-manager-operator/pkg/swanctl"
)

func GenerateConf(k8sClient client.Client, crList api.IPsecPolicyList) error {
	ctx := context.Background()
	log := log.FromContext(ctx)

	nodesConf, err := kubernetes.GetNodesConfiguration()
	if err != nil {
		log.Error(err, "Unable to retrieve nodes configuration")
		return err
	}

	var configFile *swanctl.ConfigurationFile
	for _, node := range nodesConf.Nodes {
		configFile = new(swanctl.ConfigurationFile)
		if err = configFile.GetNodesConf(node.Hostname, crList); err != nil {
			log.Error(err, "Unable to generate IPsec configuration", "Node", node.Hostname)
			return err
		}

		configData, err := configFile.GetConfigData()
		if err != nil {
			log.Error(err, "Unable to retrieve IPsec configuration data", "Node", node.Hostname)
			return err
		}

		// Create or update the ConfigMap
		configMapName := kubernetes.IPsecConfigMapPrefix + node.Hostname
		err = kubernetes.CreateOrUpdateConfigMap(
			k8sClient, kubernetes.OperatorNamespace, configMapName, configData)
		if err != nil {
			log.Error(err, "Failed to create/update configMap", "Node", node.Hostname)
			return err
		}

		configFile = nil
	}

	return nil
}
