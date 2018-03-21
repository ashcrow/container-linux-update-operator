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
)

// The possible update statuses returned from the update engine
//
// These correspond to current operation values exposed over DBus and defined by `update_engine`:
// https://github.com/coreos/update_engine/blob/v0.4.3/src/update_engine/update_attempter.h#L34-L43
// TODO: Update to match new update backend
const (
	UpdateStatusIdle                = "UPDATE_STATUS_IDLE"
	UpdateStatusCheckingForUpdate   = "UPDATE_STATUS_CHECKING_FOR_UPDATE"
	UpdateStatusUpdateAvailable     = "UPDATE_STATUS_UPDATE_AVAILABLE"
	UpdateStatusDownloading         = "UPDATE_STATUS_DOWNLOADING"
	UpdateStatusVerifying           = "UPDATE_STATUS_VERIFYING"
	UpdateStatusFinalizing          = "UPDATE_STATUS_FINALIZING"
	UpdateStatusUpdatedNeedReboot   = "UPDATE_STATUS_UPDATED_NEED_REBOOT"
	UpdateStatusReportingErrorEvent = "UPDATE_STATUS_REPORTING_ERROR_EVENT"
)

// Status represents the current status of an update
// TODO: This doesn't map to the current rpm-ostree dbus types as well as
// I'd like. Should a new type be created or should this end up wrapping
// multiple smaller types?
type Status struct {
	OSName           string
	Checksum         string
	Version          string
	Timestamp        int64
	Origin           string
	Signatures       []string
	GPGEnabled       bool
	RefHasNewCommit  bool
	RPMDiff          map[string]string
	CurrentOperation string
	//Advisories  a(suuasa{sv})
	// LastCheckedTime  int64
	// Progress         float64
	// CurrentOperation string
	// NewVersion       string
	// NewSize          int64
}

// NewStatus creates a new status
func NewStatus(body []interface{}) (s Status) {
	s.OSName = body[0].(string)
	s.Checksum = body[1].(string)
	s.Version = body[2].(string)
	s.Timestamp = body[3].(int64)
	s.Origin = body[4].(string)
	s.Signatures = body[5].([]string)
	s.GPGEnabled = body[6].(bool)
	s.RefHasNewCommit = body[7].(bool)
	s.RPMDiff = body[8].(map[string]string)
	s.CurrentOperation = "" // TODO
	return
}

// String turns the associated Status into a string
func (s *Status) String() string {
	return fmt.Sprintf("Timestamp=%v CurrentOperation=%v NewVersion=%v Checksum=%v",
		s.Timestamp,
		s.CurrentOperation,
		s.Version,
		s.Checksum,
	)
}
