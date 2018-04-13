// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package updateengine

import (
	"fmt"

	"github.com/godbus/dbus"
)

const (
	RpmOstreeUpdateNone             = "RPM_OSTREE_UPDATE_NONE"
	RpmOstreeUpdateChecking         = "RPM_OSTREE_UPDATE_CHECKING"
	// We could further break this down into Downloading, Staging, Staged using
	// the download-only and cache-only options if we need that level of
	// flexibility.
	RpmOstreeUpdateStaged           = "RPM_OSTREE_UPDATE_STAGED"
	RpmOstreeUpdateError            = "RPM_OSTREE_UPDATE_ERROR"
)

type Status struct {
	CurrentStatus string
	// Let's just proxy "version" and "checksum" for now. AFAICT, this is only
	// for the benefit of a sysadmin doing `oc describe node`.
	NewVersion string
	NewChecksum string
}

func NewStatus(state string, cachedUpdate map[string]dbus.Variant) (s Status) {

	if cachedUpdate == nil || len(cachedUpdate) == 0 {
		return Status{CurrentStatus: state}
	}

	checksum := cachedUpdate["checksum"].Value().(string)

	versionProp, ok := cachedUpdate["version"]
	var version string
	if ok {
		version = versionProp.Value().(string)
	}

	return Status{state, version, checksum}
}

func (s *Status) String() string {
	return fmt.Sprintf("CurrentStatus=%s NewVersion=%s NewChecksum=%s",
	                   s.CurrentStatus, s.NewVersion, s.NewChecksum)
}
