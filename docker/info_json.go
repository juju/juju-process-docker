// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

import (
	"encoding/json"
	"fmt"
)

// ParseInfoJSON converts the JSON output of docker inspect into an Info.
func ParseInfoJSON(id string, data []byte) (*Info, error) {
	var infos []info
	if err := json.Unmarshal(data, &infos); err != nil {
		return nil, fmt.Errorf("can't decode response from docker inspect %s: %s", id, err)
	}
	if len(infos) == 0 {
		return nil, fmt.Errorf("no status returned from docker inspect %s", id)
	}
	if len(infos) > 1 {
		return nil, fmt.Errorf("multiple status values returned from docker inspect %s", id)
	}
	rawInfo := infos[0]

	info := rawInfo.expose()
	return &info, nil

}

// info holds the data deserialized from the output of docker inspect.
type info struct {
	Id    string `json:"Id"`
	Name  string `json:"Name"`
	State state  `json:"State"`
}

// expose converts the deserialized docker inspect output into an Info.
func (i info) expose() Info {
	info := Info{
		ID:   i.Id,
		Name: i.Name,
	}
	info.Process = Process{
		State:    i.State.value(),
		PID:      i.State.Pid,
		ExitCode: i.State.ExitCode,
		Error:    i.State.Error,
	}
	return info
}

type state struct {
	Running    bool   `json:"Running"`
	Paused     bool   `json:"Paused"`
	Restarting bool   `json:"Restarting"`
	OOMKilled  bool   `json:"OOMKilled"`
	Dead       bool   `json:"Dead"`
	Pid        int    `json:"Pid"`
	ExitCode   int    `json:"ExitCode"`
	Error      string `json:"Error"`
}

func (st state) value() State {
	switch {
	case st.Running:
		return StateRunning
	case st.OOMKilled:
		return StateOOMKilled
	case st.Dead:
		return StateDead
	case st.Restarting:
		return StateRestarting
	case st.Paused:
		return StatePaused
	}
	return StateUnknown
}
