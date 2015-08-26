// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju-process-docker/docker"
)

var _ = gc.Suite(&versionInfoSuite{})

type versionInfoSuite struct{}

func (versionInfoSuite) TestZeroValue(c *gc.C) {
	var vi docker.VersionInfo

	c.Check(vi, jc.DeepEquals, docker.VersionInfo{
		Major: 0,
		Minor: 0,
		Patch: 0,
	})
}

func (versionInfoSuite) TestParseVersionInfoOkay(c *gc.C) {
	tests := map[string]docker.VersionInfo{
		"1.3.2": docker.VersionInfo{Major: 1, Minor: 3, Patch: 2},
		"1.3":   docker.VersionInfo{Major: 1, Minor: 3, Patch: 0},
		"1":     docker.VersionInfo{Major: 1, Minor: 0, Patch: 0},

		"abc1.3.2xyz": docker.VersionInfo{Major: 1, Minor: 3, Patch: 2},
		"abc1.3xyz":   docker.VersionInfo{Major: 1, Minor: 3, Patch: 0},
		"abc1xyz":     docker.VersionInfo{Major: 1, Minor: 0, Patch: 0},
	}

	for vers, expected := range tests {
		c.Logf("checking %q", vers)
		vi, err := docker.ParseVersionInfo(vers)
		if !c.Check(err, jc.ErrorIsNil) {
			continue
		}

		expected.Raw = vers
		c.Check(vi, jc.DeepEquals, &expected)
	}
}

func (versionInfoSuite) TestParseVersionInfoEmpty(c *gc.C) {
	_, err := docker.ParseVersionInfo("")

	c.Check(err, gc.ErrorMatches, `invalid version.*`)
}

func (versionInfoSuite) TestParseVersionInfoInvalid(c *gc.C) {
	_, err := docker.ParseVersionInfo("spam")

	c.Check(err, gc.ErrorMatches, `invalid version.*`)
}

func (versionInfoSuite) TestStringFull(c *gc.C) {
	vi := docker.VersionInfo{Major: 1, Minor: 3, Patch: 2}
	vers := vi.String()

	c.Check(vers, gc.Equals, "1.3.2")
}

func (versionInfoSuite) TestStringMinimal(c *gc.C) {
	vi := docker.VersionInfo{Major: 1}
	vers := vi.String()

	c.Check(vers, gc.Equals, "1.0.0")
}

func (versionInfoSuite) TestCompareEqual(c *gc.C) {
	vi := docker.VersionInfo{Major: 1, Minor: 3, Patch: 2}
	other := docker.VersionInfo{Major: 1, Minor: 3, Patch: 2}
	compared := vi.Compare(other)

	c.Check(compared, gc.Equals, 0)
}

func (versionInfoSuite) TestCompareGreater(c *gc.C) {
	vi := docker.VersionInfo{Major: 1, Minor: 3, Patch: 2}
	c.Logf("comparing to %#v", vi)
	check := func(other docker.VersionInfo) {
		c.Logf("checking %#v", other)
		compared := vi.Compare(other)

		c.Check(compared, jc.GreaterThan, 0)
	}

	others := [][]int{
		{1, 3, 3},
		{1, 4, 0},
		{1, 4, 1},
		{1, 4, 2},
		{1, 4, 3},
		{2, 0, 0},
		{2, 0, 1},
		{2, 0, 2},
		{2, 0, 3},
		{2, 1, 0},
		{2, 1, 1},
		{2, 1, 2},
		{2, 1, 3},
		{2, 2, 0},
		{2, 2, 1},
		{2, 2, 2},
		{2, 2, 3},
		{2, 4, 0},
		{2, 4, 1},
		{2, 4, 2},
		{2, 4, 3},
	}
	for _, other := range others {
		check(docker.VersionInfo{
			Major: other[0],
			Minor: other[1],
			Patch: other[2],
		})
	}
}

func (versionInfoSuite) TestCompareLess(c *gc.C) {
	vi := docker.VersionInfo{Major: 1, Minor: 3, Patch: 2}
	c.Logf("comparing to %#v", vi)
	check := func(other docker.VersionInfo) {
		c.Logf("checking %#v", other)
		compared := vi.Compare(other)

		c.Check(compared, jc.LessThan, 0)
	}

	others := [][]int{
		{1, 3, 1},
		{1, 3, 0},
		{1, 2, 0},
		{1, 2, 1},
		{1, 2, 2},
		{1, 2, 3},
		{1, 1, 0},
		{1, 1, 1},
		{1, 1, 2},
		{1, 1, 3},
		{1, 0, 0},
		{1, 0, 1},
		{1, 0, 2},
		{1, 0, 3},
		{0, 0, 0},
		{0, 0, 1},
		{0, 0, 2},
		{0, 0, 3},
		{0, 1, 0},
		{0, 1, 1},
		{0, 1, 2},
		{0, 1, 3},
		{0, 2, 0},
		{0, 2, 1},
		{0, 2, 2},
		{0, 2, 3},
		{0, 4, 0},
		{0, 4, 1},
		{0, 4, 2},
		{0, 4, 3},
	}
	for _, other := range others {
		check(docker.VersionInfo{
			Major: other[0],
			Minor: other[1],
			Patch: other[2],
		})
	}
}
