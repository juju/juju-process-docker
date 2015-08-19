// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v5"
)

var _ = gc.Suite(&pluginSuite{})

type pluginSuite struct{}

func (pluginSuite) TestLaunchArgs(c *gc.C) {
	runArgs := launchArgs(fakeProc)
	args := runArgs.CommandlineArgs()

	expected := []string{
		"run",
		"--detach",
		"--name", "juju-name",
		"-e", "foo=bar",
		"-e", "baz=bat",
		"-p", "8080:80/tcp",
		"-p", "8022:22/tcp",
		"-v", "/foo:/bar:ro",
		"-v", "/baz:/bat:rw",
		"docker/whalesay",
		"cowsay", "boo!",
	}
	c.Check(args, gc.DeepEquals, expected)
}

func (pluginSuite) TestBrief(c *gc.C) {
	type test struct {
		in  status
		out string
	}
	tests := []test{
		{
			in:  status{State: Process{State: StateRunning}},
			out: "Running",
		},
		{
			in:  status{State: Process{State: StateOOMKilled}},
			out: "OOMKilled",
		},
		{
			in:  status{State: Process{State: StateDead}},
			out: "Dead",
		},
		{
			in:  status{State: Process{State: StateRestarting}},
			out: "Restarting",
		},
		{
			in:  status{State: Process{State: StatePaused}},
			out: "Paused",
		},
		{
			in:  status{},
			out: "Unknown",
		},
	}
	for _, t := range tests {
		c.Check(t.out, gc.Equals, t.in.brief())
	}
}

var fakeProc = charm.Process{
	Name:        "juju-name",
	Description: "desc",
	Type:        "docker",
	TypeOptions: map[string]string{"foo": "bar"},
	// TODO(natefinch): update this when Command becomes a slice
	Command: "cowsay boo!",
	Image:   "docker/whalesay",
	// TODO(natefinch): update this when we use portranges
	Ports: []charm.ProcessPort{
		charm.ProcessPort{
			External: 8080,
			Internal: 80,
		},
		charm.ProcessPort{
			External: 8022,
			Internal: 22,
		},
	},
	Volumes: []charm.ProcessVolume{
		charm.ProcessVolume{
			ExternalMount: "/foo",
			InternalMount: "/bar",
			Mode:          "ro",
		},
		charm.ProcessVolume{
			ExternalMount: "/baz",
			InternalMount: "/bat",
			Mode:          "rw",
		},
	},
	EnvVars: map[string]string{"foo": "bar", "baz": "bat"},
}

func (pluginSuite) TestLaunch(c *gc.C) {
	calls := []call{{
		out: []byte("/sad_perlman"),
	}, {
		out: []byte(fakeInspectOutput),
	}}
	defaultRunCommand = fakeRunCommand(calls...)
	defer func() { defaultRunCommand = runCommand }()

	pd, err := Launch(fakeProc)
	c.Assert(err, jc.ErrorIsNil)
	expected := ProcDetails{
		ID: "sad_perlman",
		Status: ProcStatus{
			State: "Running",
		},
	}
	c.Assert(pd, jc.DeepEquals, expected)
}

func (pluginSuite) TestStatus(c *gc.C) {
	calls := []call{{
		out: []byte(fakeInspectOutput),
	}}
	defaultRunCommand = fakeRunCommand(calls...)
	defer func() { defaultRunCommand = runCommand }()

	ps, err := Status("someid")
	c.Assert(err, jc.ErrorIsNil)
	expected := ProcStatus{"Running"}
	c.Assert(ps, jc.DeepEquals, expected)
}

func (pluginSuite) TestDestroy(c *gc.C) {
	calls := make([]call, 2, 2)
	defaultRunCommand = fakeRunCommand(calls...)
	defer func() { defaultRunCommand = runCommand }()

	err := Destroy("someid")
	c.Assert(err, jc.ErrorIsNil)
}

type call struct {
	out []byte
	err error

	argsIn []string
}

func fakeRunCommand(calls ...call) func([]string) ([]byte, error) {
	index := 0
	return func(args []string) ([]byte, error) {
		calls[index].argsIn = args
		call := calls[index]
		index += 1

		if call.err != nil {
			return nil, call.err
		}
		return call.out, nil
	}
}

const fakeInspectOutput = `
[
{
    "Id": "b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232",
    "Created": "2015-06-25T11:05:53.694518797Z",
    "Path": "sleep",
    "Args": [
        "30"
    ],
    "State": {
        "Running": true,
        "Paused": false,
        "Restarting": false,
        "OOMKilled": false,
        "Dead": false,
        "Pid": 11820,
        "ExitCode": 0,
        "Error": "",
        "StartedAt": "2015-06-25T11:05:53.8401024Z",
        "FinishedAt": "0001-01-01T00:00:00Z"
    },
    "Image": "fb434121fc77c965f255cbb848927f577bbdbd9325bdc1d7f1b33f99936b9abb",
    "NetworkSettings": {
        "Bridge": "",
        "EndpointID": "9915c7299be4f77c18f3999ef422b79996ea8c5796e2befd1442d67e5cefb50d",
        "Gateway": "172.17.42.1",
        "GlobalIPv6Address": "",
        "GlobalIPv6PrefixLen": 0,
        "HairpinMode": false,
        "IPAddress": "172.17.0.2",
        "IPPrefixLen": 16,
        "IPv6Gateway": "",
        "LinkLocalIPv6Address": "",
        "LinkLocalIPv6PrefixLen": 0,
        "MacAddress": "02:42:ac:11:00:02",
        "NetworkID": "3346546be8f76006e44000b007da48e576e788ba1d3e3cd275545837d4d7c80a",
        "PortMapping": null,
        "Ports": {},
        "SandboxKey": "/var/run/docker/netns/b508c7d5c272",
        "SecondaryIPAddresses": null,
        "SecondaryIPv6Addresses": null
    },
    "ResolvConfPath": "/var/lib/docker/containers/b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232/resolv.conf",
    "HostnamePath": "/var/lib/docker/containers/b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232/hostname",
    "HostsPath": "/var/lib/docker/containers/b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232/hosts",
    "LogPath": "/var/lib/docker/containers/b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232/b508c7d5c2722b7ac4f105fedf835789fb705f71feb6e264f542dc33cdc41232-json.log",
    "Name": "/sad_perlman",
    "RestartCount": 0,
    "Driver": "aufs",
    "ExecDriver": "native-0.2",
    "MountLabel": "",
    "ProcessLabel": "",
    "Volumes": {},
    "VolumesRW": {},
    "AppArmorProfile": "",
    "ExecIDs": null,
    "HostConfig": {
        "Binds": null,
        "ContainerIDFile": "",
        "LxcConf": [],
        "Memory": 0,
        "MemorySwap": 0,
        "CpuShares": 0,
        "CpuPeriod": 0,
        "CpusetCpus": "",
        "CpusetMems": "",
        "CpuQuota": 0,
        "BlkioWeight": 0,
        "OomKillDisable": false,
        "Privileged": false,
        "PortBindings": {},
        "Links": null,
        "PublishAllPorts": false,
        "Dns": null,
        "DnsSearch": null,
        "ExtraHosts": null,
        "VolumesFrom": null,
        "Devices": [],
        "NetworkMode": "bridge",
        "IpcMode": "",
        "PidMode": "",
        "UTSMode": "",
        "CapAdd": null,
        "CapDrop": null,
        "RestartPolicy": {
            "Name": "no",
            "MaximumRetryCount": 0
        },
        "SecurityOpt": null,
        "ReadonlyRootfs": false,
        "Ulimits": null,
        "LogConfig": {
            "Type": "json-file",
            "Config": {}
        },
        "CgroupParent": ""
    },
    "Config": {
        "Hostname": "b508c7d5c272",
        "Domainname": "",
        "User": "",
        "AttachStdin": false,
        "AttachStdout": false,
        "AttachStderr": false,
        "PortSpecs": null,
        "ExposedPorts": null,
        "Tty": false,
        "OpenStdin": false,
        "StdinOnce": false,
        "Env": [
            "PATH=/usr/local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
        ],
        "Cmd": [
            "sleep",
            "30"
        ],
        "Image": "docker/whalesay",
        "Volumes": null,
        "VolumeDriver": "",
        "WorkingDir": "/cowsay",
        "Entrypoint": null,
        "NetworkDisabled": false,
        "MacAddress": "",
        "OnBuild": null,
        "Labels": {}
    }
}
]
`
