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

package kubernetes

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RetrieveResourceInfo gets resource information from kubernetes and returns
// a list with the resources and/or an error in case of failure
func (r *K8sResource) RetrieveResourceInfo() (*unstructured.UnstructuredList, error) {
	var config *rest.Config
	var err error

	/* When the operator is running inside a pod, we have all env
	   variables configured, but to run in a dev environment, we need
	   to read the kubeconfig */
	if _, inCluster := os.LookupEnv(KubernetesServiceHost); inCluster {
		// Running inside the cluster
		config, _ = rest.InClusterConfig()
	} else {
		// Running outside the cluster, use kubeconfig
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, _ = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Create a dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error getting kubernetes conf: %w", err)
	}

	// Get the resource client for the given API group
	resourceClient := client.Resource(
		schema.GroupVersionResource{
			Group:    r.ApiGroup,
			Version:  r.ApiVersion,
			Resource: r.Resource,
		},
	).Namespace(r.NameSpace)

	// Retrieve data from the resource
	results, err := resourceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Error retrieving data from the resource: %w", err)
	}

	return results, nil
}

// RetrieveResourceInfoByName gets resource information from kubernetes and returns
// the resource specified by name and/or an error in case of failure
func (r *K8sResource) RetrieveResourceInfoByName(name string) (*unstructured.Unstructured, error) {
	var config *rest.Config
	var err error

	/* When the operator is running inside a pod, we have all env
	   variables configured, but to run in a dev environment, we need
	   to read the kubeconfig */
	if _, inCluster := os.LookupEnv(KubernetesServiceHost); inCluster {
		// Running inside the cluster
		config, _ = rest.InClusterConfig()
	} else {
		// Running outside the cluster, use kubeconfig
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, _ = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Create a dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error getting kubernetes conf: %w", err)
	}

	// Get the resource client for the given API group
	resourceClient := client.Resource(
		schema.GroupVersionResource{
			Group:    r.ApiGroup,
			Version:  r.ApiVersion,
			Resource: r.Resource,
		},
	).Namespace(r.NameSpace)

	// Retrieve data from the resource
	results, err := resourceClient.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Error retrieving data from the resource: %w", err)
	}

	return results, nil
}

// RetrieveResourceListByLabel gets the list of resource items in cluster by label
// selector. Returns an error in case of failure.
func (r *K8sResource) RetrieveResourceListByLabel(label string) (*unstructured.UnstructuredList, error) {
	var config *rest.Config
	var err error

	if _, inCluster := os.LookupEnv(KubernetesServiceHost); inCluster {
		// Running inside the cluster
		config, _ = rest.InClusterConfig()
	} else {
		// Running outside the cluster, use kubeconfig
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, _ = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Create a dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error getting kubernetes conf: %w", err)
	}

	// Get the resource client for the given API group
	resourceClient := client.Resource(
		schema.GroupVersionResource{
			Group:    r.ApiGroup,
			Version:  r.ApiVersion,
			Resource: r.Resource,
		},
	).Namespace(r.NameSpace)

	// Retrieve data from the resource
	results, err := resourceClient.List(context.TODO(), metav1.ListOptions {
        LabelSelector: label,
    })
	if err != nil {
		return nil, fmt.Errorf("Error retrieving data from the resource: %w", err)
	}

	return results, nil
}

func GetNodeNameByPodName(client client.Client, ctx context.Context, podName string) (string, error) {
	var pod corev1.Pod
	if err := client.Get(ctx, types.NamespacedName{
		Name: podName,
		Namespace: OperatorNamespace,
	}, &pod); err != nil {
		return "", fmt.Errorf("unable to retrieve Pod info. Error: %w", err)
	}
	return pod.Spec.NodeName, nil
}

func GetCurrentNodeConfiguration(currentNodeName string) (NodeInfo, error) {
	ctx := context.Background()
	log := log.FromContext(ctx)

	// Get nodes in the cluster
	var (
		currentNode unstructured.Unstructured
		rNode = K8sResource{
			ApiGroup:   "",
			ApiVersion: "v1",
			Resource:   "nodes",
			NameSpace:  "",
		}
		nodeInfo NodeInfo
	)

	nodes, err := rNode.RetrieveResourceInfo()
	if err != nil {
		return nodeInfo, err
	}

	// Print the info of the nodes
	for _, node := range nodes.Items {
		// Node name
		nodeName, found, err := unstructured.NestedString(node.Object, "metadata", "name")
		if err != nil || !found {
			log.Error(err, "Error or name not found")
			continue
		}
		if nodeName == currentNodeName {
			currentNode = node
			break
		}
		//fmt.Println("node name:", nodeName)
	}
	nodeInfo.Hostname = currentNodeName

	// Node cluster host IPs
	hostIPs, found, err := unstructured.NestedSlice(currentNode.Object, "status", "addresses")
	if err != nil || !found {
		return NodeInfo{}, fmt.Errorf("Error or addresses not found for current node: %w", err)
	}
	//map[address:192.168.206.2 type:InternalIP] map[address:controller-0 type:Hostname]
	for _, hostIP := range hostIPs {
		if nodeMap, ret := hostIP.(map[string]interface{}); ret {
			if nodeMap["type"] == "InternalIP" {
				nodeInfo.ClusterHostAddr = append(nodeInfo.ClusterHostAddr, nodeMap["address"].(string))
			}
		}
	}

	// Get blockaffinities in the cluster
	rBlock := K8sResource{
		ApiGroup:   "crd.projectcalico.org",
		ApiVersion: "v1",
		Resource:   "blockaffinities",
		NameSpace:  "",
	}
	blockaffinities, err := rBlock.RetrieveResourceInfo()
	if err != nil {
		return nodeInfo, err
	}

	// Print the info of the blockaffinities
	for _, blockaffinity := range blockaffinities.Items {
		// Node
		node, found, err := unstructured.NestedString(blockaffinity.Object, "spec", "node")
		if err != nil || !found {
			log.Error(err, "Error or node not found")
			continue
		}

		if node != nodeInfo.Hostname {
			continue
		}

		// Node cidr
		cidr, found, err := unstructured.NestedString(blockaffinity.Object, "spec", "cidr")
		if err != nil || !found {
			log.Error(err, "Error or cidr not found")
			continue
		}

		nodeInfo.PodSubnet = append(nodeInfo.PodSubnet, cidr)
	}

	return nodeInfo, nil
}

func GetNodesConfiguration() (NodesInfo, error) {
	ctx := context.Background()
	log := log.FromContext(ctx)

	// Get nodes in the cluster
	var (
		rNode = K8sResource{
			ApiGroup:   "",
			ApiVersion: "v1",
			Resource:   "nodes",
			NameSpace:  "",
		}
		nodesInfo NodesInfo
	)

	nodes, err := rNode.RetrieveResourceInfo()
	if err != nil {
		return nodesInfo, err
	}

	// Print the info of the nodes
	for _, node := range nodes.Items {
		var nodeInfo NodeInfo

		// Node name
		nodeName, found, err := unstructured.NestedString(node.Object, "metadata", "name")
		if err != nil || !found {
			log.Error(err, "Error or name not found")
			continue
		}
		//fmt.Println("node name:", nodeName)

		nodeInfo.Hostname = nodeName

		// Node cluster host IPs
		hostIPs, found, err := unstructured.NestedSlice(node.Object, "status", "addresses")
		if err != nil || !found {
			log.Error(err, "Error or addresses not found")
			continue
		}
		//map[address:192.168.206.2 type:InternalIP] map[address:controller-0 type:Hostname]
		for _, hostIP := range hostIPs {
			if nodeMap, ret := hostIP.(map[string]interface{}); ret {
				if nodeMap["type"] == "InternalIP" {
					nodeInfo.ClusterHostAddr = append(nodeInfo.ClusterHostAddr, nodeMap["address"].(string))
				}
			}
		}

		// Get blockaffinities in the cluster
		rBlock := K8sResource{
			ApiGroup:   "crd.projectcalico.org",
			ApiVersion: "v1",
			Resource:   "blockaffinities",
			NameSpace:  "",
		}
		blockaffinities, err := rBlock.RetrieveResourceInfo()
		if err != nil {
			return nodesInfo, err
		}

		// Print the info of the blockaffinities
		for _, blockaffinity := range blockaffinities.Items {
			// Node
			node, found, err := unstructured.NestedString(blockaffinity.Object, "spec", "node")
			if err != nil || !found {
				log.Error(err, "Error or node not found")
				continue
			}
			//fmt.Println("node:", node)

			if node != nodeInfo.Hostname {
				continue
			}

			// Node cidr
			cidr, found, err := unstructured.NestedString(blockaffinity.Object, "spec", "cidr")
			if err != nil || !found {
				log.Error(err, "Error or cidr not found")
				continue
			}
			//fmt.Println("node cidr:", cidr)

			nodeInfo.PodSubnet = append(nodeInfo.PodSubnet, cidr)
		}

		nodesInfo.Nodes = append(nodesInfo.Nodes, nodeInfo)
	}

	return nodesInfo, nil
}

// CreateOrUpdateConfigMap creates or updates a ConfigMap in the given namespace
func CreateOrUpdateConfigMap(k8sClient client.Client, namespace string, name string, data map[string]string) error {
	ctx := context.Background()
	log := log.FromContext(ctx)

	// Define the ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	// Check if the ConfigMap already exists
	existingConfigMap := &corev1.ConfigMap{}
	objKey := client.ObjectKey{Name: name, Namespace: namespace}
	err := k8sClient.Get(ctx, objKey, existingConfigMap)
	if err == nil {
		// If it exists, update the data
		existingConfigMap.Data = data
		if updateErr := k8sClient.Update(ctx, existingConfigMap); updateErr != nil {
			return fmt.Errorf("failed to update ConfigMap: %w", updateErr)
		}
		log.Info("Updated existing ConfigMap:", "name", name)
		return nil
	}

	// If not found, create the ConfigMap
	if createErr := k8sClient.Create(ctx, configMap); createErr != nil {
		return fmt.Errorf("failed to create ConfigMap: %w", createErr)
	}

	log.Info("Created new ConfigMap", "name", name)
	return nil
}

// IsBlockaffinityConfigured validates if blockaffinities is already configured
// to a specific node and returns true or false.
func IsBlockaffinityConfigured(nodeName string) bool {
	ctx := context.Background()
	log := log.FromContext(ctx)

	// Get blockaffinities in the cluster
	rBlock := K8sResource{
		ApiGroup:   "crd.projectcalico.org",
		ApiVersion: "v1",
		Resource:   "blockaffinities",
		NameSpace:  "",
	}
	blockaffinities, err := rBlock.RetrieveResourceInfo()
	if err != nil {
		log.Error(err, "Error retrieving blockaffinities information")
		return false
	}

	// Print the info of the blockaffinities
	for _, blockaffinity := range blockaffinities.Items {
		// Node
		node, found, err := unstructured.NestedString(blockaffinity.Object, "spec", "node")
		if err != nil {
			log.Error(err, "Error getting node blockaffinities")
			continue
		}

		if !found {
			log.Info("Node not found")
			continue
		}

		if node != nodeName {
			continue
		}

		return true
	}

	return false
}

// DeleteConfigMap deletes a ConfigMap in the given namespace
func DeleteConfigMap(k8sClient client.Client, namespace string, name string) error {
	ctx := context.Background()
	log := log.FromContext(ctx)

	configMap := &corev1.ConfigMap{}
	objKey := client.ObjectKey{Name: name, Namespace: namespace}
	err := k8sClient.Get(ctx, objKey, configMap)
	if err != nil {
		return err
	}

	if deleteErr := k8sClient.Delete(ctx, configMap); deleteErr != nil {
		return fmt.Errorf("failed to delete ConfigMap: %w", deleteErr)
	}

	log.Info("Deleted configmap", "name", name, "namespace:", namespace)
	return nil
}
