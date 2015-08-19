// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Package docker exposes an API to convert Jujuisms to dockerisms.
package docker

import (
	"bytes"
)

// Run runs a new docker container with the given info.
//
// If exec is nil then the default (via exec.Command) is used.
func Run(args RunArgs, exec func(string, ...string) ([]byte, error)) (string, error) {
	if exec == nil {
		exec = runDocker
	}

	cmdArgs := args.CommandlineArgs()
	out, err := exec("run", cmdArgs...)
	if err != nil {
		return "", err
	}
	id := string(bytes.TrimSpace(out))
	return id, nil
}

// Inspect gets info about the given container ID (or name).
//
// If exec is nil then the default (via exec.Command) is used.
func Inspect(id string, exec func(string, ...string) ([]byte, error)) (*Info, error) {
	if exec == nil {
		exec = runDocker
	}

	out, err := exec("inspect", id)
	if err != nil {
		return nil, err
	}

	info, err := ParseInfoJSON(id, out)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// Stop stops the identified container.
//
// If exec is nil then the default (via exec.Command) is used.
func Stop(id string, exec func(string, ...string) ([]byte, error)) error {
	if exec == nil {
		exec = runDocker
	}

	if _, err := exec("stop", id); err != nil {
		return err
	}
	return nil
}

// Remove removes the identified container.
//
// If exec is nil then the default (via exec.Command) is used.
func Remove(id string, exec func(string, ...string) ([]byte, error)) error {
	if exec == nil {
		exec = runDocker
	}

	if _, err := exec("rm", id); err != nil {
		return err
	}
	return nil
}
