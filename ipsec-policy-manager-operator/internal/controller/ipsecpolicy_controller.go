/*
Copyright 2025.

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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	api "starlingx.windriver.com/ipsec-policy-manager-operator/api/v1"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/kubernetes"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/swanctl"
)

// IPsecPolicyReconciler reconciles a IPsecPolicy object
type IPsecPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func endpointEventReconcile(obj client.Object, client client.Client) bool {
	ctx := context.Background()

	var ipsecPoliciesList api.IPsecPolicyList
	if err := client.List(ctx, &ipsecPoliciesList); err != nil {
		fmt.Println(err, "Unable to list IPsecPolicies")
		return false
	}

	for _, ipsecPolicyConf := range ipsecPoliciesList.Items {
		for _, policy := range ipsecPolicyConf.Spec.Policies {
			if obj.GetName() == policy.ServiceName {
				fmt.Printf("Endpoints for service %s was modified. Reconciling operator. \n", policy.ServiceName)
				return true
			}
		}
	}

	return false
}

func endpointPredicate(mgr ctrl.Manager) predicate.Funcs {
	client := mgr.GetClient()

	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return endpointEventReconcile(e.ObjectNew, client)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return endpointEventReconcile(e.Object, client)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return endpointEventReconcile(e.Object, client)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list
//+kubebuilder:rbac:groups="",resources=endpoints;pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;create;delete;update;patch;watch
//+kubebuilder:rbac:groups=crd.projectcalico.org,resources=blockaffinities,verbs=get;list

//+kubebuilder:rbac:groups=starlingx.windriver.com,resources=ipsecpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=starlingx.windriver.com,resources=ipsecpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=starlingx.windriver.com,resources=ipsecpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IPsecPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *IPsecPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling IPsecPolicy custom resource")

	var err error
	var crList api.IPsecPolicyList

	// Fetch all CRs in all namespaces (NO `client.InNamespace()` filter)
	err = r.List(ctx, &crList)
	if err != nil {
		fmt.Println("Error fetching CRs:", err)
		return ctrl.Result{}, err
	}

	nodesConf, err := kubernetes.GetNodesConfiguration()
	if err != nil {
		log.Error(err, "unable to retrieve nodes configuration.")
		return ctrl.Result{}, err
	}

	var configFile *swanctl.ConfigurationFile
	for _, node := range nodesConf.Nodes {
		configFile = new(swanctl.ConfigurationFile)
		if err = configFile.Generate(node.Hostname, crList); err != nil {
			fmt.Println(node.Hostname, ": Unable to generate IPsec configuration: ", err)
			return ctrl.Result{}, err
		}
		configData, err := configFile.GetConfigData()
		if err != nil {
			fmt.Println(node.Hostname, ": Unable to retrieve IPsec configuration data: ", err)
			return ctrl.Result{}, err
		}

		// Create or update the ConfigMap
		configMapName := kubernetes.IPsecConfigMapPrefix + node.Hostname
		err = kubernetes.CreateOrUpdateConfigMap(
			r.Client, kubernetes.OperatorNamespace, configMapName, configData)
		if err != nil {
			fmt.Println(node.Hostname, ": Failed to create/update configMap: ", err)
			return ctrl.Result{}, err
		}

		configFile = nil
	}

	log.Info("Reconciling IPsecPolicy custom resource complete.")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPsecPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.IPsecPolicy{}).
		Watches(
			&corev1.Endpoints{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(endpointPredicate(mgr)),
		).
		Complete(r)
}
