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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Policy defines the structure to configure an IPsec policy through
// the Service info and its protocols and ports to be protected.
type Policy struct {
	Name         string `json:"name"`
	ServiceName  string `json:"servicename"`
	ServiceNS    string `json:"servicens"`
	ServicePorts string `json:"serviceports"`
}

// IPsecPolicySpec defines the desired state of IPsecPolicy
type IPsecPolicySpec struct {
	Policies []Policy `json:"policies"`
}

// IPsecPolicyStatus defines the observed state of IPsecPolicy
type IPsecPolicyStatus struct {
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// IPsecPolicy is the Schema for the ipsecpolicies API
type IPsecPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPsecPolicySpec   `json:"spec,omitempty"`
	Status IPsecPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPsecPolicyList contains a list of IPsecPolicy
type IPsecPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPsecPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPsecPolicy{}, &IPsecPolicyList{})
}
