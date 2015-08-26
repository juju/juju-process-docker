// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package docker

import (
	"fmt"
	"regexp"
	"strings"
)

// TODO(ericsnow) This file *could* move to a more generic library.

// TODO(ericsnow) Ensure compliance with SemVer?

// VersionInfo expresses the details of a version.
type VersionInfo struct {
	// Raw is the string on which the info is based, if available.
	Raw string
	// Major is the major version.
	Major int
	// Minor is the minor version.
	Minor int
	// Patch is the patch (or "micro") version.
	Patch int
}

var versionInfoRE = regexp.MustCompile(`[^\d]*(\d+)(?:\.(\d+)(?:\.(\d+))?)?.*`)

// ParseVersionInfo converts the provided version string into
// a new VersionInfo.
func ParseVersionInfo(vers string) (*VersionInfo, error) {
	vi := &VersionInfo{
		Raw: vers,
	}

	parts := versionInfoRE.FindStringSubmatch(vers)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid version %q", vers)
	}
	parts = parts[1:]
	for i, part := range parts {
		if part == "" {
			parts[i] = "0"
		}
	}

	actual := strings.Join(parts, ".")
	_, err := fmt.Sscanf(actual, "%d.%d.%d", &vi.Major, &vi.Minor, &vi.Patch)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q", vers)
	}
	return vi, nil
}

// String returns the string representation of the version.
func (vi VersionInfo) String() string {
	if vi.Raw != "" {
		return vi.Raw
	}
	return fmt.Sprintf("%d.%d.%d", vi.Major, vi.Minor, vi.Patch)
}

// Compare compares this VersionInfo to another. If they are the same
// then 0 is returned. If other is an earlier version then a negative
// value is returned. Otherwise a positive value is returned.
func (vi VersionInfo) Compare(other VersionInfo) int {
	if other.Major < vi.Major {
		return -1
	}
	if other.Major > vi.Major {
		return 1
	}

	if other.Minor < vi.Minor {
		return -1
	}
	if other.Minor > vi.Minor {
		return 1
	}

	if other.Patch < vi.Patch {
		return -1
	}
	if other.Patch > vi.Patch {
		return 1
	}

	return 0
}
