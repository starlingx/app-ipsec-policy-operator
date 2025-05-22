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

package utility

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/kubernetes"
)

// Convert a struct into YAML format
func GetYamlConf(data interface{}) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal struct to YAML: %w", err)
	}
	return string(yamlData), nil
}

// ContainsPort reports whether a port is present or not in an array of ports
func (m *PortProtocol) ContainsPort(newPort int64) bool {
	for _, port := range m.Ports {
		if port == newPort {
			return true
		}
	}
	return false
}

// ContainsProtocol reports whether a protocol is present or not in an array
// of PortProtocol structs
func ContainsProtocol(portProtocols []PortProtocol, protocol string) bool {
	for _, portProtocol := range portProtocols {
		if portProtocol.Protocol == protocol {
			return true
		}
	}
	return false
}

// GetPolicyPorts gets the ports specified by the users in the policy.
// It returns an array of PortProtocol structs
func GetPolicyPorts(servicePorts string) []PortProtocol {
	var portProtocols []PortProtocol

	for _, servicePort := range strings.Split(servicePorts, ",") {
		servicePort = strings.Trim(servicePort, " ")
		portInfo := strings.Split(servicePort, "/")
		protocol := portInfo[0]
		if ContainsProtocol(portProtocols, protocol) == false {
			var portProt PortProtocol
			portProt.Protocol = protocol
			portProtocols = append(portProtocols, portProt)
		}

		for i := range portProtocols {
			if portProtocols[i].Protocol == portInfo[0] {
				port, _ := strconv.ParseInt(portInfo[1], 10, 64)
				if portProtocols[i].ContainsPort(port) == false {
					portProtocols[i].Ports = append(portProtocols[i].Ports, port)
				}
			}
		}
	}

	return portProtocols
}

// GetServicePorts gets the service ports for a specific node using the
// service's Name and Namespace. It returns an array of PortProtocol structs
func GetServicePorts(nodeName string, serviceName string, serviceNamespace string) ([]PortProtocol, error) {
	var (
		rEndpoints = kubernetes.K8sResource{
			ApiGroup:   "",
			ApiVersion: "v1",
			Resource:   "endpoints",
			NameSpace:  serviceNamespace,
		}
		portProtocols []PortProtocol
	)

	endpoint, err := rEndpoints.RetrieveResourceInfoByName(serviceName)
	if err != nil {
		errMsg := fmt.Errorf("Service: %s - Namespace: %s", serviceName, serviceNamespace)
		return portProtocols, errMsg
	}

	subsets, found, err := unstructured.NestedSlice(endpoint.Object, "subsets")
	if err != nil || !found {
		errMsg := fmt.Errorf("Service: %s - Namespace: %s - error retrieving subsets: %w",
			serviceName, serviceNamespace, err)
		return portProtocols, errMsg
	}

	for _, subset := range subsets {
		if subsetMap, ret := subset.(map[string]interface{}); ret {
			// Extract Protocols/Ports
			ports, _, _ := unstructured.NestedSlice(subsetMap, "ports")
			for _, port := range ports {
				if portMap, ok := port.(map[string]interface{}); ok {
					if portNum, ok := portMap["port"].(int64); ok {
						if protocol, ok := portMap["protocol"].(string); ok {
							prot := strings.ToLower(protocol)
							if ContainsProtocol(portProtocols, prot) == false {
								var portProt PortProtocol
								portProt.Protocol = prot
								portProtocols = append(portProtocols, portProt)
							}

							for i := range portProtocols {
								if portProtocols[i].Protocol == prot {
									if portProtocols[i].ContainsPort(portNum) == false {
										portProtocols[i].Ports = append(portProtocols[i].Ports, portNum)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return portProtocols, nil
}

// ProtectedPortsAndProtocols validates and protect misconfigurations in the
// protocols/ports specified by the user. It returns a list of protocol/ports
// that user specified in policies to protect for the service.
func ProtectedPortsAndProtocols(serviceName string, policyPortProtocols []PortProtocol, servicePortProtocols []PortProtocol) []PortProtocol {
	ctx := context.Background()
	log := log.FromContext(ctx)
	var portProtocols []PortProtocol
	for _, policyPortProtocol := range policyPortProtocols {
		if ContainsProtocol(servicePortProtocols, policyPortProtocol.Protocol) == true {
			for _, servicePortProtocol := range servicePortProtocols {
				if policyPortProtocol.Protocol == servicePortProtocol.Protocol {
					for _, policyPort := range policyPortProtocol.Ports {
						if servicePortProtocol.ContainsPort(policyPort) {
							if ContainsProtocol(portProtocols, policyPortProtocol.Protocol) == false {
								var portProt PortProtocol
								portProt.Protocol = policyPortProtocol.Protocol
								portProtocols = append(portProtocols, portProt)
							}

							for i := range portProtocols {
								if portProtocols[i].Protocol == policyPortProtocol.Protocol {
									if portProtocols[i].ContainsPort(policyPort) == false {
										portProtocols[i].Ports = append(portProtocols[i].Ports, policyPort)
									}
								}
							}
						} else {
							log.Info(fmt.Sprintf("Service: %v - Protocol/Port: %v/%v not configured in the service\n",
								serviceName, policyPortProtocol.Protocol, policyPort))
						}
					}
				}
			}
		} else {
			log.Info(fmt.Sprintf("Service: %v - Protocol: %v not configured in the service\n", serviceName, policyPortProtocol.Protocol))
		}
	}

	return portProtocols
}

// GetServiceAddress gets the Service Endpoint IP address for a specific
// Node using the service's Name and Namespace. It returns a string with the IP address.
func GetServiceAddress(nodeName string, serviceName string, serviceNamespace string) (string, error) {
	var (
		rEndpoints = kubernetes.K8sResource {
			ApiGroup:   "",
			ApiVersion: "v1",
			Resource:   "endpoints",
			NameSpace:  serviceNamespace,
		}
		serviceAddr string
	)

	endpoint, err := rEndpoints.RetrieveResourceInfoByName(serviceName)
	if err != nil {
		errMsg := fmt.Errorf("Service: %s - Namespace: %s", serviceName, serviceNamespace)
		return "", errMsg
	}

	subsets, found, err := unstructured.NestedSlice(endpoint.Object, "subsets")
	if err != nil || !found {
		errMsg := fmt.Errorf("Node: %s - Service: %s - Namespace: %s - error retrieving subsets: %w",
			nodeName, serviceName, serviceNamespace, err)
		return "", errMsg
	}

	for _, subset := range subsets {
		if subsetMap, ret := subset.(map[string]interface{}); ret {
			// Extract IPs
			addresses, _, _ := unstructured.NestedSlice(subsetMap, "addresses")
			for _, addr := range addresses {
				if addrMap, ret := addr.(map[string]interface{}); ret {
					if ndName, ret := addrMap["nodeName"].(string); ret {
						if ip, ret := addrMap["ip"].(string); ret && nodeName == ndName {
							serviceAddr = ip
						}
					}
				}
			}
		}
	}

	return serviceAddr, nil
}
