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
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	api "starlingx.windriver.com/ipsec-policy-manager-operator/api/v1"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/kubernetes"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/swanctl"
)

func GenerateConf(k8sClient client.Client, crList api.IPsecPolicyList) error {
	nodesConf, err := kubernetes.GetNodesConfiguration()
	if err != nil {
		fmt.Println("unable to retrieve nodes configuration.", err)
		return err
	}

	var configFile *swanctl.ConfigurationFile
	for _, node := range nodesConf.Nodes {
		configFile = new(swanctl.ConfigurationFile)
		if err = configFile.GetNodesConf(node.Hostname, crList); err != nil {
			fmt.Println(node.Hostname, ": Unable to generate IPsec configuration: ", err)
			return err
		}
		configData, err := configFile.GetConfigData()
		if err != nil {
			fmt.Println(node.Hostname, ": Unable to retrieve IPsec configuration data: ", err)
			return err
		}

		// Create or update the ConfigMap
		configMapName := kubernetes.IPsecConfigMapPrefix + node.Hostname
		err = kubernetes.CreateOrUpdateConfigMap(
			k8sClient, kubernetes.OperatorNamespace, configMapName, configData)
		if err != nil {
			fmt.Println(node.Hostname, ": Failed to create/update configMap: ", err)
			return err
		}

		configFile = nil
	}

	return nil
}
