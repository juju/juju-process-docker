// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju-process-docker/docker"
)

var _ = gc.Suite(&versionSuite{})

type versionSuite struct{}

func (versionSuite) TestParseVersionCLIFull(c *gc.C) {
	versions := map[string]string{
		"1.0.1": versionOutput_1_0,
		"1.2.1": versionOutput_1_2,
		"1.6.1": versionOutput_1_6,
		"1.8.1": versionOutput_1_8,
	}
	for vers, out := range versions {
		c.Logf("checking %q", vers)
		vi, err := docker.ParseVersionCLI([]byte(out))
		if !c.Check(err, jc.ErrorIsNil) {
			continue
		}

		c.Check(vi.String(), jc.DeepEquals, vers)
	}
}

func (versionSuite) TestParseVersionCLINoServer(c *gc.C) {
	versions := map[string]string{
		"1.0.1": versionOutput_1_0_noServer,
		"1.2.1": versionOutput_1_2_noServer,
		"1.6.1": versionOutput_1_6_noServer,
		"1.8.1": versionOutput_1_8_noServer,
	}
	for vers, out := range versions {
		c.Logf("checking %q", vers)
		vi, err := docker.ParseVersionCLI([]byte(out))
		if !c.Check(err, jc.ErrorIsNil) {
			continue
		}

		c.Check(vi.String(), jc.DeepEquals, vers)
	}
}

const (
	versionOutput_1_0 = `
Client version: 1.0.1
Client API version: 1.12
Go version (client): go1.2.1
Git commit (client): 990021a
Server version: 1.0.1
Server API version: 1.12
Go version (server): go1.2.1
Git commit (server): 990021a
`
	versionOutput_1_0_noServer = `
Client version: 1.0.1
Client API version: 1.12
Go version (client): go1.2.1
Git commit (client): 990021a
`
	versionOutput_1_2 = `
Client version: 1.2.1
Client API version: 1.14
Go version (client): go1.2.1
Git commit (client): 990021a
OS/Arch (client): linux/amd64
Server version: 1.0.1
Server API version: 1.12
Go version (server): go1.2.1
Git commit (server): 990021a
`
	versionOutput_1_2_noServer = `
Client version: 1.2.1
Client API version: 1.14
Go version (client): go1.2.1
Git commit (client): 990021a
OS/Arch (client): linux/amd64
`
	versionOutput_1_6 = `
Client version: 1.6.1
Client API version: 1.18
Go version (client): go1.2.1
Git commit (client): 990021a
OS/Arch (client): linux/amd64
Server version: 1.0.1
Server API version: 1.12
Go version (server): go1.2.1
Git commit (server): 990021a
OS/Arch (server): linux/amd64
`
	versionOutput_1_6_noServer = `
Client version: 1.6.1
Client API version: 1.18
Go version (client): go1.2.1
Git commit (client): 990021a
OS/Arch (client): linux/amd64
`
	versionOutput_1_8 = `
Client:
 Version:      1.8.1
 API version:  1.20
 Go version:   go1.2.1
 Git commit:   990021a
 Built:        Tue Jun 23 17:56:00 UTC 2015
 OS/Arch:      linux/amd64
Server:
 Version:      1.8.1
 API version:  1.20
 Go version:   go1.2.1
 Git commit:   990021a
 Built:        Tue Jun 23 17:56:00 UTC 2015
 OS/Arch:      linux/amd64
`
	versionOutput_1_8_noServer = `
Client:
 Version:      1.8.1
 API version:  1.20
 Go version:   go1.2.1
 Git commit:   990021a
 Built:        Tue Jun 23 17:56:00 UTC 2015
 OS/Arch:      linux/amd64
`
)
