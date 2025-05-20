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

package swanctl

import (
	"context"
	"encoding/json"
	"fmt"

	api "starlingx.windriver.com/ipsec-policy-manager-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/kubernetes"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/utility"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/vici"
)

func (c *ConfigurationFile) MarshalLocalConn() string {
	ctx := context.Background()
	log := log.FromContext(ctx)
	var err error
	var jsonData string

	// Marshal the slice of VICI messages into JSON.
	jsonBytes, err := json.MarshalIndent(c.LocalConn, "", "  ")
	if err != nil {
		log.Error(err, "failed to marshal VICI messages")
		return jsonData
	}

	jsonData = string(jsonBytes)

	return jsonData
}

func (c *ConfigurationFile) MarshalConnections() string {
	ctx := context.Background()
	log := log.FromContext(ctx)
	var err error
	var jsonData string

	// Marshal the slice of VICI messages into JSON.
	jsonBytes, err := json.MarshalIndent(c.Connections, "", "  ")
	if err != nil {
		log.Error(err, "failed to marshal VICI messages")
		return jsonData
	}

	fmt.Println(jsonData)
	jsonData = string(jsonBytes)

	return jsonData
}

func (c *ConfigurationFile) getLocalConf() {
	conn := vici.Connection{
		Name: "k8s-node-local",
		Children: map[string]*vici.ChildSA{
			"k8s-node-bypass": &(vici.ChildSA{
				Mode:                   BypassMode,
				StartAction:            BypassStartAction,
				LocalTrafficSelectors:  []string{c.PodSubnet},
				RemoteTrafficSelectors: []string{c.PodSubnet},
			}),
		},
	}

	c.LocalConn = append(c.LocalConn, conn)
}

func (c *ConfigurationFile) GetNodesConf(nodeName string, policiesList api.IPsecPolicyList) error {
	ctx := context.Background()
	log := log.FromContext(ctx)
	currentNode, err := kubernetes.GetCurrentNodeConfiguration(nodeName)
	if err != nil {
		return fmt.Errorf("unable to retrive current node configuration. %w", err)
	}

	c.Hostname = currentNode.Hostname
	c.PodSubnet = currentNode.PodSubnet
	c.ClusterHostAddr = currentNode.ClusterHostAddr

	nodesConf, err := kubernetes.GetNodesConfiguration()
	if err != nil {
		return fmt.Errorf("unable to retrive nodes configuration. %w", err)
	}
	for _, node := range nodesConf.Nodes {
		if node.Hostname == c.Hostname {
			c.getLocalConf()
			continue
		}

		nodeConnection := vici.SystemNodeConnection{
			Name:        "k8s-node-" + node.Hostname,
			ReauthTime:  ReauthTime,
			RekeyTime:   RekeyTime,
			Unique:      Unique,
			LocalAddrs:  []string{c.ClusterHostAddr},
			RemoteAddrs: []string{node.ClusterHostAddr},
			Local: &vici.LocalOpts{
				Auth: LocalOptsAuth,
				Cert: &vici.CertBlock{
					File: IPsecCertPath + CertificatePrefix + c.Hostname + CertificateExtension,
				},
			},
			Remote: &vici.RemoteOpts{
				ID:   RemoteOptsID,
				Auth: RemoteOptsAuth,
				CACert0: &vici.CertBlock{
					File: IPsecCertCAPath + SystemLocalCACert0,
				},
				CACert1: &vici.CertBlock{
					File: IPsecCertCAPath + SystemLocalCACert1,
				},
			},
		}

		nodeConnection.Children = make(map[string]*vici.ChildSA)
		for _, policies := range policiesList.Items {
			// List of policies
			for _, policy := range policies.Spec.Policies {
				// Capture Service IP of the nodes
				localServiceEndpointAddr, err := utility.GetServiceAddress(nodeName, policy.ServiceName, policy.ServiceNS)
				if err != nil {
					log.Error(err, "Unable to retrieve current node endpoints configuration")
					continue
				}
				c.ServiceEndpointAddr = localServiceEndpointAddr

				nodeServiceEndpointAddr, err := utility.GetServiceAddress(node.Hostname, policy.ServiceName, policy.ServiceNS)
				if err != nil {
					log.Error(err, "Unable to retrieve node endpoints configuration", "Node", node.Hostname)
					continue
				}

				// ServicePorts: udp/XXXX,tcp/XXXX
				policyPortProtocols := utility.GetPolicyPorts(policy.ServicePorts)

				servicePortProtocols, err := utility.GetServicePorts(c.Hostname, policy.ServiceName, policy.ServiceNS)
				if err != nil {
					log.Error(err, "Unable to retrieve node endpoints configuration", "Node", node.Hostname)
					continue
				}

				portProtocols := utility.ProtectedPortsAndProtocols(policy.ServiceName, policyPortProtocols, servicePortProtocols)

				// ChildrenName: udp_serviceName_[egress|ingress]
				for _, portProtocol := range portProtocols {
					childName := fmt.Sprintf("%v_%v", portProtocol.Protocol, policy.ServiceName)

					if nodeServiceEndpointAddr != "" {
						policyEgress := childName + "_egress"
						localTS := []string{c.PodSubnet}
						remoteTS := []string{}
						for _, port := range portProtocol.Ports {
							portsSpec := portProtocol.Protocol + "/" + fmt.Sprint(port)
							remoteTS = append(remoteTS, nodeServiceEndpointAddr+"["+portsSpec+"]")
						}

						childEgress := &vici.ChildSA{
							Mode:                   EgressMode,
							StartAction:            EgressStartAction,
							LocalTrafficSelectors:  localTS,
							RemoteTrafficSelectors: remoteTS,
							Updown:                 EgressUpdown,
						}
						nodeConnection.Children[policyEgress] = childEgress
					}

					if c.ServiceEndpointAddr != "" {
						policyIngress := childName + "_ingress"
						localTS := []string{}
						for _, port := range portProtocol.Ports {
							portsSpec := portProtocol.Protocol + "/" + fmt.Sprint(port)
							localTS = append(localTS, c.ServiceEndpointAddr+"["+portsSpec+"]")
						}
						remoteTS := []string{node.PodSubnet}
						childIngress := &vici.ChildSA{
							Mode:                   IngressMode,
							StartAction:            IngressStartAction,
							LocalTrafficSelectors:  localTS,
							RemoteTrafficSelectors: remoteTS,
							Updown:                 IngressUpdown,
						}
						nodeConnection.Children[policyIngress] = childIngress
					}
				}
			}
		}

		if len(nodeConnection.Children) > 0 {
			c.Connections = append(c.Connections, nodeConnection)
		}
	}

	return nil
}

func (c *ConfigurationFile) GetConfigData() (map[string]string, error) {
	configData := map[string]string{}

	localconn := c.MarshalLocalConn()
	configData["local_conn"] = localconn
	connections := c.MarshalConnections()
	configData["connections"] = connections

	return configData, nil
}
