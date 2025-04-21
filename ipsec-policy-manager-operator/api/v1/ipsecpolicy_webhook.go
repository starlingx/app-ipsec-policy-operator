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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var ipsecpolicylog = logf.Log.WithName("ipsecpolicy-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *IPsecPolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
//+kubebuilder:webhook:path=/validate-starlingx-windriver-com-v1-ipsecpolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=starlingx.windriver.com,resources=ipsecpolicies,verbs=create;update,versions=v1,name=vipsecpolicy.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &IPsecPolicy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *IPsecPolicy) ValidateCreate() (admission.Warnings, error) {
	ipsecpolicylog.Info("validate create", "name", r.Name)
	ipsecpolicylog.Info("validate create", "spec", r.Spec.Policies)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *IPsecPolicy) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	ipsecpolicylog.Info("validate update", "name", r.Name)
	ipsecpolicylog.Info("validate update", "new spec", r.Spec.Policies)

	oldPolicy := old.(*IPsecPolicy)
	ipsecpolicylog.Info("validate update", "old spec", oldPolicy.Spec.Policies)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *IPsecPolicy) ValidateDelete() (admission.Warnings, error) {
	ipsecpolicylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
