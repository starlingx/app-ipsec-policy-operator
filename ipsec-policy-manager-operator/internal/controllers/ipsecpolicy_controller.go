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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	api "starlingx.windriver.com/ipsec-policy-manager-operator/api/v1"
	"starlingx.windriver.com/ipsec-policy-manager-operator/pkg/config"
)

// IPsecPolicyReconciler reconciles a IPsecPolicy object
type IPsecPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func IPsecPolicyPredicate() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			log := ctrl.Log.WithName("IPsecPolicy")
			log.Info("Update event detected", "name", e.ObjectNew.GetName())
			return true
		},
		CreateFunc: func(e event.CreateEvent) bool {
			log := ctrl.Log.WithName("IPsecPolicy")
			log.Info("Create event detected", "name", e.Object.GetName())
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			log := ctrl.Log.WithName("IPsecPolicy")
			log.Info("Delete event detected", "name", e.Object.GetName())
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
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
		log.Error(err, "Error fetching CRs")
		return ctrl.Result{}, err
	}

	if len(crList.Items) == 0 {
		log.Info("There are no IPsecPolicies to configure")
		log.Info("Reconciling IPsecPolicy custom resource complete.")
		return ctrl.Result{}, err
	}

	if err = config.GenerateConf(r.Client, crList); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Reconciling IPsecPolicy custom resource complete.")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPsecPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.IPsecPolicy{}).
		WithEventFilter(IPsecPolicyPredicate()).
		Complete(r)
}
