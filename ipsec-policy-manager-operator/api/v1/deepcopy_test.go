/*
Copyright (c) 2026 Wind River Systems, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package v1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIPsecPolicy_DeepCopy(t *testing.T) {
	orig := &IPsecPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: IPsecPolicySpec{Policies: []Policy{
			{Name: "p1", ServiceName: "svc", ServiceNS: "ns", ServicePorts: "tcp/80"},
		}},
		Status: IPsecPolicyStatus{Status: "active"},
	}
	cp := orig.DeepCopy()
	if cp.Name != orig.Name || cp.Status.Status != orig.Status.Status {
		t.Error("DeepCopy did not preserve fields")
	}
	cp.Spec.Policies[0].Name = "modified"
	if orig.Spec.Policies[0].Name == "modified" {
		t.Error("DeepCopy did not create independent copy of Policies")
	}
}

func TestIPsecPolicy_DeepCopy_Nil(t *testing.T) {
	var p *IPsecPolicy
	if p.DeepCopy() != nil {
		t.Error("DeepCopy of nil should return nil")
	}
}

func TestIPsecPolicy_DeepCopyObject(t *testing.T) {
	orig := &IPsecPolicy{ObjectMeta: metav1.ObjectMeta{Name: "test"}}
	obj := orig.DeepCopyObject()
	if obj == nil {
		t.Fatal("DeepCopyObject returned nil")
	}
	if _, ok := obj.(*IPsecPolicy); !ok {
		t.Error("DeepCopyObject did not return *IPsecPolicy")
	}
}

func TestIPsecPolicyList_DeepCopy(t *testing.T) {
	orig := &IPsecPolicyList{
		Items: []IPsecPolicy{
			{ObjectMeta: metav1.ObjectMeta{Name: "a"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "b"}},
		},
	}
	cp := orig.DeepCopy()
	if len(cp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(cp.Items))
	}
	cp.Items[0].Name = "modified"
	if orig.Items[0].Name == "modified" {
		t.Error("DeepCopy did not create independent copy of Items")
	}
}

func TestIPsecPolicyList_DeepCopy_Nil(t *testing.T) {
	var l *IPsecPolicyList
	if l.DeepCopy() != nil {
		t.Error("DeepCopy of nil should return nil")
	}
}

func TestIPsecPolicyList_DeepCopyObject(t *testing.T) {
	orig := &IPsecPolicyList{}
	obj := orig.DeepCopyObject()
	if obj == nil {
		t.Fatal("DeepCopyObject returned nil")
	}
}

func TestIPsecPolicySpec_DeepCopy(t *testing.T) {
	orig := &IPsecPolicySpec{Policies: []Policy{{Name: "p1"}}}
	cp := orig.DeepCopy()
	cp.Policies[0].Name = "modified"
	if orig.Policies[0].Name == "modified" {
		t.Error("DeepCopy did not create independent copy")
	}
}

func TestIPsecPolicySpec_DeepCopy_Nil(t *testing.T) {
	var s *IPsecPolicySpec
	if s.DeepCopy() != nil {
		t.Error("DeepCopy of nil should return nil")
	}
}

func TestIPsecPolicyStatus_DeepCopy(t *testing.T) {
	orig := &IPsecPolicyStatus{Status: "active"}
	cp := orig.DeepCopy()
	if cp.Status != "active" {
		t.Error("DeepCopy did not preserve Status")
	}
}

func TestIPsecPolicyStatus_DeepCopy_Nil(t *testing.T) {
	var s *IPsecPolicyStatus
	if s.DeepCopy() != nil {
		t.Error("DeepCopy of nil should return nil")
	}
}

func TestPolicy_DeepCopy(t *testing.T) {
	orig := &Policy{Name: "p1", ServiceName: "svc", ServiceNS: "ns", ServicePorts: "tcp/80"}
	cp := orig.DeepCopy()
	if cp.Name != "p1" || cp.ServiceName != "svc" {
		t.Error("DeepCopy did not preserve fields")
	}
	cp.Name = "modified"
	if orig.Name == "modified" {
		t.Error("DeepCopy did not create independent copy")
	}
}

func TestPolicy_DeepCopy_Nil(t *testing.T) {
	var p *Policy
	if p.DeepCopy() != nil {
		t.Error("DeepCopy of nil should return nil")
	}
}
