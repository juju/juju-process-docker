// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

import (
	"fmt"

	"gopkg.in/yaml.v1"
)

// TODO(ericsnow) Use something like "github.com/hashicorp/go-version"
// instead of VersionInfo?

// TODO(ericsnow) Alias "github.com/docker/docker/api/types".Version instead?
//
// Note that the output from "docker version" is different between some
// versions. The output for the different versions is grouped as
// follows:
//   1.0-1.1
//   1.2-1.5
//   1.6-1.7
//   1.8
//
// See github.com/docker/docker:
//  (1.7-1.8) api/client/version.go
//  (1.0-1.6) api/client/commands.go

// Version describes docker's version.
type Version struct {
	// Client is the docker client's version.
	Client VersionInfo
	// APIClient is the version of the docker API client used by
	// the local docker.
	APIClient VersionInfo
}

// ParseVersionCLI converts the CLI output of "docker version" to
// the corresponding Version.
func ParseVersionCLI(out []byte) (Version, error) {
	var version Version

	vers, apis, err := findVersionStrings(out)
	if err != nil {
		return version, err
	}
	vi, err := ParseVersionInfo(vers)
	if err != nil {
		return version, err
	}
	api, err := ParseVersionInfo(apis)
	if err != nil {
		return version, err
	}

	version.Client = *vi
	version.APIClient = *api
	return version, nil
}

func findVersionStrings(out []byte) (string, string, error) {
	// Try pre-1.8 first.
	oldData := make(map[string]string)
	if err := yaml.Unmarshal(out, &oldData); err != nil {
		// XXX retry...
		return "", "", err
	}
	if vers, ok := oldData["Client version"]; ok {
		return vers, oldData["Client API version"], nil
	}

	// Fall back to 1.8's format.
	newData := make(map[string]map[string]string)
	if err := yaml.Unmarshal(out, &newData); err != nil {
		// XXX retry...
		return "", "", err
	}
	if clientInfo, ok := newData["Client"]; ok {
		if vers, ok := clientInfo["Version"]; ok {
			return vers, clientInfo["API version"], nil
		}
	}

	return "", "", fmt.Errorf("could not determine version")
}
