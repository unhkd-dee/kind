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

// Package cluster implements the `create cluster` command
package cluster

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/config/encoding"
	"sigs.k8s.io/kind/pkg/util"
)

type flagpole struct {
	Name      string
	Config    string
	ImageName string
	Retain    bool
	Wait      time.Duration
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	flags := &flagpole{}
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Creates a local Kubernetes cluster",
		Long:  "Creates a local Kubernetes cluster using Docker container 'nodes'",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}
	cmd.Flags().StringVar(&flags.Name, "name", "1", "cluster context name")
	cmd.Flags().StringVar(&flags.Config, "config", "", "path to a kind config file")
	cmd.Flags().StringVar(&flags.ImageName, "image", "", "node docker image to use for booting the cluster")
	cmd.Flags().BoolVar(&flags.Retain, "retain", false, "retain nodes for debugging when cluster creation fails")
	cmd.Flags().DurationVar(&flags.Wait, "wait", time.Duration(0), "Wait for control plane node to be ready (default 0s)")
	return cmd
}

func runE(flags *flagpole, cmd *cobra.Command, args []string) error {
	// load the config
	cfg, err := encoding.Load(flags.Config)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// validate the config
	err = cfg.Validate()
	if err != nil {
		log.Error("Invalid configuration!")
		configErrors := err.(*util.Errors)
		for _, problem := range configErrors.Errors() {
			log.Error(problem)
		}
		return fmt.Errorf("aborting due to invalid configuration")
	}

	// TODO(fabrizio pandini): this check is temporary / WIP
	// kind v1alpha2 config fully supports multi nodes, but the cluster creation logic implemented in
	// pkg/cluster/contex.go does not (yet).
	// As soon a multi node support is implemented in pkg/cluster/contex.go, this should go away
	if len(cfg.AllReplicas()) > 1 {
		return fmt.Errorf("multi node support is still a work in progress, currently only single node cluster are supported")
	}

	// create a cluster context and create the cluster
	ctx := cluster.NewContext(flags.Name)
	if flags.ImageName != "" {
		cfg.BootStrapControlPlane().Image = flags.ImageName
		err := cfg.Validate()
		if err != nil {
			log.Errorf("Invalid flags, configuration failed validation: %v", err)
			return fmt.Errorf("aborting due to invalid configuration")
		}
	}
	if err = ctx.Create(cfg, flags.Retain, flags.Wait); err != nil {
		return fmt.Errorf("failed to create cluster: %v", err)
	}

	return nil
}
