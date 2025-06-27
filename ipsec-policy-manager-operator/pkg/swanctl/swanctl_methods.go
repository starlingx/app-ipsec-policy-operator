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
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	for _, podSubnet := range c.PodSubnet {
		conn := vici.Connection{
			Children: map[string]*vici.ChildSA{
				"node-local-bypass": &(vici.ChildSA{
					Mode:                   BypassMode,
					StartAction:            BypassStartAction,
					LocalTrafficSelectors:  []string{podSubnet},
					RemoteTrafficSelectors: []string{podSubnet},
				}),
			},
		}

		if utility.GetIPVersion(podSubnet) == "IPv4" {
			conn.Name = "k8s-node-local"
		} else {
			conn.Name = "k8s-node-local-ipv6"
		}

		c.LocalConn = append(c.LocalConn, conn)
	}
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

		for _, clusterAddress := range c.ClusterHostAddr {
			ipVersion := utility.GetIPVersion(clusterAddress)
			nodeConnection := vici.SystemNodeConnection{
				ReauthTime:  ReauthTime,
				RekeyTime:   RekeyTime,
				Unique:      Unique,
				LocalAddrs:  []string{clusterAddress},
				RemoteAddrs: []string{utility.GetClusterHostIP(node.ClusterHostAddr, ipVersion)},
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

			if ipVersion == "IPv6" {
				nodeConnection.Name = fmt.Sprintf("k8s-node-%s-ipv6", node.Hostname)
			} else {
				nodeConnection.Name = fmt.Sprintf("k8s-node-%s", node.Hostname)
			}

			nodeConnection.Children = make(map[string]*vici.ChildSA)
			for _, policies := range policiesList.Items {
				// List of policies
				for _, policy := range policies.Spec.Policies {
					// Check service IP family
					ipFamilies, err := utility.GetServiceIPFamilies(policy.ServiceName, policy.ServiceNS)
					if err != nil {
						log.Info("Warning: IP Families not found", "Node", node.Hostname, "Service",
						            policy.ServiceName, "Namespace", policy.ServiceNS, "IP Version", ipVersion)
					} else if len(ipFamilies) == 1 && ipFamilies[0] != ipVersion {
						continue
					}

					// Capture Service IP of the nodes
					localServiceEndpointAddresses, err := utility.GetServiceAddresses(nodeName, policy.ServiceName, policy.ServiceNS, ipVersion)
					if err != nil {
						if client.IgnoreNotFound(err) == nil {
							log.Info("Warning: Service not found", "Node", node.Hostname,
									"Service", policy.ServiceName, "Namespace", policy.ServiceNS)
							continue
						}
						log.Error(err, "Unable to retrieve endpoints on current node for the service", "Node", node.Hostname,
								"Service", policy.ServiceName, "namespace", policy.ServiceNS)
						continue
					}
					log.Info(fmt.Sprintf("Endpoints on current node: %s for service: %s in namespace: %s: %v\n",
							nodeName, policy.ServiceName, policy.ServiceNS, localServiceEndpointAddresses))
					c.ServiceEndpointAddresses = localServiceEndpointAddresses

					nodeServiceEndpointAddresses, err := utility.GetServiceAddresses(node.Hostname, policy.ServiceName, policy.ServiceNS, ipVersion)
					if err != nil {
						if client.IgnoreNotFound(err) == nil {
							log.Info("Warning: Service not found", "Node", node.Hostname,
									"Service", policy.ServiceName, "Namespace", policy.ServiceNS)
							continue
						}
						log.Error(err, "Unable to retrieve endpoints on this node for the service", "Node", node.Hostname,
								"Service", policy.ServiceName, "namespace", policy.ServiceNS)
						continue
					}
					log.Info(fmt.Sprintf("Endpoints on node: %s for service: %s in namespace: %s: %v\n",
							node.Hostname, policy.ServiceName, policy.ServiceNS, nodeServiceEndpointAddresses))

					// ServicePorts: udp/XXXX,tcp/XXXX
					policyPortProtocols := utility.GetPolicyPorts(policy.ServicePorts)

					servicePortProtocols, err := utility.GetServicePorts(policy.ServiceName, policy.ServiceNS)
					if err != nil {
						if client.IgnoreNotFound(err) == nil {
							log.Info("Warning: Service not found", "Node", node.Hostname,
									"Service", policy.ServiceName, "Namespace", policy.ServiceNS)
							continue
						}
						log.Error(err, "Unable to retrieve service's port and protocol", "Node", node.Hostname,
								"Service", policy.ServiceName, "namespace", policy.ServiceNS)
						continue
					}

					portProtocols := utility.ProtectedPortsAndProtocols(policy.ServiceName, policyPortProtocols, servicePortProtocols)

					// ChildrenName: udp_serviceName_[egress|ingress]
					for _, portProtocol := range portProtocols {
						childName := fmt.Sprintf("%v_%v", portProtocol.Protocol, policy.ServiceName)

						if len(nodeServiceEndpointAddresses) > 0 {
							policyEgress := childName + "_egress"
							podSubnet := utility.GetPodSubnet(c.PodSubnet, ipVersion)
							trafficSelector := fmt.Sprintf("%s[%s]", podSubnet, portProtocol.Protocol)

							localTS := []string{trafficSelector}
							remoteTS := []string{}

							// loop through all the endpoints on this node
							for _, nodeServiceEndpointAddr := range nodeServiceEndpointAddresses {
								for _, port := range portProtocol.Ports {
									portsSpec := portProtocol.Protocol + "/" + fmt.Sprint(port)
									remoteTS = append(remoteTS, nodeServiceEndpointAddr+"["+portsSpec+"]")
								}
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

						if len(c.ServiceEndpointAddresses) > 0 {
							policyIngress := childName + "_ingress"
							podSubnet := utility.GetPodSubnet(node.PodSubnet, ipVersion)
							trafficSelector := fmt.Sprintf("%s[%s]", podSubnet, portProtocol.Protocol)

							localTS := []string{}
							remoteTS := []string{trafficSelector}

							// loop through all the endpoints on the current node
							for _, serviceEndpointAddr := range c.ServiceEndpointAddresses {
								for _, port := range portProtocol.Ports {
									portsSpec := portProtocol.Protocol + "/" + fmt.Sprint(port)
									localTS = append(localTS, serviceEndpointAddr+"["+portsSpec+"]")
								}
							}

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
