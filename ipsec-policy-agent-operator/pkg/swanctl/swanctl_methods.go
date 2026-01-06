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

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"starlingx.io/ipsec-policy-agent/pkg/vici"
)

const (
	LocalConn     = "k8s-node-local"
	LocalConnIPv6 = "k8s-node-local-ipv6"
)

// CleanConnections terminates the SAs and unloads all the connections
// specified by the connections list
func (c *ConfigurationFile) CleanConnections(connections []string) {
	ctx := context.Background()
	log := log.FromContext(ctx)

	for _, conn := range connections {
		if conn != LocalConn && conn != LocalConnIPv6 {
			if err := vici.TerminateConnection(conn); err != nil {
				logMsg := fmt.Sprintf("Warning: Connection %s: %s", conn, err.Error())
				log.Info(logMsg)
			}
		}

		if err := vici.UnloadConnection(conn); err != nil {
			errMsg := fmt.Sprintf("Connection %s", conn)
			log.Error(err, errMsg)
		}
	}
}

// LoadConnections loads all the connections from the struct obtained
// by the configmap
func (c *ConfigurationFile) LoadConnections() error {
	var err error

	for _, local := range c.LocalConn {
		_, err = vici.LoadConnections(local)
	}

	for _, conn := range c.Connections {
		_, err = vici.LoadConnections(conn)
	}

	return err
}

// generateChildrenSAConf generates children SAs configuration to be written in
// the IPSec conf file
func (c *ConfigurationFile) generateChildrenSAConf(childSAs map[string]*vici.ChildSA) {
	c.Data = append(c.Data, "\t\tchildren {")

	for name, sa := range childSAs {
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t%v {", name))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tstart_action = %v", sa.StartAction))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tlocal_ts = %v", strings.Join(sa.LocalTrafficSelectors, ",")))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tremote_ts = %v", strings.Join(sa.RemoteTrafficSelectors, ",")))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tmode = %v", sa.Mode))
		c.Data = append(c.Data, "\t\t\t}")
	}

	c.Data = append(c.Data, "\t\t}")
}

// GenerateConf generates the IPSec configuration to be written in the IPSec
// conf file
func (c *ConfigurationFile) GenerateConf() error {
	var err error

	c.Data = append(c.Data, "connections {")

	for _, local := range c.LocalConn {
		c.Data = append(c.Data, fmt.Sprintf("\t%v {", local.Name))
		c.generateChildrenSAConf(local.Children)
		c.Data = append(c.Data, "\t}")
	}

	for _, n := range c.Connections {
		c.Data = append(c.Data, fmt.Sprintf("\t%v {", n.Name))
		c.Data = append(c.Data, fmt.Sprintf("\t\treauth_time = %v", n.ReauthTime))
		c.Data = append(c.Data, fmt.Sprintf("\t\trekey_time = %v", n.RekeyTime))
		c.Data = append(c.Data, fmt.Sprintf("\t\tunique = %v", n.Unique))
		c.Data = append(c.Data, fmt.Sprintf("\t\tlocal_addrs = %v", n.LocalAddrs[0]))
		c.Data = append(c.Data, fmt.Sprintf("\t\tremote_addrs = %v", n.RemoteAddrs[0]))

		c.Data = append(c.Data, "\t\tlocal {")
		c.Data = append(c.Data, fmt.Sprintf("\t\t\tauth = %v", n.Local.Auth))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\tcerts = %v", filepath.Base(n.Local.Cert.File)))
		c.Data = append(c.Data, "\t\t}")

		c.Data = append(c.Data, "\t\tremote {")
		c.Data = append(c.Data, fmt.Sprintf("\t\t\tid = %v", n.Remote.ID))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\tauth = %v", n.Remote.Auth))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\tcacerts = %v,%v", SystemLocalCACert0, SystemLocalCACert1))
		c.Data = append(c.Data, "\t\t}")

		c.generateChildrenSAConf(n.Children)

		c.Data = append(c.Data, "\t}")
	}

	c.Data = append(c.Data, "}")

	return err
}

// WriteFile writes the IPSec configuration file
// in the specific directory
func (c *ConfigurationFile) WriteFile() error {
	var err error
	c.File, err = os.Create(IPsecConfFilePath)
	if err != nil {
		fmt.Println(err)
		_ = c.File.Close()
		return err
	}

	for _, item := range c.Data {
		if _, err := fmt.Fprintln(c.File, item); err != nil {
			return err
		}
	}

	return nil
}
