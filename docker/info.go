// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

import (
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
)

var v120 = VersionInfo{Major: 1, Minor: 20}

// ParseInfoJSON converts the JSON output of docker inspect into an Info.
func ParseInfoJSON(id string, data []byte, version Version) (*Info, error) {
	if version.APIClient.Compare(v120) < 0 {
		return parseInfoPost120(id, data)
	}
	return parseInfoPre120(id, data)
}

func parseInfoPre120(id string, data []byte) (*Info, error) {
	var infos []Info
	if err := json.Unmarshal(data, &infos); err != nil {
		return nil, fmt.Errorf("can't decode response from docker inspect %s: %s", id, err)
	}
	if len(infos) == 0 {
		return nil, fmt.Errorf("no status returned from docker inspect %s", id)
	}
	if len(infos) > 1 {
		return nil, fmt.Errorf("multiple status values returned from docker inspect %s", id)
	}
	return &infos[0], nil
}

func parseInfoPost120(id string, data []byte) (*Info, error) {
	// TODO(ericsnow) finish!
	return nil, fmt.Errorf("version 1.8+ not supported")
}

// Info holds all available information about a docker container.
type Info types.ContainerJSONPre120

// These are the different possible states of a container.
const (
	StateUnknown    = ""
	StateRunning    = "Running"
	StatePaused     = "Paused"
	StateRestarting = "Restarting"
	StateOOMKilled  = "OOMKilled"
	StateDead       = "Dead"
)

// StateValue returns the label for the current state of the container.
func (info Info) StateValue() string {
	switch {
	case info.State.Running:
		return StateRunning
	case info.State.OOMKilled:
		return StateOOMKilled
	case info.State.Dead:
		return StateDead
	case info.State.Restarting:
		return StateRestarting
	case info.State.Paused:
		return StatePaused
	}
	return StateUnknown
}
