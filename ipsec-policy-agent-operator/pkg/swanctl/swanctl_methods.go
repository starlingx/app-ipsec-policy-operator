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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"starlingx.windriver.com/ipsec-policy-agent/pkg/vici"
)

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

func (c *ConfigurationFile) FormatTrafficSelectors(trafficSelectors []string) string {
	var protPorts []string

	ip := ""

	for _, trafficSelector := range trafficSelectors {
		parts := strings.SplitN(trafficSelector, "[", 2)
		if len(parts) != 2 || !strings.HasSuffix(parts[1], "]") {
			return trafficSelector
		}

		if ip == "" {
			ip = parts[0]
		}
		protPort := strings.TrimSuffix(parts[1], "]")
		protPorts = append(protPorts, protPort)
	}

	result := fmt.Sprintf("%s[%s]", ip, strings.Join(protPorts, ","))

	return result
}

func (c *ConfigurationFile) generateChildrenSAConf(childSAs map[string]*vici.ChildSA) {
	c.Data = append(c.Data, "\t\tchildren {")

	for name, sa := range childSAs {
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t%v {", name))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tstart_action = %v", sa.StartAction))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tlocal_ts = %v", c.FormatTrafficSelectors(sa.LocalTrafficSelectors)))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tremote_ts = %v", c.FormatTrafficSelectors(sa.RemoteTrafficSelectors)))
		c.Data = append(c.Data, fmt.Sprintf("\t\t\t\tmode = %v", sa.Mode))
		c.Data = append(c.Data, "\t\t\t}")
	}

	c.Data = append(c.Data, "\t\t}")
}

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

func (c *ConfigurationFile) WriteFile() error {
	var err error
	c.File, err = os.Create(IPsecConfFilePath)
	if err != nil {
		fmt.Println(err)
		c.File.Close()
		return err
	}

	for _, item := range c.Data {
		fmt.Fprintln(c.File, item)
	}

	return nil
}
