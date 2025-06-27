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

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	api "starlingx.windriver.com/ipsec-policy-manager-operator/api/v1"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/config"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/kubernetes"
)

// EndpointSliceReconciler reconciles a Node object
type EndpointSliceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func endpointSliceEventReconcile(obj client.Object, client client.Client, event string) bool {
	ctx := context.Background()
	log := log.FromContext(ctx)

	var ipsecPoliciesList api.IPsecPolicyList
	if err := client.List(ctx, &ipsecPoliciesList); err != nil {
		log.Error(err, "Unable to list IPsecPolicies")
		return false
	}

	objServiceName := (obj.GetLabels())[kubernetes.EndpointSliceServiceNameLabel]

	for _, ipsecPolicyConf := range ipsecPoliciesList.Items {
		for _, policy := range ipsecPolicyConf.Spec.Policies {
			if objServiceName == policy.ServiceName {
				logMsg := fmt.Sprintf("EndpointSlices for service was %s. Reconciling operator", event)
				log.Info(logMsg, "ServiceName", policy.ServiceName)
				return true
			}
		}
	}

	return false
}

func endpointSlicePredicate(mgr ctrl.Manager) predicate.Funcs {
	client := mgr.GetClient()

	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return endpointSliceEventReconcile(e.ObjectNew, client, "updated")
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return endpointSliceEventReconcile(e.Object, client, "created")
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return endpointSliceEventReconcile(e.Object, client, "deleted")
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IPsecPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *EndpointSliceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling EndpointSlice")

	var err error
	var crList api.IPsecPolicyList

	// Fetch all CRs in all namespaces (NO `client.InNamespace()` filter)
	err = r.List(ctx, &crList)
	if err != nil {
		log.Error(err, "Error fetching CRs")
		return ctrl.Result{}, err
	}

	if err = config.GenerateConf(r.Client, crList); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Reconciling EndpointSlice complete.")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EndpointSliceReconciler) SetupEndpointSliceManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.EndpointSlice{}).
		WithEventFilter(endpointSlicePredicate(mgr)).
		Complete(r)
}
