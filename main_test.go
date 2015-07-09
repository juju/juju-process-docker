// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"bytes"
	"errors"
	"log"

	"github.com/juju/juju-process-docker/docker"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v5"
)

type suite struct {
	log    *bytes.Buffer
	stdout *bytes.Buffer
}

var _ = gc.Suite(&suite{})

func (s *suite) SetUpSuite(c *gc.C) {
	s.log = &bytes.Buffer{}
	s.stdout = &bytes.Buffer{}
	log.SetOutput(s.log)
	stdout = log.New(s.stdout, "", 0)
}

func (s *suite) SetUpTest(c *gc.C) {
	s.log.Reset()
	s.stdout.Reset()
}

func (s *suite) TestMaybeHelp(c *gc.C) {
	type test struct {
		args   []string
		code   int
		exit   bool
		stdout string
		log    string
	}
	tests := []test{
		{
			args: []string{},
			code: 0,
			exit: true,
			log:  mainUsage,
		},
		{
			args: []string{"help"},
			code: 0,
			exit: true,
			log:  mainUsage,
		},
		{
			args: []string{"nothelp"},
			exit: false,
			log:  "",
		},
		{
			args: []string{"launch"},
			exit: false,
			log:  "",
		},
		{
			args: []string{"destroy", "foo"},
			exit: false,
			log:  "",
		},
		{
			args: []string{"help", "launch"},
			code: 0,
			exit: true,
			log:  launchUsage,
		},
		{
			args: []string{"help", "status"},
			code: 0,
			exit: true,
			log:  statusUsage,
		},
		{
			args: []string{"help", "destroy"},
			code: 0,
			exit: true,
			log:  destroyUsage,
		},
		{
			args: []string{"help", "destroy", "extraArgsOk"},
			code: 0,
			exit: true,
			log:  destroyUsage,
		},
		{
			args:   []string{"help", "notAcommand"},
			code:   1,
			exit:   true,
			stdout: "unknown command \"notAcommand\"\n",
			log:    "\n" + mainUsage,
		},
	}

	for i, t := range tests {
		c.Logf("%d. calling maybeHelp with %#v", i, t.args)
		code, exit := maybeHelp(t.args)
		c.Check(exit, gc.Equals, t.exit)
		// if we're not exiting, we shouldn't care about the code.
		if t.exit {
			c.Check(code, gc.Equals, t.code)
		}
		c.Check(s.log.String(), gc.Equals, t.log)
		c.Check(s.stdout.String(), gc.Equals, t.stdout)
		s.log.Reset()
		s.stdout.Reset()
	}
}

func (s *suite) TestMain(c *gc.C) {
	f := &fakeCmds{}

	for name, cmd := range cmds {
		cmd.fn = makeFn(f, name)
	}
	type test struct {
		args   []string
		code   int
		cmd    string
		stdout string
		log    string
	}
	tests := []test{
		{
			args: []string{"launch", "foo"},
			cmd:  "launch",
		},
		{
			args: []string{"status", "foo"},
			cmd:  "status",
		},
		{
			args: []string{"destroy", "foo"},
			cmd:  "destroy",
		},
		{
			args:   []string{"launch"},
			code:   1,
			log:    "\n" + launchUsage,
			stdout: "wrong number of arguments for cmd \"launch\"\n",
		},
		{
			args:   []string{"status"},
			code:   1,
			log:    "\n" + statusUsage,
			stdout: "wrong number of arguments for cmd \"status\"\n",
		},
		{
			args:   []string{"destroy"},
			code:   1,
			log:    "\n" + destroyUsage,
			stdout: "wrong number of arguments for cmd \"destroy\"\n",
		},
		{
			args:   []string{"launch", "foo", "bar"},
			code:   1,
			log:    "\n" + launchUsage,
			stdout: "wrong number of arguments for cmd \"launch\"\n",
		},
		{
			args:   []string{"status", "foo", "bar"},
			code:   1,
			log:    "\n" + statusUsage,
			stdout: "wrong number of arguments for cmd \"status\"\n",
		},
		{
			args:   []string{"destroy", "foo", "bar"},
			code:   1,
			log:    "\n" + destroyUsage,
			stdout: "wrong number of arguments for cmd \"destroy\"\n",
		},
		{
			args:   []string{"blah"},
			log:    "\n" + mainUsage,
			stdout: "unknown command \"blah\"\n",
			code:   1,
		},
		{
			args:   []string{"blah", "foo"},
			log:    "\n" + mainUsage,
			stdout: "unknown command \"blah\"\n",
			code:   1,
		},
		{
			args: []string{"launch", "code 1"},
			code: 1,
			cmd:  "launch",
		},
		{
			args: []string{"status", "code 1"},
			code: 1,
			cmd:  "status",
		},
		{
			args: []string{"destroy", "code 1"},
			code: 1,
			cmd:  "destroy",
		},
	}

	for i, t := range tests {
		c.Logf("%d. calling run with %#v", i, t.args)
		f.code = t.code
		code := run(t.args)
		c.Check(code, gc.Equals, t.code)
		if code == 0 {
			// if code is non-zero, we don't need to check args
			c.Check(f.arg, gc.Equals, t.args[1])
		}
		c.Check(s.log.String(), gc.Equals, t.log)
		c.Check(s.stdout.String(), gc.Equals, t.stdout)
		c.Check(f.name, gc.Equals, t.cmd)
		s.log.Reset()
		s.stdout.Reset()
		f.code = 0
		f.arg = ""
		f.name = ""
	}
}

type fakeCmds struct {
	name string
	arg  string
	code int
}

func makeFn(f *fakeCmds, name string) func(string) int {
	return func(arg string) int {
		f.name = name
		f.arg = arg
		return f.code
	}
}

// TODO(natefinch): update this when we make command into a slice.
// TODO(natefinch): update this when we start using portranges
const fakeProcJson = `{
	"Name": "unique",
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
}
`

func (s *suite) TestLaunch(c *gc.C) {
	details := docker.ProcDetails{
		ID: "unique",
		Status: docker.ProcStatus{
			Label: "Running",
		},
	}
	out := `{"id":"unique","status":{"label":"Running"}}
`
	type test struct {
		desc    string
		proc    string
		err     error
		details docker.ProcDetails
		code    int
		stdout  string
	}
	tests := []test{
		{
			desc:    "Default good case.",
			proc:    fakeProcJson,
			details: details,
			stdout:  out,
		},
		{
			desc:   "Bad JSON for charm.",
			proc:   "badjson",
			code:   1,
			stdout: "can't decode proc-info:.*\n",
		},
		{
			desc:   "Launch returns error.",
			proc:   fakeProcJson,
			err:    errors.New("foooo"),
			code:   1,
			stdout: "foooo\n",
		},
	}

	for i, t := range tests {
		c.Logf("%d. %s", i, t.desc)
		dockerLaunch = func(charm.Process) (docker.ProcDetails, error) {
			return t.details, t.err
		}
		code := launch(t.proc)
		c.Check(code, gc.Equals, t.code)
		if code == 0 {
			c.Check(s.stdout.String(), gc.Equals, t.stdout)
		} else {
			c.Check(s.stdout.String(), gc.Matches, t.stdout)
		}
		s.stdout.Reset()
	}
}

func (s *suite) TestStatus(c *gc.C) {
	type test struct {
		desc   string
		err    error
		status docker.ProcStatus
		code   int
		stdout string
	}
	tests := []test{
		{
			desc: "Default good case.",
			status: docker.ProcStatus{
				Label: "Running",
			},
			stdout: `{"label":"Running"}` + "\n",
		},
		{
			desc:   "Docker status error.",
			code:   1,
			err:    errors.New("foooo"),
			stdout: "foooo\n",
		},
	}

	for i, t := range tests {
		c.Logf("%d. %s", i, t.desc)
		dockerStatus = func(string) (docker.ProcStatus, error) {
			return t.status, t.err
		}
		code := status("foo")
		c.Check(code, gc.Equals, t.code)
		if code == 0 {
			c.Check(s.stdout.String(), gc.Equals, t.stdout)
		} else {
			c.Check(s.stdout.String(), gc.Matches, t.stdout)
		}
		s.stdout.Reset()

	}
}

func (s *suite) TestDestroy(c *gc.C) {
	type test struct {
		desc   string
		err    error
		code   int
		stdout string
	}
	tests := []test{
		{
			desc: "Default good case.",
		},
		{
			desc:   "Docker status error.",
			code:   1,
			err:    errors.New("foooo"),
			stdout: "foooo\n",
		},
	}

	for i, t := range tests {
		c.Logf("%d. %s", i, t.desc)
		dockerDestroy = func(string) error {
			return t.err
		}
		code := destroy("foo")
		c.Check(code, gc.Equals, t.code)
		if code == 0 {
			c.Check(s.stdout.String(), gc.Equals, t.stdout)
		} else {
			c.Check(s.stdout.String(), gc.Matches, t.stdout)
		}
		s.stdout.Reset()
	}
}
