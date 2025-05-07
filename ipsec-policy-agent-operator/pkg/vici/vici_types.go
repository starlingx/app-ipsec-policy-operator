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

package vici

type CertBlock struct {
	File string `vici:"file"` // absolute path to certificate file
}

type LocalOpts struct {
	Auth string     `vici:"auth"`
	Cert *CertBlock `vici:"cert"`
}

type RemoteOpts struct {
	ID      string     `vici:"id"`
	Auth    string     `vici:"auth"`
	CACert0 *CertBlock `vici:"cacert0"`
	CACert1 *CertBlock `vici:"cacert1"`
}

type ChildSA struct {
	Mode                   string   `vici:"mode"`
	StartAction            string   `vici:"start_action"`
	LocalTrafficSelectors  []string `vici:"local_ts"`
	RemoteTrafficSelectors []string `vici:"remote_ts"`
	Updown                 string   `vici:"updown"`
	Inactivity             int      `vici:"inactivity"`
}

type Connection struct {
	Name     string
	Children map[string]*ChildSA `vici:"children"`
}

type SystemNodeConnection struct {
	ReauthTime  int         `vici:"reauth_time"`
	RekeyTime   int         `vici:"rekey_time"`
	Unique      string      `vici:"unique"`
	Mobike      string      `vici:"mobike"`
	DPDDelay    int         `vici:"dpd_delay"`
	DPDTimeout  int         `vici:"dpd_timeout"`
	LocalAddrs  []string    `vici:"local_addrs"`
	RemoteAddrs []string    `vici:"remote_addrs"`
	Local       *LocalOpts  `vici:"local"`
	Remote      *RemoteOpts `vici:"remote"`

	// Note: govici does not support golang interfaces
	Name     string
	Children map[string]*ChildSA `vici:"children"`
}
