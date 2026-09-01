/*
Copyright (c) 2026 Wind River Systems, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package utility

import (
	"strings"
	"testing"
)

func TestGetYamlConf(t *testing.T) {
	type sample struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}
	out, err := GetYamlConf(sample{Name: "test", Port: 80})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "name: test") || !strings.Contains(out, "port: 80") {
		t.Errorf("unexpected YAML output: %s", out)
	}
}

func TestGetYamlConf_EmptyStruct(t *testing.T) {
	out, err := GetYamlConf(struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "{}\n" {
		t.Errorf("expected empty YAML, got: %q", out)
	}
}

func TestGetIPVersion(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"192.168.1.1", "IPv4"},
		{"10.0.0.1/24", "IPv4"},
		{"fd00::1", "IPv6"},
		{"fd00::1/64", "IPv6"},
	}
	for _, tc := range tests {
		got := GetIPVersion(tc.addr)
		if got != tc.want {
			t.Errorf("GetIPVersion(%q) = %q, want %q", tc.addr, got, tc.want)
		}
	}
}

func TestGetClusterHostIP(t *testing.T) {
	addrs := []string{"192.168.1.1", "fd00::1"}
	if got := GetClusterHostIP(addrs, "IPv4"); got != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %q", got)
	}
	if got := GetClusterHostIP(addrs, "IPv6"); got != "fd00::1" {
		t.Errorf("expected fd00::1, got %q", got)
	}
	if got := GetClusterHostIP([]string{"10.0.0.1"}, "IPv6"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestGetPodSubnet(t *testing.T) {
	subnets := []string{"10.244.0.0/24", "fd00:10:244::/64"}
	if got := GetPodSubnet(subnets, "IPv4"); got != "10.244.0.0/24" {
		t.Errorf("expected 10.244.0.0/24, got %q", got)
	}
	if got := GetPodSubnet(subnets, "IPv6"); got != "fd00:10:244::/64" {
		t.Errorf("expected fd00:10:244::/64, got %q", got)
	}
	if got := GetPodSubnet([]string{"10.0.0.0/8"}, "IPv6"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestContainsPort(t *testing.T) {
	ports := []int64{80, 443, 8080}
	if !ContainsPort(ports, 443) {
		t.Error("expected true for 443")
	}
	if ContainsPort(ports, 9090) {
		t.Error("expected false for 9090")
	}
	if ContainsPort(nil, 80) {
		t.Error("expected false for nil slice")
	}
}

func TestContainsProtocol(t *testing.T) {
	pp := []PortProtocol{
		{Protocol: "tcp", Ports: []int64{80}},
		{Protocol: "udp", Ports: []int64{53}},
	}
	if !ContainsProtocol(pp, "tcp") {
		t.Error("expected true for tcp")
	}
	if ContainsProtocol(pp, "sctp") {
		t.Error("expected false for sctp")
	}
	if ContainsProtocol(nil, "tcp") {
		t.Error("expected false for nil slice")
	}
}

func TestGetPolicyPorts(t *testing.T) {
	result := GetPolicyPorts("tcp/80, tcp/443, udp/53")
	if len(result) != 2 {
		t.Fatalf("expected 2 protocols, got %d", len(result))
	}
	for _, pp := range result {
		switch pp.Protocol {
		case "tcp":
			if len(pp.Ports) != 2 || !ContainsPort(pp.Ports, 80) || !ContainsPort(pp.Ports, 443) {
				t.Errorf("tcp ports wrong: %v", pp.Ports)
			}
		case "udp":
			if len(pp.Ports) != 1 || pp.Ports[0] != 53 {
				t.Errorf("udp ports wrong: %v", pp.Ports)
			}
		default:
			t.Errorf("unexpected protocol: %s", pp.Protocol)
		}
	}
}

func TestGetPolicyPorts_DuplicatePorts(t *testing.T) {
	result := GetPolicyPorts("tcp/80, tcp/80")
	if len(result) != 1 {
		t.Fatalf("expected 1 protocol, got %d", len(result))
	}
	if len(result[0].Ports) != 1 {
		t.Errorf("expected 1 port (deduped), got %d", len(result[0].Ports))
	}
}

func TestProtectedPortsAndProtocols(t *testing.T) {
	policyPP := []PortProtocol{
		{Protocol: "tcp", Ports: []int64{80, 443}},
		{Protocol: "udp", Ports: []int64{53}},
	}
	servicePP := []PortProtocol{
		{Protocol: "tcp", Ports: []int64{80, 8080}},
	}
	result := ProtectedPortsAndProtocols("test-svc", policyPP, servicePP)
	if len(result) != 1 {
		t.Fatalf("expected 1 protocol, got %d", len(result))
	}
	if result[0].Protocol != "tcp" {
		t.Errorf("expected tcp, got %s", result[0].Protocol)
	}
	if len(result[0].Ports) != 1 || result[0].Ports[0] != 80 {
		t.Errorf("expected [80], got %v", result[0].Ports)
	}
}

func TestProtectedPortsAndProtocols_NoMatch(t *testing.T) {
	policyPP := []PortProtocol{{Protocol: "sctp", Ports: []int64{5000}}}
	servicePP := []PortProtocol{{Protocol: "tcp", Ports: []int64{80}}}
	result := ProtectedPortsAndProtocols("test-svc", policyPP, servicePP)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestProtectedPortsAndProtocols_Empty(t *testing.T) {
	result := ProtectedPortsAndProtocols("test-svc", nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}
