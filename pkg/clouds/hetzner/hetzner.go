/*
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

package hetzner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/klog"
)

// API is the implementation of the cloud apis for Hetzner cloud
type API struct {
	Token string
}

// ErrAddrNotFound is returned when is requested an operation on a not-found address
var ErrAddrNotFound = errors.New("Floating IP not found")

// ErrServerNotFound is returned when is requested an operation on a not-found server
var ErrServerNotFound = errors.New("Hetzner cloud node not found")

// AssignIPToServer assigns a floating ip to a given server
// https://docs.hetzner.cloud/#floating-ip-actions-assign-a-floating-ip-to-a-server
func (h *API) AssignIPToServer(address, serverName string) error {
	klog.Infof("Assigning address %s hetzner cloud to node %s", address, serverName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	client := hcloud.NewClient(hcloud.WithToken(h.Token))
	ip, err := h.getIPByAddress(ctx, client, h.Token, address)
	if err != nil {
		klog.Error(err)
		return err
	}

	server, err := h.getServerByName(ctx, client, h.Token, serverName)
	if err != nil {
		klog.Error(err)
		if err != ErrServerNotFound {
			return err
		}
		return err
	}

	act, res, err := client.FloatingIP.Assign(ctx, ip, server)
	if err != nil {
		klog.Error(err)
		return err
	}

	if res.StatusCode != 200 && res.StatusCode != 201 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			err := fmt.Errorf("Something went wrong assigning ip %s to server %s, status code: %d, cannot decode body %v", address, serverName, res.StatusCode, err)
			klog.Error(err)
			return err
		}
		bodystr := buf.String()
		err := fmt.Errorf("Something went wrong assigning ip %s to server %s, status code: %d, response body is: %s", address, serverName, res.StatusCode, bodystr)
		klog.Error(err)
		return err
	}

	h.printRateLimit(res)

	klog.Infof("Adding address %s to server %s action %d is in state %s", address, serverName, act.ID, act.Status)
	return nil
}

// UnassignIP unassigns a floating ip
// https://docs.hetzner.cloud/#floating-ip-actions-unassign-a-floating-ip
func (h *API) UnassignIP(address string) error {
	klog.Infof("Unassigning address %s from hetzner cloud", address)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	client := hcloud.NewClient(hcloud.WithToken(h.Token))
	ip, err := h.getIPByAddress(ctx, client, h.Token, address)
	if err != nil {
		klog.Error(err)
		return err
	}

	act, res, err := client.FloatingIP.Unassign(ctx, ip)
	if err != nil {
		klog.Error(err)
		return err
	}

	if res.StatusCode != 200 && res.StatusCode != 201 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			err := fmt.Errorf("Something went wrong unassigning ip %s, status code: %d, cannot decode body %v", address, res.StatusCode, err)
			klog.Error(err)
			return err
		}
		bodystr := buf.String()
		err := fmt.Errorf("Something went wrong unassigning ip %s, status code: %d, response body is: %s", address, res.StatusCode, bodystr)
		klog.Error(err)
		return err
	}

	h.printRateLimit(res)

	klog.Infof("Unassigning address %s action %d is in state %s", address, act.ID, act.Status)
	return nil
}

// GetAndAssignNewAddress creates a new floating ip and assigns it to the given server
// https://docs.hetzner.cloud/#floating-ips-create-a-floating-ip
func (h *API) GetAndAssignNewAddress(serverName, ipName string) (string, error) {
	klog.Infof("Getting new address from hetzner cloud, name: %s", ipName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	client := hcloud.NewClient(hcloud.WithToken(h.Token))

	server, err := h.getServerByName(ctx, client, h.Token, serverName)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	if err != nil {
		klog.Error(err)
		if err != ErrServerNotFound {
			return "", err
		}
		return "", err
	}

	opts := hcloud.FloatingIPCreateOpts{
		Type:   hcloud.FloatingIPTypeIPv4,
		Server: server,
		Labels: map[string]string{
			"managed-by": "plenuslb",
		},
		Name: &ipName,
	}
	act, res, err := client.FloatingIP.Create(ctx, opts)
	if err != nil {
		klog.Error(err)
		return "", err
	}

	if res.StatusCode != 200 && res.StatusCode != 201 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			err := fmt.Errorf("Something went wrong getting new ip, status code: %d, cannot decode body %v", res.StatusCode, err)
			klog.Error(err)
			return "", err
		}
		bodystr := buf.String()
		err := fmt.Errorf("Something went wrong getting new ip, status code: %d, response body is: %s", res.StatusCode, bodystr)
		klog.Error(err)
		return "", err
	}

	h.printRateLimit(res)

	klog.Infof("Got new address %s action %d is in state %s", act.FloatingIP.IP, act.Action.ID, act.Action.Status)
	return act.FloatingIP.IP.String(), nil
}

// DeleteAddress deletes a floating IP from Hetzner cloud
// https://docs.hetzner.cloud/#floating-ips-delete-a-floating-ip
func (h *API) DeleteAddress(address string) error {
	klog.Infof("Deleting address %s from hetzner cloud", address)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	client := hcloud.NewClient(hcloud.WithToken(h.Token))
	ip, err := h.getIPByAddress(ctx, client, h.Token, address)
	if err != nil {
		klog.Error(err)
		return err
	}

	res, err := client.FloatingIP.Delete(ctx, ip)
	if err != nil {
		klog.Error(err)
		return err
	}

	if res.StatusCode != 204 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			err := fmt.Errorf("Something went wrong deleting ip %s, status code: %d, cannot decode body %v", address, res.StatusCode, err)
			klog.Error(err)
			return err
		}
		bodystr := buf.String()
		err := fmt.Errorf("Something went wrong deleting ip %s, status code: %d, response body is: %s", address, res.StatusCode, bodystr)
		klog.Error(err)
		return err
	}

	h.printRateLimit(res)

	klog.Infof("Deleted address %s from hetzner cloud", address)
	return nil
}

func (h *API) getServerByName(ctx context.Context, client *hcloud.Client, token, name string) (*hcloud.Server, error) {
	servers, err := client.Server.AllWithOpts(ctx, hcloud.ServerListOpts{Name: name})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, server := range servers {
		if server.Name == name {
			return server, nil
		}
	}

	return nil, ErrServerNotFound
}

func (h *API) getIPByAddress(ctx context.Context, client *hcloud.Client, token, address string) (*hcloud.FloatingIP, error) {
	ips, err := client.FloatingIP.All(ctx)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, ip := range ips {
		if ip.IP.String() == address {
			return ip, nil
		}
	}

	return nil, ErrAddrNotFound
}

func (h *API) printRateLimit(res *hcloud.Response) {
	// https://docs.hetzner.cloud/#overview-rate-limiting
	limit := res.Header.Get("RateLimit-Limit")
	remaining := res.Header.Get("RateLimit-Remaining")
	reset := res.Header.Get("RateLimit-Reset")
	i, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		klog.Error(err)
		return
	}
	tm := time.Unix(i, 0)
	msg := fmt.Sprintf("Hetzner API remaining calls is %s/%s, will be reset at %s", remaining, limit, tm)

	intLimit, _ := strconv.Atoi(limit)
	intRemaining, _ := strconv.Atoi(limit)
	remainingPercentage := (intRemaining * 100) / intLimit
	if remainingPercentage < 20 {
		klog.Error(msg)
	} else if remainingPercentage < 50 {
		klog.Warning(msg)
	} else {
		klog.Info(msg)
	}
}
