/*
Copyright (c) 2026 Wind River Systems, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kubernetes

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newFakeClient(objs ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	builder := fakeclient.NewClientBuilder().WithScheme(scheme)
	for _, o := range objs {
		builder = builder.WithObjects(o.(client.Object))
	}
	return builder.Build()
}

// --- CreateOrUpdateConfigMap ---

func TestCreateOrUpdateConfigMap_Create(t *testing.T) {
	c := newFakeClient()
	err := CreateOrUpdateConfigMap(c, "default", "test-cm", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it was created
	cm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: "test-cm", Namespace: "default"}, cm); err != nil {
		t.Fatalf("configmap not found: %v", err)
	}
	if cm.Data["key"] != "val" {
		t.Errorf("expected key=val, got %v", cm.Data)
	}
}

func TestCreateOrUpdateConfigMap_Update(t *testing.T) {
	existing := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-cm", Namespace: "default"},
		Data:       map[string]string{"key": "old"},
	}
	c := newFakeClient(existing)
	err := CreateOrUpdateConfigMap(c, "default", "test-cm", map[string]string{"key": "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: "test-cm", Namespace: "default"}, cm); err != nil {
		t.Fatalf("configmap not found: %v", err)
	}
	if cm.Data["key"] != "new" {
		t.Errorf("expected key=new, got %v", cm.Data)
	}
}

// --- DeleteConfigMap ---

func TestDeleteConfigMap_Exists(t *testing.T) {
	existing := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-cm", Namespace: "default"},
	}
	c := newFakeClient(existing)
	err := DeleteConfigMap(c, "default", "test-cm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it was deleted
	cm := &corev1.ConfigMap{}
	err = c.Get(context.Background(), client.ObjectKey{Name: "test-cm", Namespace: "default"}, cm)
	if err == nil {
		t.Error("expected configmap to be deleted")
	}
}

func TestDeleteConfigMap_NotFound(t *testing.T) {
	c := newFakeClient()
	err := DeleteConfigMap(c, "default", "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent configmap")
	}
}

// --- GetNodeNameByPodName ---

func TestGetNodeNameByPodName_Found(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: OperatorNamespace},
		Spec:       corev1.PodSpec{NodeName: "worker-0"},
	}
	c := newFakeClient(pod)
	nodeName, err := GetNodeNameByPodName(c, context.Background(), "test-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nodeName != "worker-0" {
		t.Errorf("expected worker-0, got %s", nodeName)
	}
}

func TestGetNodeNameByPodName_NotFound(t *testing.T) {
	c := newFakeClient()
	_, err := GetNodeNameByPodName(c, context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent pod")
	}
}
