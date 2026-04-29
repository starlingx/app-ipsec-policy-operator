/*
Copyright (c) 2026 Wind River Systems, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package swanctl

import (
	"encoding/json"
	"strings"
	"testing"

	"starlingx.io/ipsec-policy-manager-operator/pkg/vici"
)

func TestMarshalLocalConn_Empty(t *testing.T) {
	c := &ConfigurationFile{}
	got := c.MarshalLocalConn()
	if got != "null" {
		t.Errorf("expected null JSON for nil slice, got %q", got)
	}
}

func TestMarshalLocalConn_WithData(t *testing.T) {
	c := &ConfigurationFile{
		LocalConn: []vici.Connection{
			{
				Name: "k8s-node-local",
				Children: map[string]*vici.ChildSA{
					"node-local-bypass": {
						Mode:                   "pass",
						StartAction:            "trap",
						LocalTrafficSelectors:  []string{"10.244.0.0/24"},
						RemoteTrafficSelectors: []string{"10.244.0.0/24"},
					},
				},
			},
		},
	}
	got := c.MarshalLocalConn()
	if got == "" {
		t.Fatal("expected non-empty JSON")
	}
	var parsed []vici.Connection
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed) != 1 || parsed[0].Name != "k8s-node-local" {
		t.Errorf("unexpected parsed result: %+v", parsed)
	}
}

func TestMarshalConnections_Empty(t *testing.T) {
	c := &ConfigurationFile{}
	got := c.MarshalConnections()
	if got != "null" {
		t.Errorf("expected null JSON for nil slice, got %q", got)
	}
}

func TestMarshalConnections_WithData(t *testing.T) {
	c := &ConfigurationFile{
		Connections: []vici.SystemNodeConnection{
			{
				Name:        "k8s-node-worker-0",
				ReauthTime:  14400,
				RekeyTime:   3600,
				Unique:      "replace",
				LocalAddrs:  []string{"192.168.1.1"},
				RemoteAddrs: []string{"192.168.1.2"},
				Children:    map[string]*vici.ChildSA{},
			},
		},
	}
	got := c.MarshalConnections()
	if !strings.Contains(got, "k8s-node-worker-0") {
		t.Errorf("expected connection name in JSON, got %q", got)
	}
}

func TestGetLocalConf_IPv4(t *testing.T) {
	c := &ConfigurationFile{
		PodSubnet: []string{"10.244.0.0/24"},
	}
	c.getLocalConf()
	if len(c.LocalConn) != 1 {
		t.Fatalf("expected 1 local conn, got %d", len(c.LocalConn))
	}
	if c.LocalConn[0].Name != "k8s-node-local" {
		t.Errorf("expected k8s-node-local, got %s", c.LocalConn[0].Name)
	}
	child := c.LocalConn[0].Children["node-local-bypass"]
	if child == nil {
		t.Fatal("expected node-local-bypass child SA")
	}
	if child.Mode != BypassMode {
		t.Errorf("expected mode %s, got %s", BypassMode, child.Mode)
	}
}

func TestGetLocalConf_IPv6(t *testing.T) {
	c := &ConfigurationFile{
		PodSubnet: []string{"fd00:10:244::/64"},
	}
	c.getLocalConf()
	if len(c.LocalConn) != 1 {
		t.Fatalf("expected 1 local conn, got %d", len(c.LocalConn))
	}
	if c.LocalConn[0].Name != "k8s-node-local-ipv6" {
		t.Errorf("expected k8s-node-local-ipv6, got %s", c.LocalConn[0].Name)
	}
}

func TestGetLocalConf_DualStack(t *testing.T) {
	c := &ConfigurationFile{
		PodSubnet: []string{"10.244.0.0/24", "fd00:10:244::/64"},
	}
	c.getLocalConf()
	if len(c.LocalConn) != 2 {
		t.Fatalf("expected 2 local conns, got %d", len(c.LocalConn))
	}
	names := map[string]bool{}
	for _, conn := range c.LocalConn {
		names[conn.Name] = true
	}
	if !names["k8s-node-local"] || !names["k8s-node-local-ipv6"] {
		t.Errorf("expected both IPv4 and IPv6 names, got %v", names)
	}
}

func TestGetConfigData(t *testing.T) {
	c := &ConfigurationFile{
		LocalConn: []vici.Connection{
			{Name: "k8s-node-local", Children: map[string]*vici.ChildSA{}},
		},
		Connections: []vici.SystemNodeConnection{
			{Name: "k8s-node-worker-0", Children: map[string]*vici.ChildSA{}},
		},
	}
	data, err := c.GetConfigData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := data["local_conn"]; !ok {
		t.Error("missing local_conn key")
	}
	if _, ok := data["connections"]; !ok {
		t.Error("missing connections key")
	}
}
