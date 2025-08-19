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

package webhook

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	api "starlingx.io/ipsec-policy-manager-operator/api/v1"
	"starlingx.io/ipsec-policy-manager-operator/pkg/utility"
)

// log is for logging in this package.
var ipsecpolicylog = logf.Log.WithName("ipsecpolicy-resource")

type IPsecPolicyValidator struct {
	Client client.Client
}

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *IPsecPolicyValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&api.IPsecPolicy{}).
		WithValidator(r).
		Complete()
}

// isDuplicateService reports whether there is an existing policy defined for
// this service or not
func isDuplicateService(newPolicy api.Policy, policiesList api.IPsecPolicyList) (bool, error) {
	for _, policies := range policiesList.Items {
		for _, policy := range policies.Spec.Policies {
			if policy.ServiceName == newPolicy.ServiceName && policy.ServiceNS == newPolicy.ServiceNS {
				err := fmt.Errorf("an IPsecPolicy for this Service already exists: %s", policy.Name)
				return true, err
			}
		}
	}

	return false, nil
}

// isValidProtocol reports whether is a valid protocol (UDP/TCP) or not
func isValidProtocol(newPolicy api.Policy) (bool, error) {
	for _, servicePort := range strings.Split(newPolicy.ServicePorts, ",") {
		servicePort = strings.Trim(servicePort, " ")
		portInfo := strings.Split(servicePort, "/")
		protocol := strings.ToLower(portInfo[0])

		if protocol != "udp" && protocol != "tcp" {
			err := fmt.Errorf("%s - Protocol not valid: %s", newPolicy.Name, protocol)
			return false, err
		}
	}

	return true, nil
}

// isValidPort reports whether is a valid port (1..65535) or not
func isValidPort(newPolicy api.Policy) (bool, error) {
	for _, servicePort := range strings.Split(newPolicy.ServicePorts, ",") {
		servicePort = strings.Trim(servicePort, " ")
		portInfo := strings.Split(servicePort, "/")
		port, _ := strconv.ParseInt(portInfo[1], 10, 64)

		if port < 1 || port > 65535 {
			err := fmt.Errorf("%s - Port not valid: %s", newPolicy.Name, portInfo[1])
			return false, err
		}
	}

	return true, nil
}

// isValidServicePortAndProtocol reports whether the ports and protocols
// specified by the user are configured in the service or not
func isValidServicePortAndProtocol(newPolicy api.Policy) (bool, error) {
	policyPortProtocols := utility.GetPolicyPorts(newPolicy.ServicePorts)

	servicePortProtocols, err := utility.GetServicePorts(newPolicy.ServiceName, newPolicy.ServiceNS)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return true, nil
		}
		errMsg := fmt.Errorf("Sevice: %v - Namespace: %v - Unable to retrieve service's" +
		                     "port and protocol: %w", newPolicy.ServiceName, newPolicy.ServiceNS, err)
		return false, errMsg
	}

	for _, policyPortProtocol := range policyPortProtocols {
		if utility.ContainsProtocol(servicePortProtocols, policyPortProtocol.Protocol) == false {
			errMsg := fmt.Errorf("Service: %v - is not running on Protocol: %v\n",
				newPolicy.ServiceName, policyPortProtocol.Protocol)
			return false, errMsg
		}

		for _, servicePortProtocol := range servicePortProtocols {
			if policyPortProtocol.Protocol == servicePortProtocol.Protocol {
				for _, policyPort := range policyPortProtocol.Ports {
					if utility.ContainsPort(servicePortProtocol.Ports, policyPort) == false {
						errMsg := fmt.Errorf("Service: %v - is not running on Protocol/Port: %v/%v\n",
							newPolicy.ServiceName, policyPortProtocol.Protocol, policyPort)
						return false, errMsg
					}
				}
			}
		}
	}

	return true, nil
}

//+kubebuilder:webhook:path=/validate-starlingx-io-v1-ipsecpolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=starlingx.io,resources=ipsecpolicies,verbs=create;update,versions=v1,name=vipsecpolicy.kb.io,admissionReviewVersions=v1

// ValidateCreate checks for existing objects with the same name.
func (r *IPsecPolicyValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	newPolicies, ok := obj.(*api.IPsecPolicy)
	if !ok {
		return nil, fmt.Errorf("expected IPsecPolicy, got %T", obj)
	}

	var policiesList api.IPsecPolicyList
	if err := r.Client.List(ctx, &policiesList); err != nil {
		return nil, err
	}

	for _, newPolicy := range newPolicies.Spec.Policies {
		if ret, err := isDuplicateService(newPolicy, policiesList); ret == true {
			return nil, err
		}

		if ret, err := isValidProtocol(newPolicy); ret == false {
			return nil, err
		}

		if ret, err := isValidPort(newPolicy); ret == false {
			return nil, err
		}

		if ret, err := isValidServicePortAndProtocol(newPolicy); ret == false {
			return nil, err
		}
	}

	ipsecpolicylog.Info("Validated create for", "name", newPolicies.Name)

	return nil, nil
}

// ValidateUpdate ensures updates are valid.
func (r *IPsecPolicyValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newPolicies, ok := newObj.(*api.IPsecPolicy)
	if !ok {
		return nil, fmt.Errorf("expected IPsecPolicy, got %T", newObj)
	}

	for _, newPolicy := range newPolicies.Spec.Policies {
		if ret, err := isValidProtocol(newPolicy); ret == false {
			return nil, err
		}

		if ret, err := isValidPort(newPolicy); ret == false {
			return nil, err
		}

		if ret, err := isValidServicePortAndProtocol(newPolicy); ret == false {
			return nil, err
		}
	}

	ipsecpolicylog.Info("Validated update for", "name", newPolicies.Name)

	return nil, nil
}

// ValidateDelete always allows deletion (or you can add checks).
func (r *IPsecPolicyValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	policies, ok := obj.(*api.IPsecPolicy)
	if !ok {
		return nil, fmt.Errorf("expected IPsecPolicy, got %T", obj)
	}

	ipsecpolicylog.Info("Validated delete for", "name", policies.Name)

	return nil, nil
}
