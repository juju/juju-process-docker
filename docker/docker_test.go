// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v5"
)

type suite struct{}

var _ = gc.Suite(&suite{})

func (suite) TestLaunchArgs(c *gc.C) {
	args, err := launchArgs(fakeProc)
	c.Assert(err, jc.ErrorIsNil)
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

func (suite) TestStatusFromInspectNone(c *gc.C) {
	b := []byte("not json")
	_, err := statusFromInspect("id", b)
	c.Assert(err, gc.ErrorMatches, "can't decode response from docker inspect id.*")
}

func (suite) TestStatusFromInspectEmpty(c *gc.C) {
	b := []byte(`[]`)
	_, err := statusFromInspect("id", b)
	c.Assert(err, gc.ErrorMatches, "no status returned from docker inspect id")
}

func (suite) TestStatusFromInspectMultiple(c *gc.C) {
	b := []byte(`[{"Name":"foo"},{"Name":"bar"}]`)
	_, err := statusFromInspect("id", b)
	c.Assert(err, gc.ErrorMatches, "multiple status values returned from docker inspect id")
}

func (suite) TestStatusFromInspect(c *gc.C) {
	s, err := statusFromInspect("id", []byte(fakeInspectOutput))
	c.Assert(err, jc.ErrorIsNil)
	expected := status{
		State: state{
			Running: true,
			Pid:     11820,
		},
		Name: "/sad_perlman",
	}
	c.Assert(s, gc.Equals, expected)
}

func (suite) TestBrief(c *gc.C) {
	type test struct {
		in  status
		out string
	}
	tests := []test{
		{
			in:  status{State: state{Running: true}},
			out: "Running",
		},
		{
			in:  status{State: state{OOMKilled: true}},
			out: "OOMKilled",
		},
		{
			in:  status{State: state{Dead: true}},
			out: "Dead",
		},
		{
			in:  status{State: state{Restarting: true}},
			out: "Restarting",
		},
		{
			in:  status{State: state{Paused: true}},
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

func (suite) TestLaunch(c *gc.C) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	pd, err := Launch(fakeProc)
	c.Assert(err, jc.ErrorIsNil)
	expected := ProcDetails{
		ID: "sad_perlman",
		Status: ProcStatus{
			State: "Running",
		},
	}
	c.Assert(pd, gc.Equals, expected)
}

func (suite) TestStatus(c *gc.C) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	ps, err := Status("someid")
	c.Assert(err, jc.ErrorIsNil)
	expected := ProcStatus{"Running"}
	c.Assert(ps, gc.Equals, expected)
}

func (suite) TestDestroy(c *gc.C) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	err := Destroy("someid")
	c.Assert(err, jc.ErrorIsNil)
}

// fakeExecCommand replaces the normal exec.Command call to produce executables.
// It returns a command that calls this test executable, telling it to run our
// TestExecHelper test.  The original command and arguments are passed as
// arguments to the testhelper after a "--" argument.
func fakeExecCommand(name string, args ...string) *exec.Cmd {
	args = append([]string{"-test.run=TestExecHelper", "--", name}, args...)
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func fakeExecCommandError(name string, args ...string) *exec.Cmd {
	cmd := fakeExecCommand(name, args...)
	cmd.Env = append(cmd.Env, "GO_HELPER_PROCESS_ERROR=1")
	return cmd
}

// TestExecHelper is a fake test that is just used to do predictable things when
// we run commands.
func TestExecHelper(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := getTestArgs()
	shouldErr := os.Getenv("GO_HELPER_PROCESS_ERROR") == "1"
	if shouldErr {
		defer os.Exit(1)
	} else {
		defer os.Exit(0)
	}
	switch args[0] {
	case "run":
		doRun(shouldErr)
	case "inspect":
		doInspect(shouldErr)
	case "stop":
		doStop(shouldErr)
	case "rm":
		doRm(shouldErr)
	}
}

func doStop(shouldErr bool) {
	if shouldErr {
		fmt.Fprintln(os.Stderr, "Error response from daemon: no such id: foo")
		return
	}
	// when docker successfully stops a container, it prints out the container id.
	fmt.Fprintln(os.Stdout, "somebigid")
}

func doRm(shouldErr bool) {
	if shouldErr {
		fmt.Fprintln(os.Stderr, "Error response from daemon: no such id: foo")
		return
	}
	// when docker successfully removes a container, it prints out the container id.
	fmt.Fprintln(os.Stdout, "somebigid")
}

func doRun(shouldErr bool) {
	if shouldErr {
		fmt.Fprintf(os.Stderr, `Unable to find image 'foo:latest' locally
Pulling repository foo
Error: image library/foo:latest not found
`)
		return
	}
	fmt.Fprintln(os.Stdout, "somebigid")
}

func doInspect(shouldErr bool) {
	if shouldErr {
		fmt.Fprintln(os.Stderr, "Error: No such image or container: wegwgwg")
		fmt.Fprintln(os.Stdout, "[]")
		return
	}
	fmt.Fprint(os.Stdout, fakeInspectOutput)
}

// getTestArgs returns the args being passed to docker, so arg[0] would be the
// docker command, like "run" or "stop".  This function will exit out of the
// helper exec if it was not passed at least 3 arguments (e.g. docker stop id),
// or if the first arg is not "docker".
func getTestArgs() []string {
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "Not enough arguments passed to docker: %#v\n", args)
		os.Exit(2)
	}
	if args[0] != "docker" {
		fmt.Fprintf(os.Stderr, "Calling %q instead of docker\n", args[0])
		os.Exit(2)
	}

	return args[1:]
}
