// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// Command juju-process-docker is a plugin for Juju which enables Juju's
// workload process management to manage Docker containers.
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/juju/juju-process-docker/docker"
	"gopkg.in/juju/charm.v5"
)

func init() {
	log.SetFlags(0)
}

var stdout = log.New(os.Stdout, "", 0)

const (
	mainUsage = `juju-process-docker is a plugin for Juju which enables Juju's workload process
management to manage Docker containers.

Usage:
  juju-process-docker <command> [args]

Commands:
  help [command]	show this help, or help for a command
  launch <process>	launch a docker container with the given parameters
  status <id> 		return status for the process with the given id
  destroy <id>		stop and clean up the process with the given id

For more details about a command, use help <command>.
`

	// TODO(natefinch): fix this when Command becomes a []string.
	// TODO(natefinch): fix this when we start using portranges.
	launchUsage = `launch starts a docker container with the given parameters.

Usage:
  juju-process-docker launch <process>

process is expected to be a json object with the following format:

{
	"Name": "unique-container-name",
	"Command": "command to run",
	"Image": "docker/whalesay",
	"Ports": [
		{
			"External": 7888,
			"Internal": 37888,
			"Endpoint": ""
		}
	],
	"Volumes": [
		{
			"ExternalMount": "/foo/bar",
			"InternalMount": "/baz/bat",
			"Mode": "ro",
			"Name": "foobar"
		}
	],
	"EnvVars": {
		"foo": "bar"
	}
	"TypeOptions": {
		"RepoFile": "/home/foo/repository-to-load.tar"
	}
}

If RepoFile is non-empty, the given repository tar file will be loaded before
attempting to launch the docker image.
`
	statusUsage = `status returns information about the docker container with the given id.

Usage:
  juju-process-docker status <id>

Id is expected to be the id or name of a docker container running on this machine.
`
	destroyUsage = `destroy stops and cleans up after the docker container with the given id.

Usage:
  juju-process-docker destroy <id>

Id is expected to be the id or name of a docker container running on this machine.
`
)

type command struct {
	name  string
	usage string
	fn    func(string) int
}

var cmds = map[string]*command{
	"launch": &command{
		usage: launchUsage,
		fn:    launch,
	},
	"destroy": &command{
		usage: destroyUsage,
		fn:    destroy,
	},
	"status": &command{
		usage: statusUsage,
		fn:    status,
	},
}

func main() {
	os.Exit(Main(os.Args[1:]))
}

func Main(args []string) int {
	if code, exit := maybeHelp(args); exit {
		return code
	}

	return run(args)
}

func maybeHelp(args []string) (code int, exit bool) {
	if len(args) == 0 {
		log.Print(mainUsage)
		return 0, true
	}

	if args[0] == "help" {
		if len(args) == 1 {
			log.Print(mainUsage)
			return 0, true
		}
		for name, cmd := range cmds {
			if name == args[1] {
				log.Print(cmd.usage)
				return 0, true
			}
		}
		stdout.Printf("unknown command %q", args[1])
		log.Println()
		log.Print(mainUsage)
		return 1, true
	}
	return 0, false
}

func run(args []string) int {
	for name, cmd := range cmds {
		if args[0] == name {
			if len(args) != 2 {
				stdout.Printf("wrong number of arguments for cmd %q", name)
				log.Println()
				log.Print(cmd.usage)
				return 1
			}
			// valid command.
			return cmd.fn(args[1])
		}
	}
	stdout.Printf("unknown command %q", args[0])
	log.Println()
	log.Print(mainUsage)
	return 1
}

var (
	dockerLaunch  = docker.Launch
	dockerStatus  = docker.Status
	dockerDestroy = docker.Destroy
)

func launch(proc string) int {
	p := charm.Process{}
	if err := json.Unmarshal([]byte(proc), &p); err != nil {
		stdout.Printf("can't decode proc-info: %s", err)
		return 1
	}
	details, err := dockerLaunch(p)
	if err != nil {
		stdout.Print(err)
		return 1
	}

	b, err := json.Marshal(details)
	if err != nil {
		stdout.Print(err)
		return 1
	}
	stdout.Print(string(b))
	return 0
}

func status(id string) int {
	status, err := dockerStatus(id)
	if err != nil {
		stdout.Print(err)
		return 1
	}
	b, err := json.Marshal(status)
	if err != nil {
		stdout.Print(err)
		return 1
	}
	stdout.Print(string(b))
	return 0
}

func destroy(id string) int {
	err := dockerDestroy(id)
	if err != nil {
		stdout.Print(err)
		return 1
	}
	return 0
}
