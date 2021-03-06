/*
Copyright 2018 The Kubernetes Authors.

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

package config

import (
	"fmt"

	"sigs.k8s.io/kind/pkg/util"
)

// Validate returns a ConfigErrors with an entry for each problem
// with the config, or nil if there are none
func (c *Config) Validate() error {
	errs := []error{}

	// All nodes in the config should be valid
	for i, n := range c.Nodes {
		if err := n.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("please fix invalid configuration for node %d: \n%v", i, err))
		}
	}

	// There should be at least one control plane
	if c.BootStrapControlPlane() == nil {
		errs = append(errs, fmt.Errorf("please add at least one node with role %q", ControlPlaneRole))
	}
	// There should be one load balancer if more than one control plane exists in the cluster
	if len(c.ControlPlanes()) > 1 && c.ExternalLoadBalancer() == nil {
		errs = append(errs, fmt.Errorf("please add a node with role %s because in the cluster there are more than one node with role %s", ExternalLoadBalancerRole, ControlPlaneRole))
	}

	if len(errs) > 0 {
		return util.NewErrors(errs)
	}
	return nil
}

// Validate returns a ConfigErrors with an entry for each problem
// with the Node, or nil if there are none
func (n *Node) Validate() error {
	errs := []error{}

	// validate node role should be one of the expected values
	switch n.Role {
	case ControlPlaneRole,
		WorkerRole,
		ExternalEtcdRole,
		ExternalLoadBalancerRole:
	default:
		errs = append(errs, fmt.Errorf("role is a required field"))
	}

	// image should be defined
	if n.Image == "" {
		errs = append(errs, fmt.Errorf("image is a required field"))
	}

	// replicas >= 0
	if n.Replicas != nil && int32(*n.Replicas) < 0 {
		errs = append(errs, fmt.Errorf("replicas number should not be a negative number"))
	}

	// validate NodeLifecycle
	if n.ControlPlane != nil {
		if n.ControlPlane.NodeLifecycle != nil {
			for _, hook := range n.ControlPlane.NodeLifecycle.PreBoot {
				if len(hook.Command) == 0 {
					errs = append(errs, fmt.Errorf(
						"preBoot hooks must set command to a non-empty value",
					))
					// we don't need to repeat this error and we don't
					// have any others for this field
					break
				}
			}
			for _, hook := range n.ControlPlane.NodeLifecycle.PreKubeadm {
				if len(hook.Command) == 0 {
					errs = append(errs, fmt.Errorf(
						"preKubeadm hooks must set command to a non-empty value",
					))
					// we don't need to repeat this error and we don't
					// have any others for this field
					break
				}
			}
			for _, hook := range n.ControlPlane.NodeLifecycle.PostKubeadm {
				if len(hook.Command) == 0 {
					errs = append(errs, fmt.Errorf(
						"postKubeadm hooks must set command to a non-empty value",
					))
					// we don't need to repeat this error and we don't
					// have any others for this field
					break
				}
			}
			for _, hook := range n.ControlPlane.NodeLifecycle.PostSetup {
				if len(hook.Command) == 0 {
					errs = append(errs, fmt.Errorf(
						"postKubeadm hooks must set command to a non-empty value",
					))
					// we don't need to repeat this error and we don't
					// have any others for this field
					break
				}
			}
		}
	}
	if len(errs) > 0 {
		return util.NewErrors(errs)
	}

	return nil
}
