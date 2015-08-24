// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

// Info holds all available information about a docker container.
type Info struct {
	// ID is the unique identifier docker uses to identify the container.
	ID string
	// Name is the unique name that may be used to identify the container.
	Name string
	// Process describes the process running the container.
	Process Process
}

// Process holds the information about the process running the container,
// as provided by the "docker inspect" command.
type Process struct {
	// State is the state of the container process.
	State State
	// PID is the PID of the container process.
	PID int
	// ExitCode is the exit code of a stopped container.
	ExitCode int
	// Error is the error message from a failed container.
	Error string
}

// These are the different possible values of State.Current.
const (
	StateUnknown    State = ""
	StateRunning    State = "Running"
	StatePaused     State = "Paused"
	StateRestarting State = "Restarting"
	StateOOMKilled  State = "OOMKilled"
	StateDead       State = "Dead"
)

// State describes the high-level state of a docker container.
type State string

// String returns the correct representation of the container state.
func (st State) String() string {
	switch st {
	case StateRunning:
	case StatePaused:
	case StateRestarting:
	case StateOOMKilled:
	case StateDead:
	default:
		return "Unknown"
	}
	return string(st)
}
