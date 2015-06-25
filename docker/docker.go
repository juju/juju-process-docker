// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package docker exposes an API to convert Jujuisms to dockerisms.
package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/juju/deputy"
	"gopkg.in/juju/charm.v5/process"
)

// Launch runs a new docker container with the given process data.
func Launch(p process.Process) (ProcDetails, error) {
	args, err := launchArgs(p)
	if err != nil {
		return ProcDetails{}, err
	}
	d := deputy.Deputy{
		Errors: deputy.FromStderr,
	}
	cmd := exec.Command("docker", args...)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	if err := d.Run(cmd); err != nil {
		return ProcDetails{}, err
	}
	id := string(bytes.TrimSpace(out.Bytes()))
	status, err := inspect(id)
	if err != nil {
		return ProcDetails{}, fmt.Errorf("can't get status for container %q: %s", id, err)
	}
	return ProcDetails{ID: status.Name, ProcStatus: ProcStatus{Status: status.brief()}}, nil
}

// Status returns the ProcStatus for the docker container with the given id.
func Status(id string) (ProcStatus, error) {
	status, err := inspect(id)
	if err != nil {
		return ProcStatus{}, err
	}
	return ProcStatus{Status: status.brief()}, nil
}

// Destroy stops and removes the docker container with the given id.
func Destroy(id string) error {
	d := deputy.Deputy{
		Errors: deputy.FromStderr,
	}
	cmd := exec.Command("docker", "stop", id)
	if err := d.Run(cmd); err != nil {
		return fmt.Errorf("error while stopping container %q: %s", err)
	}

	cmd = exec.Command("docker", "rm", id)
	if err := d.Run(cmd); err != nil {
		return fmt.Errorf("error while removing container %q: %s", err)
	}
	return nil
}

// launchArgs converts the Process struct into arguments for the docker run
// command.
func launchArgs(p process.Process) ([]string, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid proc-info: %s", err)
	}

	args := []string{"run", "--detach", "--name", p.Name}
	for k, v := range p.EnvVars {
		args = append(args, "-e", k+"="+v)
	}

	for _, p := range p.Ports {
		if p.External.To > 0 {
			args = append(args, "-p",
				fmt.Sprintf("%d-%d:%d-%d/%s",
					p.External.From,
					p.External.To,
					p.Internal.From,
					p.Internal.To,
					p.Protocol,
				),
			)

		} else {
			args = append(args, "-p", fmt.Sprintf("%d:%d/%s", p.External.From, p.Internal.From, p.Protocol))
		}
	}

	for _, v := range p.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s:%s", v.ExternalMount, v.InternalMount, v.Mode))
	}

	// Image and Command must come after all options.
	args = append(args, p.Image)
	if len(p.Command) > 0 {
		args = append(args, p.Command...)
	}
	return args, nil
}

// status is the struct that contains the schema returned by docker's inspect command
type status struct {
	State struct {
		Running    bool
		Paused     bool
		Restarting bool
		OOMKilled  bool
		Dead       bool
		Pid        int
		ExitCode   int
		Error      string
	}
	Name string
}

// brief returns a short summary for the status.
func (s *status) brief() string {
	switch {
	case s.State.Running:
		return "Running"
	case s.State.OOMKilled:
		return "OOMKilled"
	case s.State.Dead:
		return "Dead"
	case s.State.Restarting:
		return "Restarting"
	case s.State.Paused:
		return "Paused"
	}
	return "Unknown"
}

// inspect calls docker inspect and returns the unmarshaled json response.
func inspect(id string) (status, error) {
	cmd := exec.Command("docker", "inspect", id)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	d := deputy.Deputy{
		Errors: deputy.FromStderr,
	}
	if err := d.Run(cmd); err != nil {
		return status{}, err
	}
	var st []status
	if err := json.Unmarshal(out.Bytes(), &st); err != nil {
		return status{}, fmt.Errorf("can't decode response from docker inspect %s: %s", id, err)
	}
	if len(st) == 0 {
		return status{}, errors.New("no status returned from docker inspect " + id)
	}
	if len(st) > 1 {
		return status{}, errors.New("multiple status values returned from docker inspect " + id)
	}
	return st[0], nil
}

// These two structs are copied from juju/process/plugin

// ProcDetails represents information about a process launched by a plugin.
type ProcDetails struct {
	// ID is a unique string identifying the process to the plugin.
	ID string `json:"id"`
	// ProcStatus is the status of the process after launch.
	ProcStatus
}

// ProcStatus represents the data returned from the Status call.
type ProcStatus struct {
	// Status represents the human-readable string returned by the plugin for
	// the process.
	Status string `json:"status"`
}
