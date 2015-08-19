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
func Run(args RunArgs, exec func([]string) ([]byte, error)) (string, error) {
	if exec == nil {
		exec = runCommand
	}

	cmdArgs := args.CommandlineArgs()
	out, err := exec(cmdArgs)
	if err != nil {
		return "", err
	}
	id := string(bytes.TrimSpace(out))
	return id, nil
}

// Inspect gets info about the given container ID (or name).
//
// If exec is nil then the default (via exec.Command) is used.
func Inspect(id string, exec func([]string) ([]byte, error)) (*Info, error) {
	if exec == nil {
		exec = runCommand
	}

	args := []string{"inspect", id}
	out, err := exec(args)
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
func Stop(id string, exec func([]string) ([]byte, error)) error {
	if exec == nil {
		exec = runCommand
	}

	args := []string{"stop", id}
	if _, err := exec(args); err != nil {
		return err
	}
	return nil
}

// Remove removes the identified container.
//
// If exec is nil then the default (via exec.Command) is used.
func Remove(id string, exec func([]string) ([]byte, error)) error {
	if exec == nil {
		exec = runCommand
	}

	args := []string{"rm", id}
	if _, err := exec(args); err != nil {
		return err
	}
	return nil
}
