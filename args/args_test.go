// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package args

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/clearlinux/clr-installer/log"
)

var (
	testHTTPPort string
)

func init() {
	testHTTPPort = os.Getenv("TEST_HTTP_PORT")
}

func makeTestKernelCmd(cmd string) (string, error) {
	kernelCmd := []byte(cmd)
	tmpfile, err := ioutil.TempFile("/tmp", "kargTestCmd")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write(kernelCmd); err != nil {
		return tmpfile.Name(), err
	}
	if err := tmpfile.Close(); err != nil {
		return tmpfile.Name(), err
	}

	return tmpfile.Name(), nil
}

func serveHTTPDescFile(t *testing.T) (*http.Server, error) {
	srv := &http.Server{Addr: ":" + testHTTPPort}

	http.HandleFunc("/clr-installer.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{}")
	})

	go func() {
		_ = srv.ListenAndServe()
	}()

	return srv, nil
}

func TestKernelCmdInvalidFile(t *testing.T) {

	var testArgs Args
	var err error

	// Check for read error
	kernelCmdlineFile = "/proc/not-a-real-filename"

	err = testArgs.setKernelArgs()
	if err == nil {
		t.Errorf("Failed to detect a valid error reading kernel command")
		return
	}
}

func TestParseArgsKernelCmdInvalidFile(t *testing.T) {
	var testArgs Args
	var err error

	// Check for read error
	kernelCmdlineFile = "/proc/not-a-real-filename"

	err = testArgs.ParseArgs()
	if err == nil {
		t.Fatal("Failed to detect a valid error reading kernel command")
	}
}

func TestKernelCmdDemoTrue(t *testing.T) {

	var testArgs Args
	var kernelCmd string
	var err error

	// Check for Demo mode set true
	kernelCmd = "root=PARTUUID=694da991-29f6-4cbd-ab72-6da064a799c0 quiet modprobe.blacklist=ccipciedrv,aalbus,aalrms,aalrmc console=tty0 console=ttyS0,115200n8 init=/usr/lib/systemd/systemd-bootchart initcall_debug tsc=reliable no_timer_check noreplace-smp kvm-intel.nested=1 rootfstype=ext4,btrfs,xfs intel_iommu=igfx_off cryptomgr.notests rcupdate.rcu_expedited=1 i915.fastboot=1 rcu_nocbs=0-64 rw" + " " + kernelCmdlineDemo
	kernelCmdlineFile, err = makeTestKernelCmd(kernelCmd)
	defer func() {
		_ = os.Remove(kernelCmdlineFile)
	}()
	if err != nil {
		t.Errorf("Failed to makeTestKernelCmd with error %q", err)
		return
	}

	err = testArgs.setKernelArgs()
	if err != nil {
		t.Errorf("Failed to setKernelArgs with error %q", err)
		return
	}

	if testArgs.DemoMode == false {
		t.Errorf("Failed to detect Demo Mode TRUE with error kernel command %q", kernelCmd)
	}
}

func TestKernelCmdDemoFalse(t *testing.T) {

	var testArgs Args
	var kernelCmd string
	var err error

	// Check for Demo mode set false
	kernelCmd = "root=PARTUUID=694da991-29f6-4cbd-ab72-6da064a799c0 quiet modprobe.blacklist=ccipciedrv,aalbus,aalrms,aalrmc console=tty0 console=ttyS0,115200n8 init=/usr/lib/systemd/systemd-bootchart initcall_debug tsc=reliable no_timer_check noreplace-smp kvm-intel.nested=1 rootfstype=ext4,btrfs,xfs intel_iommu=igfx_off cryptomgr.notests rcupdate.rcu_expedited=1 i915.fastboot=1 rcu_nocbs=0-64 rw .demo"
	kernelCmdlineFile, err = makeTestKernelCmd(kernelCmd)
	defer func() {
		_ = os.Remove(kernelCmdlineFile)
	}()
	if err != nil {
		t.Errorf("Failed to makeTestKernelCmd with error %q", err)
		return
	}

	err = testArgs.setKernelArgs()
	if err != nil {
		t.Errorf("Failed to setKernelArgs with error %q", err)
		return
	}

	if testArgs.DemoMode == true {
		t.Errorf("Failed to detect Demo Mode FALSE with error kernel command %q", kernelCmd)
	}
}

func TestKernelCmdConfPresent(t *testing.T) {

	var testArgs Args
	var kernelCmd string
	var err error

	// Check for configuration file present
	kernelCmd = "root=PARTUUID=694da991-29f6-4cbd-ab72-6da064a799c0 quiet modprobe.blacklist=ccipciedrv,aalbus,aalrms,aalrmc console=tty0 console=ttyS0,115200n8 init=/usr/lib/systemd/systemd-bootchart initcall_debug tsc=reliable no_timer_check noreplace-smp kvm-intel.nested=1 rootfstype=ext4,btrfs,xfs intel_iommu=igfx_off cryptomgr.notests rcupdate.rcu_expedited=1 i915.fastboot=1 rcu_nocbs=0-64 rw" +
		" " + kernelCmdlineConf + "=http://google.com"
	kernelCmdlineFile, err = makeTestKernelCmd(kernelCmd)
	defer func() {
		_ = os.Remove(kernelCmdlineFile)
	}()
	if err != nil {
		t.Errorf("Failed to makeTestKernelCmd with error %q", err)
		return
	}

	err = testArgs.setKernelArgs()
	if err != nil {
		t.Errorf("Failed to setKernelArgs with error %q", err)
		return
	}

	if testArgs.ConfigFile == "" {
		t.Errorf("Failed to detect Configuration File with kernel command %q", kernelCmd)
	}
}

func TestKernelCmdInvalidProtocol(t *testing.T) {
	var testArgs Args
	var kernelCmd string
	var err error

	// the remote fetch supports only http protocol for now
	kernelCmd = kernelCmdlineConf + "=file:///proc/cmdline"
	kernelCmdlineFile, err = makeTestKernelCmd(kernelCmd)
	defer func() {
		_ = os.Remove(kernelCmdlineFile)
	}()
	if err != nil {
		t.Errorf("Failed to makeTestKernelCmd with error %q", err)
		return
	}

	err = testArgs.setKernelArgs()
	if err == nil {
		t.Errorf("setKernelArgs() should fail with unsupported protocol")
		return
	}
}

func TestKernelCmdValidFetch(t *testing.T) {

	var testArgs Args
	var kernelCmd string
	var err error

	// Check for configuration file present
	kernelCmd = kernelCmdlineConf + "=http://localhost:" + testHTTPPort + "/clr-installer.yaml"
	kernelCmdlineFile, err = makeTestKernelCmd(kernelCmd)
	defer func() {
		_ = os.Remove(kernelCmdlineFile)
	}()
	if err != nil {
		t.Errorf("Failed to makeTestKernelCmd with error %q", err)
		return
	}

	srv, err := serveHTTPDescFile(t)
	if err != nil {
		t.Fatalf("Failed to serve http desc file with error %q", err)
	}

	defer func() {
		if err = srv.Shutdown(nil); err != nil {
			t.Fatal(err)
		}
	}()

	err = testArgs.setKernelArgs()
	if err != nil {
		t.Errorf("Failed to setKernelArgs with error %q", err)
		return
	}

	if testArgs.ConfigFile == "" {
		t.Errorf("Failed to detect Configuration File with kernel command %q", kernelCmd)
	}
}

func TestKernelCmdConfEmpty(t *testing.T) {

	var testArgs Args
	var kernelCmd string
	var err error

	// Check for configuration file missing
	kernelCmd = "root=PARTUUID=694da991-29f6-4cbd-ab72-6da064a799c0 quiet modprobe.blacklist=ccipciedrv,aalbus,aalrms,aalrmc console=tty0 console=ttyS0,115200n8 init=/usr/lib/systemd/systemd-bootchart initcall_debug tsc=reliable no_timer_check noreplace-smp kvm-intel.nested=1 rootfstype=ext4,btrfs,xfs intel_iommu=igfx_off cryptomgr.notests rcupdate.rcu_expedited=1 i915.fastboot=1 rcu_nocbs=0-64 rw" +
		" " + "nothere"
	kernelCmdlineFile, err = makeTestKernelCmd(kernelCmd)
	defer func() {
		_ = os.Remove(kernelCmdlineFile)
	}()
	if err != nil {
		t.Errorf("Failed to makeTestKernelCmd with error %q", err)
		return
	}

	err = testArgs.setKernelArgs()
	if err != nil {
		t.Errorf("Failed to setKernelArgs with error %q", err)
		return
	}

	if testArgs.ConfigFile != "" {
		t.Errorf("Found Configuration File value when should be empty with kernel command %q", kernelCmd)
	}
}

func TestKernelAndCommandlineAllArgs(t *testing.T) {

	var testArgs Args
	var kernelCmd string
	var err error

	const confName = "command.conf"
	t.Logf("%v", os.Args)
	//os.Args = []string{os.Args[0], os.Args[1], "--config=" + confName, "--demo"}
	//os.Args = append(os.Args, "--config="+confName, "--demo")
	os.Args[1] = "--config=" + confName
	fmt.Println(os.Args)

	// Check for configuration file missing
	kernelCmd = "root=PARTUUID=694da991-29f6-4cbd-ab72-6da064a799c0 quiet modprobe.blacklist=ccipciedrv,aalbus,aalrms,aalrmc console=tty0 console=ttyS0,115200n8 init=/usr/lib/systemd/systemd-bootchart initcall_debug tsc=reliable no_timer_check noreplace-smp kvm-intel.nested=1 rootfstype=ext4,btrfs,xfs intel_iommu=igfx_off cryptomgr.notests rcupdate.rcu_expedited=1 i915.fastboot=1 rcu_nocbs=0-64 rw" +
		" " + kernelCmdlineConf + "=http://google.com"
	kernelCmdlineFile, err = makeTestKernelCmd(kernelCmd)
	defer func() {
		_ = os.Remove(kernelCmdlineFile)
	}()
	if err != nil {
		t.Errorf("Failed to makeTestKernelCmd with error %q", err)
		return
	}

	err = testArgs.ParseArgs()
	if err != nil {
		t.Errorf("Failed to ParseArgs with error %q", err)
		return
	}

	if testArgs.Version != false {
		t.Errorf("Command Line 'version' is not defaulted to 'false'")
	}
	if testArgs.Reboot != true {
		t.Errorf("Command Line 'reboot' is not defaulted to 'true'")
	}
	if testArgs.ConfigFile != confName {
		t.Errorf("Command Line 'config' is %q, NOT overridden to %q", testArgs.ConfigFile, confName)
	}
	if testArgs.SwupdMirror != "" {
		t.Errorf("Command Line 'mirror' is not defaulted to ''")
	}
	if testArgs.PamSalt != "" {
		t.Errorf("Command Line 'genpwd' is not defaulted to ''")
	}
	if testArgs.LogLevel != log.LogLevelDebug {
		t.Errorf("Command Line 'log-level' is not defaulted to '%d'", log.LogLevelDebug)
	}
	if testArgs.LogFile == "" {
		t.Errorf("Command Line 'log-file' is NOT set to value")
	}
}
