// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package docker exposes an API to convert Jujuisms to dockerisms.
package docker

import (
	"fmt"
	"strings"

	"gopkg.in/juju/charm.v5"
)

var defaultRunCommand = runCommand

// Launch runs a new docker container with the given process data.
func Launch(p charm.Process) (ProcDetails, error) {
	var details ProcDetails
	if err := p.Validate(); err != nil {
		return details, fmt.Errorf("invalid proc-info: %s", err)
	}

	args := launchArgs(p)
	id, err := Run(args, defaultRunCommand)
	if err != nil {
		return details, err
	}

	status, err := inspect(id)
	if err != nil {
		return details, fmt.Errorf("can't get status for container %q: %s", id, err)
	}

	details.ID = strings.TrimPrefix(status.Name, "/")
	details.Status = ProcStatus{
		State: status.brief(),
	}
	return details, nil
}

// Status returns the ProcStatus for the docker container with the given id.
func Status(id string) (ProcStatus, error) {
	status, err := inspect(id)
	if err != nil {
		return ProcStatus{}, err
	}
	return ProcStatus{
		State: status.brief(),
	}, nil
}

// Destroy stops and removes the docker container with the given id.
func Destroy(id string) error {
	if err := Stop(id, defaultRunCommand); err != nil {
		return err
	}
	if err := Remove(id, defaultRunCommand); err != nil {
		return err
	}
	return nil
}

// launchArgs converts the Process struct into arguments for the docker run
// command.
func launchArgs(p charm.Process) RunArgs {
	args := RunArgs{
		Name:  p.Name,
		Image: p.Image,
	}

	if p.EnvVars != nil {
		args.EnvVars = make(map[string]string, len(p.EnvVars))
		for name, value := range p.EnvVars {
			args.EnvVars[name] = value
		}
	}

	for _, port := range p.Ports {
		// TODO(natefinch): update this when we use portranges
		args.Ports = append(args.Ports, PortAssignment{
			External: port.External,
			Internal: port.Internal,
			Protocol: "tcp",
		})
	}

	for _, vol := range p.Volumes {
		// TODO(natefinch): update this when we use portranges
		args.Mounts = append(args.Mounts, MountAssignment{
			External: vol.ExternalMount,
			Internal: vol.InternalMount,
			Mode:     vol.Mode,
		})
	}

	// TODO(natefinch): update this when we make command a list of strings
	if p.Command != "" {
		args.Command = p.Command
	}

	return args
}

// status is the struct that contains the schema returned by docker's inspect command
type status struct {
	State Process
	Name  string
}

// brief returns a short summary for the status.
func (s *status) brief() string {
	return s.State.State.String()
}

// inspect calls docker inspect and returns the unmarshaled json response.
func inspect(id string) (status, error) {
	info, err := Inspect(id, defaultRunCommand)
	if err != nil {
		return status{}, err
	}
	st := status{
		State: info.Process,
		Name:  info.Name,
	}
	return st, nil
}

// These two structs are copied from juju/process/plugin

// ProcDetails represents information about a process launched by a plugin.
type ProcDetails struct {
	// ID is a unique string identifying the process to the plugin.
	ID string `json:"id"`
	// Status is the status of the process after launch.
	Status ProcStatus `json:"status"`
}

// ProcStatus represents the data returned from the Status call.
type ProcStatus struct {
	// State represents the human-readable string returned by the plugin for
	// the process.
	State string `json:"state"`
}
