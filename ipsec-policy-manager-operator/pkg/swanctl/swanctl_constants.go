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

package swanctl

const (
	IPsecConfFileName    string = "k8s-nodes.conf"
	IPsecConfFilePath    string = "/etc/swanctl/conf.d/" + IPsecConfFileName
	IPsecCertPath        string = "/etc/swanctl/x509/"
	IPsecCertCAPath      string = "/etc/swanctl/x509ca/"
	CurrentNode          string = "controller-0"
	CertificatePrefix    string = "system-ipsec-certificate-"
	CertificateExtension string = ".crt"
	SystemLocalCACert0   string = "system-local-ca-0" + CertificateExtension
	SystemLocalCACert1   string = "system-local-ca-1" + CertificateExtension

	ReauthTime           int    = 14400
	RekeyTime            int    = 3600
	Unique               string = "replace"
	LocalOptsAuth        string = "pubkey"
	RemoteOptsID         string = "CN=*"
	RemoteOptsAuth       string = "pubkey"

	EgressMode           string = "tunnel"
	EgressStartAction    string = "trap"
	EgressUpdown         string = "/usr/lib/ipsec/_updown iptables"

	IngressMode          string = "tunnel"
	IngressStartAction   string = "trap"
	IngressUpdown        string = "/usr/lib/ipsec/_updown iptables"

	ProtocolMode         string = "tunnel"
	ProtocolStartAction  string = "trap"
	ProtocolUpdown       string = "/usr/lib/ipsec/_updown iptables"

	BypassMode           string = "pass"
	BypassStartAction    string = "trap"
)
