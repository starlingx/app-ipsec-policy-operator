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
	"time"

	govici "github.com/strongswan/govici/vici"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// TerminateConnection terminates a connection specified by a name
func TerminateConnection(name string) error {
	session, err := govici.NewSession()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := govici.NewMessage()
	if err := msg.Set("ike", name); err != nil {
		return err
	}

	if _, err := session.Call(ctx, "terminate", msg); err != nil {
		return fmt.Errorf("terminate failed: %w", err)
	}

	return nil
}

// UnloadConnection unloads a connection specified by a name
func UnloadConnection(name string) error {
	session, err := govici.NewSession()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := govici.NewMessage()
	if err := msg.Set("name", name); err != nil {
		return err
	}

	if _, err := session.Call(ctx, "unload-conn", msg); err != nil {
		return fmt.Errorf("unload-conn failed: %w", err)
	}

	return nil
}

// LoadConnections loads all the connections specified by a struct connections
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
	defer func() {
		_ = session.Close()
	}()

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

			if err := m.Set("replace", true); err != nil {
				log.Error(err, "Unable to create message of type Connection")
				break
			}

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

			if err := m.Set("replace", true); err != nil {
				log.Error(err, "Unable to create message of type SystemNodeConnection")
				break
			}
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		request, _ := session.Call(ctx, "load-conn", m)
		if err := request.Err(); err != nil {
			log.Error(err, fmt.Sprintf("Failed load connection %s: %v\n", connName, request.Get("errmsg")))
		}

		results = append(results, request)

		log.Info(fmt.Sprintf("Connection loaded: %s\n", connName))
	}

	return results, err
}
