// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker_test

import (
	"fmt"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju-process-docker/docker"
)

var _ = gc.Suite(&dockerSuite{})

type dockerSuite struct{}

func (dockerSuite) TestRun(c *gc.C) {
	fake := fakeRunCommand{
		calls: []runCommandCall{{
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
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"run",
		"--detach",
		"--name", "spam",
		"-e", "FOO=bar",
		"my-spam",
		"do", "something",
	})
}

func (dockerSuite) TestInspect(c *gc.C) {
	fake := fakeRunCommand{
		calls: []runCommandCall{{
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
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"inspect",
		"sad_perlman",
	})
}

func (dockerSuite) TestStop(c *gc.C) {
	fake := fakeRunCommand{
		calls: []runCommandCall{{}},
	}

	err := docker.Stop("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"stop",
		"sad_perlman",
	})
}

func (dockerSuite) TestRemove(c *gc.C) {
	fake := fakeRunCommand{
		calls: []runCommandCall{{}},
	}

	err := docker.Remove("sad_perlman", fake.exec)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(fake.index, gc.Equals, 1)
	c.Check(fake.calls[0].argsIn, jc.DeepEquals, []string{
		"rm",
		"sad_perlman",
	})
}

type runCommandCall struct {
	out      []byte
	err      string
	exitcode int

	argsIn []string
}

type fakeRunCommand struct {
	calls []runCommandCall
	index int
}

// parseArgs returns the args being passed to docker, so arg[0] would be the
// docker command, like "run" or "stop".  This function will exit out of the
// helper exec if it was not passed at least 3 arguments (e.g. docker stop id),
// or if the first arg is not "docker".
func (fakeRunCommand) parseArgs(args []string) (string, []string, error) {
	if len(args) < 2 {
		return "", nil, fmt.Errorf("Not enough arguments passed to docker: %#v\n", args)
	}
	return args[0], args[1:], nil
}

func (frc *fakeRunCommand) exec(args []string) (_ []byte, rErr error) {
	frc.calls[frc.index].argsIn = args
	call := frc.calls[frc.index]
	frc.index += 1

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

	_, args, err := frc.parseArgs(args)
	if err != nil {
		exitcode = 2
		return nil, err
	}

	if call.err != "" {
		return nil, fmt.Errorf(call.err)
	}
	return call.out, rErr
}
