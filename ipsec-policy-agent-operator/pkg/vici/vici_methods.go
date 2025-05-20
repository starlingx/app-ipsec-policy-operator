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

import (
	"context"
	"errors"
	"fmt"

	govici "github.com/strongswan/govici/vici"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CommandRequest(command string, msg *govici.Message) (*govici.Message, error) {
	session, err := govici.NewSession()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer session.Close()

	ret, err := session.CommandRequest(command, msg)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return ret, err
}

func LoadConnections(connections ...any) ([]*govici.Message, error) {
	ctx := context.Background()
	log := log.FromContext(ctx)
	var connName string
	results := []*govici.Message{}
	session, err := govici.NewSession()
	if err != nil {
		fmt.Println(err)
		return results, err
	}
	defer session.Close()

	for _, connection := range connections {
		m := govici.NewMessage()
		switch conn := connection.(type) {
		case Connection:
			connName = conn.Name
			c, err := govici.MarshalMessage(&conn)
			if err != nil {
				log.Error(err, "Unable to Marshal message")
				break
			}

			m.Set("replace", true)
			if err := m.Set(connName, c); err != nil {
				log.Error(err, "Unable to create message of type Connection")
				break
			}
		case SystemNodeConnection:
			connName = conn.Name
			c, err := govici.MarshalMessage(&conn)
			if err != nil {
				log.Error(err, "Unable to Marshal message")
				break
			}

			m.Set("replace", true)
			if err = m.Set(connName, c); err != nil {
				log.Error(err, "Unable to create message of type SystemNodeConnection")
				break
			}
		default:
			err = errors.New("type not supported")
		}

		if err != nil {
			break
		}

		request, _ := session.CommandRequest("load-conn", m)
		if err := request.Err(); err != nil {
			log.Error(err, fmt.Sprintf("Failed load connection %s: %v\n", connName, request.Get("errmsg")))
		}

		results = append(results, request)

		log.Info(fmt.Sprintf("Connection loaded: %s\n", connName))
	}

	return results, err
}
