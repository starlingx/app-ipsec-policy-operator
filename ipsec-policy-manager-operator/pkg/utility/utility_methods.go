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
	"net"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"starlingx.io/ipsec-policy-manager-operator/pkg/kubernetes"
)

// GetYamlConf converts a struct into YAML format
func GetYamlConf(data interface{}) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal struct to YAML: %w", err)
	}
	return string(yamlData), nil
}

// GetIPVersion returns the IP version of an address
func GetIPVersion(address string) string {
	internalIP := net.ParseIP(strings.Split(address, "/")[0])
	if internalIP.To4() != nil {
		return "IPv4"
	}
	return "IPv6"
}

// GetClusterHostIP gets the IP address from a string slice of Cluster Host
// IP addresses based on the IP version
func GetClusterHostIP(clusterHostAddrs []string, ipVersion string) string {
	for _, ipAddress := range clusterHostAddrs {
		if GetIPVersion(ipAddress) == ipVersion {
			return ipAddress
		}
	}
	return ""
}

// GetPodSubnet gets the IP address from a string slice of Pod Subnets
// based on the IP version
func GetPodSubnet(podSubnet []string, ipVersion string) string {
	for _, subnet := range podSubnet {
		address := strings.Split(subnet, "/")
		if GetIPVersion(address[0]) == ipVersion {
			return subnet
		}
	}
	return ""
}

// ContainsPort reports whether a port is present or not in an array of ports
func ContainsPort(Ports []int64, newPort int64) bool {
	for _, port := range Ports {
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
				if ContainsPort(portProtocols[i].Ports, port) == false {
					portProtocols[i].Ports = append(portProtocols[i].Ports, port)
				}
			}
		}
	}

	return portProtocols
}

// GetServicePorts gets the service ports using the service's Name and Namespace.
// It returns an array of PortProtocol structs
func GetServicePorts(serviceName string, serviceNamespace string) ([]PortProtocol, error) {
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
		return portProtocols, err
	}

	subsets, found, err := unstructured.NestedSlice(endpoint.Object, "subsets")
	if err != nil {
		errMsg := fmt.Errorf("error retrieving subsets: %w", err)
		return portProtocols, errMsg
	}

	if !found {
		errMsg := fmt.Errorf("subsets not found")
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
									if ContainsPort(portProtocols[i].Ports, portNum) == false {
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
						if ContainsPort(servicePortProtocol.Ports, policyPort) {
							if ContainsProtocol(portProtocols, policyPortProtocol.Protocol) == false {
								var portProt PortProtocol
								portProt.Protocol = policyPortProtocol.Protocol
								portProtocols = append(portProtocols, portProt)
							}

							for i := range portProtocols {
								if portProtocols[i].Protocol == policyPortProtocol.Protocol {
									if ContainsPort(portProtocols[i].Ports, policyPort) == false {
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

// GetServiceAddress gets the slice of IP addresses in EndpointSlices for a specific
// Node using the service's Name and Namespace. It returns a list of IP addresses based
// on the specific IP version
func GetServiceAddresses(nodeName string, serviceName string, serviceNamespace string, ipVersion string) ([]string, error) {
	var (
		rEndpointSlices = kubernetes.K8sResource{
			ApiGroup:   "discovery.k8s.io",
			ApiVersion: "v1",
			Resource:   "endpointslices",
			NameSpace:  serviceNamespace,
		}
		serviceAddresses = []string{}
		serviceLabel     = fmt.Sprintf("%s=%s", kubernetes.EndpointSliceServiceNameLabel, serviceName)
	)

	endpointSlices, err := rEndpointSlices.RetrieveResourceListByLabel(serviceLabel)
	if err != nil {
		return serviceAddresses, err
	}

	for _, item := range endpointSlices.Items {
		addressType, found, err := unstructured.NestedString(item.Object, "addressType")
		if err != nil || !found {
			continue
		}
		if addressType != ipVersion {
			continue
		}
		endpoints, found, err := unstructured.NestedSlice(item.Object, "endpoints")
		if err != nil || !found {
			continue
		}

		for _, e := range endpoints {
			endpoint, ok := e.(map[string]interface{})
			if !ok {
				continue
			}

			addresses, found, err := unstructured.NestedStringSlice(endpoint, "addresses")
			if err != nil || !found {
				continue
			}

			endpointNodeName, found, err := unstructured.NestedString(endpoint, "nodeName")
			if err != nil || !found {
				continue
			}

			if endpointNodeName != nodeName {
				continue
			}

			serviceAddresses = append(serviceAddresses, addresses[0])
		}
	}

	return serviceAddresses, nil
}

// GetServiceIPFamilies returns the string slice related to IP Families described in
// a specific service
func GetServiceIPFamilies(serviceName string, serviceNamespace string) ([]string, error) {
	rService := kubernetes.K8sResource{
		ApiGroup:   "",
		ApiVersion: "v1",
		Resource:   "services",
		NameSpace:  serviceNamespace,
	}

	service, err := rService.RetrieveResourceInfoByName(serviceName)
	if err != nil {
		return nil, err
	}

	ipFamilies, found, err := unstructured.NestedStringSlice(service.Object, "spec", "ipFamilies")
	if err != nil {
		errMsg := fmt.Errorf("error retrieving IP Families: %w", err)
		return nil, errMsg
	}

	if !found {
		errMsg := fmt.Errorf("IP Families not found")
		return nil, errMsg
	}

	return ipFamilies, nil
}
