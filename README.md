# juju-process-docker
Support for docker in Juju core (github.com/juju/juju).

This includes the following:

* a top-level "docker" package containing the actual code
* docker.Client: an interface that provides the docker functionality
  needed by Juju
* docker.CLIClient: a simple Client implementation that wraps calling
  exec'ing the docker CLI
* docker.Info: a light wrapper around the Go type that older versions
  of docker use for the output of the "docker inspect" command
  (see "github.com/docker/docker/api/types".ContainerJSONPre120)
