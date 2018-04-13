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
	"os"
	"strconv"
	"time"

	"github.com/godbus/dbus"
	"github.com/golang/glog"
)

const (
	dbusName          = "org.projectatomic.rpmostree1"
	dbusSysroot       = "/org/projectatomic/rpmostree1/Sysroot"
	dbusIfaceOS       = "org.projectatomic.rpmostree1.OS"
	dbusIfaceSysroot  = "org.projectatomic.rpmostree1.Sysroot"
	statePollInterval = 1 * time.Hour
)

type Client struct {
	conn   *dbus.Conn
	osPath dbus.ObjectPath
}

func New() (*Client, error) {
	c := new(Client)
	var err error

	c.conn, err = dbus.SystemBusPrivate()
	if err != nil {
		return nil, err
	}

	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}
	err = c.conn.Auth(methods)
	if err != nil {
		c.conn.Close()
		return nil, err
	}

	err = c.conn.Hello()
	if err != nil {
		c.conn.Close()
		return nil, err
	}

	// find booted OS
	sysroot := c.conn.Object(dbusName, dbus.ObjectPath(dbusSysroot))
	prop, err := sysroot.GetProperty(dbusMember(dbusIfaceSysroot, "Booted"))
	if err != nil {
		return nil, err
	}

	c.osPath = prop.Value().(dbus.ObjectPath)
	glog.Infof("Booted OS path is %v", c.osPath)
	return c, nil
}

// godbus works on interface.member notation
func dbusMember(iface, member string) string {
	return fmt.Sprintf("%s.%s", iface, member)
}

func dbusOSMember(member string) string {
	return fmt.Sprintf("%s.%s", dbusIfaceOS, member)
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// WatchForUpdate regularly checks for new rpm-ostree updates receives signal messages from dbus and sends them as Statuses
// on the rcvr channel, until the stop channel is closed. An attempt is made to
// get the initial status and send it on the rcvr channel before receiving
// starts.
func (c *Client) WatchForUpdate(rcvr chan Status, stop <-chan struct{}) {
	t := time.Tick(statePollInterval)
	for {
		select {
		case <-stop:
			return
		case <-t:
			c.checkForUpdate(rcvr)
		}
	}
}

// Check for an update, and stages the update if found.
func (c *Client) checkForUpdate(rcvr chan Status) {
	glog.Info("Checking for update")

	// We could go straight to Upgrade() here. The nice thing with using
	// AutomaticUpdateTrigger() first is to make it easier eventually to split
	// out "checking" from "downloading & staging".
	rcvr <- Status{CurrentStatus: RpmOstreeUpdateChecking}

	bootedOS := c.conn.Object(dbusName, c.osPath)
	err := callAutomaticUpdateTrigger(bootedOS)
	if err != nil {
		glog.Warningf("Error calling AutomaticUpdateTrigger(): %v", err)
		rcvr <- NewStatus(RpmOstreeUpdateError, nil)
		return
	}

	cachedUpdate, err := getCachedUpdate(bootedOS)
	if err != nil {
		glog.Warningf("Error reading CachedUpdate property: %v", err)
		rcvr <- NewStatus(RpmOstreeUpdateError, nil)
		return
	}

	if len(cachedUpdate) == 0 {
		// No updates, we're done
		rcvr <- NewStatus(RpmOstreeUpdateNone, nil)
		return
	}

	err = callUpgrade(bootedOS)
	if err != nil {
		glog.Warningf("Error calling Upgrade(): %v", err)
		rcvr <- NewStatus(RpmOstreeUpdateError, nil)
		return
	}

	// XXX: need to sanity check that a new deployment really is staged
	// XXX: race: should use deployment info for details, not cachedUpdate, in
	// case we somehow update past what we saw during AutomaticUpdateTrigger()
	// XXX: we'll have to make this smarter for auto-rollback eventually
	rcvr <- NewStatus(RpmOstreeUpdateStaged, cachedUpdate)
	return
}

func callAutomaticUpdateTrigger(bootedOS dbus.BusObject) error {
	options_map := make(map[string]dbus.Variant)
	options_map["mode"] = dbus.MakeVariant("check")
	var enabled bool
	var addr string
	err := bootedOS.Call(dbusOSMember("AutomaticUpdateTrigger"), 0, options_map).Store(&enabled, &addr)
	if err != nil {
		return err
	}

	return runTransactionSync(addr)
}

func getCachedUpdate(bootedOS dbus.BusObject) (map[string]dbus.Variant, error) {
	// Unlike GDBus, it seem like godbus doesn't do any property caching, so we
	// don't have to call Reload() right after a transaction.
	prop, err := bootedOS.GetProperty(dbusOSMember("CachedUpdate"))
	if err != nil {
		return nil, err
	}

	return prop.Value().(map[string]dbus.Variant), nil
}

func callUpgrade(bootedOS dbus.BusObject) error {
	options_map := make(map[string]dbus.Variant)
	var addr string
	err := bootedOS.Call(dbusOSMember("Upgrade"), 0, options_map).Store(&addr)
	if err != nil {
		return err
	}

	return runTransactionSync(addr)
}

// D-Bus rpm-ostree transaction goop helper
func runTransactionSync(address string) error {

	transaction_conn, err := dbus.Dial(address)
	if err != nil {
		return err
	}
	defer transaction_conn.Close()

	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}
	err = transaction_conn.Auth(methods)
	if err != nil {
		return err
	}

	transaction := transaction_conn.Object("org.projectatomic.rpmostree1", "/")
	signalCh := make(chan *dbus.Signal, 1)
	transaction_conn.Signal(signalCh)

	err = transaction.Call("org.projectatomic.rpmostree1.Transaction.Start", 0).Err
	if err != nil {
		return err
	}

	var success bool
	var errMsg string
	for signal := range signalCh { // XXX: should probably have a timeout here
		if signal.Name == "org.projectatomic.rpmostree1.Transaction.Finished" {
			success = signal.Body[0].(bool)
			errMsg = signal.Body[1].(string)
			break
		}
	}

	if !success {
		return fmt.Errorf(errMsg)
	}

	return nil
}
