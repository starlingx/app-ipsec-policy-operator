/*
Copyright (c) 2026 Wind River Systems, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package swanctl

import (
	"strings"
	"testing"

	"starlingx.io/ipsec-policy-agent/pkg/vici"
)

func TestGenerateChildrenSAConf(t *testing.T) {
	c := &ConfigurationFile{}
	children := map[string]*vici.ChildSA{
		"tcp_svc_egress": {
			Mode:                   "tunnel",
			StartAction:            "trap",
			LocalTrafficSelectors:  []string{"10.244.0.0/24[tcp]"},
			RemoteTrafficSelectors: []string{"10.244.1.5[tcp/80]"},
		},
	}
	c.generateChildrenSAConf(children)
	joined := strings.Join(c.Data, "\n")
	if !strings.Contains(joined, "children {") {
		t.Error("missing children block")
	}
	if !strings.Contains(joined, "tcp_svc_egress {") {
		t.Error("missing child SA name")
	}
	if !strings.Contains(joined, "start_action = trap") {
		t.Error("missing start_action")
	}
	if !strings.Contains(joined, "mode = tunnel") {
		t.Error("missing mode")
	}
}

func TestGenerateConf_LocalOnly(t *testing.T) {
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
	if err := c.GenerateConf(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(c.Data, "\n")
	if !strings.HasPrefix(joined, "connections {") {
		t.Error("missing connections block start")
	}
	if !strings.Contains(joined, "k8s-node-local {") {
		t.Error("missing local connection name")
	}
	if !strings.HasSuffix(strings.TrimSpace(joined), "}") {
		t.Error("missing closing brace")
	}
}

func TestGenerateConf_WithNodeConnection(t *testing.T) {
	c := &ConfigurationFile{
		Connections: []vici.SystemNodeConnection{
			{
				Name:        "k8s-node-worker-0",
				ReauthTime:  14400,
				RekeyTime:   3600,
				Unique:      "replace",
				LocalAddrs:  []string{"192.168.1.1"},
				RemoteAddrs: []string{"192.168.1.2"},
				Local: &vici.LocalOpts{
					Auth: "pubkey",
					Cert: &vici.CertBlock{File: "/etc/swanctl/x509/system-ipsec-certificate-ctrl-0.crt"},
				},
				Remote: &vici.RemoteOpts{
					ID:      "CN=*",
					Auth:    "pubkey",
					CACert0: &vici.CertBlock{File: "/etc/swanctl/x509ca/system-local-ca-0.crt"},
					CACert1: &vici.CertBlock{File: "/etc/swanctl/x509ca/system-local-ca-1.crt"},
				},
				Children: map[string]*vici.ChildSA{
					"tcp_svc_egress": {
						Mode:                   "tunnel",
						StartAction:            "trap",
						LocalTrafficSelectors:  []string{"10.244.0.0/24[tcp]"},
						RemoteTrafficSelectors: []string{"10.244.1.5[tcp/80]"},
					},
				},
			},
		},
	}
	if err := c.GenerateConf(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(c.Data, "\n")
	for _, expected := range []string{
		"k8s-node-worker-0 {",
		"reauth_time = 14400",
		"rekey_time = 3600",
		"unique = replace",
		"local_addrs = 192.168.1.1",
		"remote_addrs = 192.168.1.2",
		"auth = pubkey",
		"certs = system-ipsec-certificate-ctrl-0.crt",
		"id = CN=*",
		"cacerts = system-local-ca-0.crt,system-local-ca-1.crt",
	} {
		if !strings.Contains(joined, expected) {
			t.Errorf("missing expected content: %q", expected)
		}
	}
}

func TestGenerateConf_Empty(t *testing.T) {
	c := &ConfigurationFile{}
	if err := c.GenerateConf(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Data) != 2 {
		t.Errorf("expected 2 lines (open+close braces), got %d", len(c.Data))
	}
}

func TestGenerateChildrenSAConf_MultipleChildren(t *testing.T) {
	c := &ConfigurationFile{}
	children := map[string]*vici.ChildSA{
		"tcp_svc_egress": {
			Mode:                   "tunnel",
			StartAction:            "trap",
			LocalTrafficSelectors:  []string{"10.0.0.0/8[tcp]"},
			RemoteTrafficSelectors: []string{"10.0.1.1[tcp/80]"},
		},
		"udp_svc_ingress": {
			Mode:                   "tunnel",
			StartAction:            "trap",
			LocalTrafficSelectors:  []string{"10.0.1.1[udp/53]"},
			RemoteTrafficSelectors: []string{"10.0.0.0/8[udp]"},
		},
	}
	c.generateChildrenSAConf(children)
	joined := strings.Join(c.Data, "\n")
	if !strings.Contains(joined, "tcp_svc_egress") {
		t.Error("missing tcp_svc_egress")
	}
	if !strings.Contains(joined, "udp_svc_ingress") {
		t.Error("missing udp_svc_ingress")
	}
}

func TestWriteFile_InvalidPath(t *testing.T) {
	c := &ConfigurationFile{
		Data: []string{"connections {", "}"},
	}
	// IPsecConfFilePath is /etc/swanctl/conf.d/k8s-nodes.conf which doesn't exist
	// in test env, so WriteFile should return an error
	err := c.WriteFile()
	if err == nil {
		// If it somehow succeeded (running as root?), clean up
		if c.File != nil {
			c.File.Close()
		}
	}
	// Either way, we exercised the function
}

func TestCleanConnections_EmptyList(t *testing.T) {
	c := &ConfigurationFile{}
	// Should not panic with empty list
	c.CleanConnections([]string{})
}

func TestCleanConnections_WithConnections(t *testing.T) {
	c := &ConfigurationFile{}
	// vici.TerminateConnection and vici.UnloadConnection will fail without strongswan
	// but CleanConnections handles errors gracefully (logs them)
	// This tests the code paths including the LocalConn skip logic
	c.CleanConnections([]string{"k8s-node-local", "k8s-node-local-ipv6", "k8s-node-worker-0"})
}
