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
	IPsecConfFileName      string = "k8s-nodes.conf"
	IPsecConfFilePath      string = "/etc/swanctl/conf.d/" + IPsecConfFileName
	CertificateExtension   string = ".crt"
	SystemLocalCACert0     string = "system-local-ca-0" + CertificateExtension
	SystemLocalCACert1     string = "system-local-ca-1" + CertificateExtension
	IPsecConfigLocalMapKey string = "local_conn"
	IPsecConfigConnsMapKey string = "connections"
	IPsecConfigMapPrefix   string = "system-ipsec-configmap-"
)
