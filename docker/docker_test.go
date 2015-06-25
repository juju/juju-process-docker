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
				Internal: process.PortRange{80, 90},
				Protocol: "tcp",
			},
			process.Port{
				External: process.PortRange{From: 8022},
				Internal: process.PortRange{From: 22},
				Protocol: "tcp",
			},
		},
		Volumes: []process.Volume{
			process.Volume{
				ExternalMount: "/foo",
				InternalMount: "/bar",
				Mode:          "ro",
			},
			process.Volume{
				ExternalMount: "/baz",
				InternalMount: "/bat",
				Mode:          "rw",
			},
		},
		EnvVars: map[string]string{"foo": "bar", "baz": "bat"},
	}
	args, err := launchArgs(p)
	c.Assert(err, jc.ErrorIsNil)
	expected := []string{
		"run",
		"--detach",
		"--name", p.Name,
		"-e", "foo=bar",
		"-e", "baz=bat",
		"-p", "8080-8090:80-90/tcp",
		"-p", "8022:22/tcp",
		"-v", "/foo:/bar:ro",
		"-v", "/baz:/bat:rw",
		p.Image,
	}
	expected = append(expected, p.Command...)
	c.Check(args, gc.DeepEquals, expected)
}
