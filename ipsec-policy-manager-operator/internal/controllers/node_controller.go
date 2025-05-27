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
	"time"

	corev1 "k8s.io/api/core/v1"
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

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func isNodeReady(node *corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func nodePredicate() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode := e.ObjectOld.(*corev1.Node)
			newNode := e.ObjectNew.(*corev1.Node)
			ready := !isNodeReady(oldNode) && isNodeReady(newNode)
			if ready {
				log := ctrl.Log.WithName("Node")
				log.Info("Update event detected", "name", e.ObjectNew.GetName())
			}
			return ready
		},
		CreateFunc: func(e event.CreateEvent) bool {
			ready := isNodeReady(e.Object.(*corev1.Node))
			if ready {
				log := ctrl.Log.WithName("Node")
				log.Info("Create event detected", "name", e.Object.GetName())
			}
			return ready
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			log := ctrl.Log.WithName("Node")
			log.Info("Delete event detected", "name", e.Object.GetName())
			return true
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
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling Node")

	var err error
	var crList api.IPsecPolicyList

	nodeAdded := true

	// Fetch all CRs in all namespaces (NO `client.InNamespace()` filter)
	err = r.List(ctx, &crList)
	if err != nil {
		log.Error(err, "Error fetching CRs")
		return ctrl.Result{}, err
	}

	if len(crList.Items) == 0 {
		log.Info("There are no IPsecPolicies to configure")
		log.Info("Reconciling Node complete.")
		return ctrl.Result{}, err
	}

	// Remove configmap from the deteted node
	node := &corev1.Node{}
	if err = r.Get(ctx, req.NamespacedName, node); err != nil {
		if client.IgnoreNotFound(err) == nil {
			nodeAdded = false
			configMapName := kubernetes.IPsecConfigMapPrefix + req.Name
			kubernetes.DeleteConfigMap(r.Client, kubernetes.OperatorNamespace, configMapName)
		}
	}

	if nodeAdded && kubernetes.IsBlockaffinityConfigured(node.Name) == false {
		log.Info("BlockAffinity not set yet, requeuing", "Node", node.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if err = config.GenerateConf(r.Client, crList); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Reconciling Node complete.")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupNodeManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(nodePredicate()).
		Complete(r)
}
