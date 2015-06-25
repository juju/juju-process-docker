// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v5/process"
)

type suite struct{}

var _ = gc.Suite(&suite{})

func (suite) TestLaunchArgs(c *gc.C) {
	p := process.Process{
		Name:        "juju-name",
		Description: "desc",
		Type:        "docker",
		TypeOptions: map[string]string{"foo": "bar"},
		Command:     []string{"cowsay", "boo!"},
		Image:       "docker/whalesay",
		Ports: []process.Port{
			process.Port{
				External: process.PortRange{8080, 8090},
				Internal: process.PortRange{80, 88},
				Protocol: "tcp",
			},
		},
		Volumes: []process.Volume{
			process.Volume{
				ExternalMount: "/foo",
				InternalMount: "/bar",
				Mode:          "ro",
			},
		},
		EnvVars: map[string]string{"foo": "bar"},
	}
	args, err := launchArgs(p)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(args, gc.DeepEquals, []string{"run", "--detach", "--name", p.Name, "-e", "foo=bar"})
}
