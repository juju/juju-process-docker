// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker_test

import (
	"fmt"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju-process-docker/docker"
)

var _ = gc.Suite(&dockerSuite{})

type dockerSuite struct {
	testing.CleanupSuite
}

func (dockerSuite) TestRunOkay(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{
			out: []byte("eggs"),
		}},
	}

	args := docker.RunArgs{
		Name:    "spam",
		Image:   "my-spam",
		Command: "do something",
		EnvVars: map[string]string{
			"FOO": "bar",
		},
	}
	id, err := docker.Run(args, fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(id, gc.Equals, "eggs")
	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "run")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"--detach",
		"--name", "spam",
		"-e", "FOO=bar",
		"my-spam",
		"do", "something",
	})
}

func (s *dockerSuite) TestRunDefaultExec(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{
			out: []byte("eggs"),
		}},
	}
	s.PatchValue(docker.DefaultExec, fake.exec)

	args := docker.RunArgs{
		Name:    "spam",
		Image:   "my-spam",
		Command: "do something",
		EnvVars: map[string]string{
			"FOO": "bar",
		},
	}
	id, err := docker.Run(args, nil)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(id, gc.Equals, "eggs")
	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "run")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"--detach",
		"--name", "spam",
		"-e", "FOO=bar",
		"my-spam",
		"do", "something",
	})
}

func (dockerSuite) TestRunMinimal(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{
			out: []byte("eggs"),
		}},
	}

	args := docker.RunArgs{
		Image: "my-spam",
	}
	id, err := docker.Run(args, fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(id, gc.Equals, "eggs")
	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "run")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"--detach",
		"my-spam",
	})
}

func (dockerSuite) TestInspectOkay(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{
			out: []byte(fakeInspectOutput),
		}},
	}

	info, err := docker.Inspect("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(info, jc.DeepEquals, &docker.Info{
		ID: "b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232",
		// TODO(ericsnow) Strip the leading slash.
		Name: "/sad_perlman",
		Process: docker.Process{
			State: docker.StateRunning,
			PID:   11820,
		},
	})
	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "inspect")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"sad_perlman",
	})
}

func (s *dockerSuite) TestInspectDefaultExec(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{
			out: []byte(fakeInspectOutput),
		}},
	}
	s.PatchValue(docker.DefaultExec, fake.exec)

	info, err := docker.Inspect("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(info, jc.DeepEquals, &docker.Info{
		ID: "b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232",
		// TODO(ericsnow) Strip the leading slash.
		Name: "/sad_perlman",
		Process: docker.Process{
			State: docker.StateRunning,
			PID:   11820,
		},
	})
	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "inspect")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"sad_perlman",
	})
}

func (dockerSuite) TestStopOkay(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{}},
	}

	err := docker.Stop("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "stop")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"sad_perlman",
	})
}

func (s *dockerSuite) TestStopDefaultExec(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{}},
	}
	s.PatchValue(docker.DefaultExec, fake.exec)

	err := docker.Stop("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "stop")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"sad_perlman",
	})
}

func (dockerSuite) TestRemoveOkay(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{}},
	}

	err := docker.Remove("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "rm")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"sad_perlman",
	})
}

func (s *dockerSuite) TestRemoveDefaultExec(c *gc.C) {
	fake := fakeRunDocker{
		calls: []runDockerCall{{}},
	}
	s.PatchValue(docker.DefaultExec, fake.exec)

	err := docker.Remove("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].commandIn, gc.Equals, "rm")
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"sad_perlman",
	})
}

type runDockerCall struct {
	out      []byte
	err      string
	exitcode int

	commandIn string
	argsIn    []string
}

type fakeRunDocker struct {
	calls []runDockerCall
	index int
}

// checkArgs verifies the args being passed to docker.
func (fakeRunDocker) checkArgs(command string, args []string) error {
	if len(args) < 1 {
		fullArgs := append([]string{command}, args...)
		return fmt.Errorf("Not enough arguments passed to docker: %#v\n", fullArgs)
	}
	return nil
}

func (frd *fakeRunDocker) exec(command string, args ...string) (_ []byte, rErr error) {
	frd.calls[frd.index].commandIn = command
	frd.calls[frd.index].argsIn = args
	call := frd.calls[frd.index]
	frd.index += 1

	exitcode := call.exitcode
	defer func() {
		if rErr == nil && exitcode != 0 {
			rErr = fmt.Errorf("ERROR")
		}
		if rErr != nil {
			if exitcode == 0 {
				exitcode = 1
			}
			rErr = fmt.Errorf("exit status %d: %v", exitcode, rErr)
		}
	}()

	if err := frd.checkArgs(command, args); err != nil {
		exitcode = 2
		return nil, err
	}

	if call.err != "" {
		return nil, fmt.Errorf(call.err)
	}
	return call.out, rErr
}
