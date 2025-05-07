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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"starlingx.windriver.com/ipsec-policy-agent/pkg/swanctl"
)

const OperatorNamespace string = "ipsec-policy-operator"

// IPsecPolicyAgentReconciler reconciles a IPsecPolicyAgent object
type IPsecPolicyAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// getNodeNameByPodName returns a double-quoted Go string literal representing the Node name
// that is associated with a specific Pod.
//
// If it is not possible to capture Pod info in order to retrieve the Node name, this function
// returns an empty Go string literal along with the error.
func getNodeNameByPodName(clientObj client.Client, ctx context.Context, podName string) (string, error) {
	var pod corev1.Pod
	if err := clientObj.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: OperatorNamespace,
	}, &pod); err != nil {
		return "", err
	}
	return pod.Spec.NodeName, nil
}

// getNodeConfigMapName returns the configmap name related of this node
func getNodeConfigMapName(clientObj client.Client) string {
	ctx := context.Background()
	logger := log.FromContext(ctx)
	var nodeConfigMapName string

	hostname, err := os.Hostname()
	if err != nil {
		logger.Error(err, "Unable to retrieve hostname. Error: ")
		return nodeConfigMapName
	}

	nodeName, err := getNodeNameByPodName(clientObj, ctx, hostname)
	if err != nil {
		logger.Error(err, "Unable to retrieve node name. Error: ")
		return nodeConfigMapName
	}

	nodeConfigMapName = swanctl.IPsecConfigMapPrefix + nodeName

	return nodeConfigMapName
}

func isConfigMapModifiedForThisNode(obj client.Object, clientObj client.Client, eventAction string) bool {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	nodeConfigMapName := getNodeConfigMapName(clientObj)
	if obj.GetName() != nodeConfigMapName {
		return false
	}

	infoMsg := fmt.Sprintf("ConfigMap %s was %s. Reconciling IPsec Policies...",
		nodeConfigMapName, eventAction)
	logger.Info(infoMsg)
	return true
}

func ipsecConfigMapPredicate(mgr ctrl.Manager) predicate.Predicate {
	clientObj := mgr.GetClient()
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isConfigMapModifiedForThisNode(e.Object, clientObj, "created")
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isConfigMapModifiedForThisNode(e.ObjectNew, clientObj, "updated")
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isConfigMapModifiedForThisNode(e.Object, clientObj, "deleted")
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IPsecPolicyAgent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *IPsecPolicyAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var config *rest.Config
	var configFile *swanctl.ConfigurationFile
	var err error

	// When the operator is running inside a pod, we have all env
	// variables configured, but to run in a dev environment, we need
	// to read the kubeconfig
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster {
		// Running inside the cluster
		config, _ = rest.InClusterConfig()
	} else {
		// Running outside the cluster, use kubeconfig
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, _ = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		return ctrl.Result{}, err
	}

	configMapName := getNodeConfigMapName(r.Client)

	configMapResource, err := clientset.CoreV1().ConfigMaps(OperatorNamespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Failed to get configmap:", err)
		return ctrl.Result{}, err
	}

	localConnData, ok := configMapResource.Data[swanctl.IPsecConfigLocalMapKey]
	if !ok {
		fmt.Printf("Key %s not found in configmap %s", swanctl.IPsecConfigLocalMapKey, configMapName)
		return ctrl.Result{}, err
	}

	connectionsData, ok := configMapResource.Data[swanctl.IPsecConfigConnsMapKey]
	if !ok {
		fmt.Printf("Key %s not found in configmap %s", swanctl.IPsecConfigConnsMapKey, configMapName)
		return ctrl.Result{}, err
	}

	configFile = new(swanctl.ConfigurationFile)

	err = json.Unmarshal([]byte(localConnData), &configFile.LocalConn)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON: ", err)
		return ctrl.Result{}, err
	}

	err = json.Unmarshal([]byte(connectionsData), &configFile.Connections)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON: ", err)
		return ctrl.Result{}, err
	}

	configFile.GenerateConf()
	configFile.WriteFile()
	configFile.LoadConnections()

	fmt.Printf("Config written to %s\n", swanctl.IPsecConfFilePath)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPsecPolicyAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(ipsecConfigMapPredicate(mgr)).
		Complete(r)
}
