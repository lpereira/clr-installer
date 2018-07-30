// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package telemetry

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/network"
	"github.com/clearlinux/clr-installer/utils"
)

const (
	// Default Telemetry configuration file
	defaultTelemetryConf = "/usr/share/defaults/telemetrics/telemetrics.conf"
	customTelemetryConf  = "/etc/telemetrics/telemetrics.conf"

	// Title is a predefined text to display on the Telemetry
	Title = `Enable Telemetry`
	// Help is a predefined text to display on the Telemetry
	// screen for interactive installations
	Help = `Allow the Clear Linux OS for Intel Architecture to collect anonymous
reports to improve system stability? These reports only relate to
operating system details - no personally identifiable information is
collected.

See http://clearlinux.org/features/telemetry for more information.
`

	// RequestNotice is a common text string to be displayed when enabling
	// telemetry by default on local networks
	RequestNotice = "NOTICE: Enabling Telemetry preferred by default on internal networks"

	// Default Telemetry server
	defaultTelemtryServer = "clr.telemetry.intel.com"
)

var (
	// Policy is the default Telemetry policy to be displayed
	// during interactive installations. Overridden by command line or
	// configuration options
	Policy = "Intel's privacy policy can be found at: http://www.intel.com/privacy."

	serverExp = regexp.MustCompile(`(?im)^(\s*server\s*=\s*)(\S+)(\s*)$`)
	tidExp    = regexp.MustCompile(`(?im)^(\s*tidheader\s*=\s*X-Telemetry-TID\s*:\s*)(\S+)(\s*)$`)
)

// Telemetry represents the target system telemetry enabling flag
type Telemetry struct {
	Enabled     bool
	Defined     bool
	URL         string
	TID         string
	requested   bool
	server      string
	userDefined bool
}

// IsUserDefined returns true if the configuration was interactively
// defined by the user
func (tl *Telemetry) IsUserDefined() bool {
	return tl.userDefined
}

// SetRequested set the Requested flag
func (tl *Telemetry) SetRequested(requested bool) {
	tl.requested = requested
}

// IsRequested returns true if we are requested telemetry be enabled
func (tl *Telemetry) IsRequested() bool {
	return tl.requested
}

// MarshalYAML marshals Telemetry into YAML format
func (tl *Telemetry) MarshalYAML() (interface{}, error) {
	return tl.Enabled, nil
}

// UnmarshalYAML unmarshals Telemetry from YAML format
func (tl *Telemetry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var enabled bool

	if err := unmarshal(&enabled); err != nil {
		return err
	}

	tl.Enabled = enabled
	tl.userDefined = false
	return nil
}

// SetEnable sets the enabled flag and sets this is an user defined configuration
func (tl *Telemetry) SetEnable(enable bool) {
	tl.Enabled = enable
	tl.userDefined = true

	if tl.server == "" {
		tl.server = defaultTelemtryServer
	}
}

// SetTelemetryServer set new defaults for the Telemetry server
// to override the built-in defaults
func (tl *Telemetry) SetTelemetryServer(telmURL string, telmID string, telmPolicy string) error {
	u, err := url.Parse(telmURL)
	if err != nil {
		return fmt.Errorf("Could not determine provide telemetry server name from URL (%s): %v", telmURL, err)
	}

	if urlErr := network.CheckURL(telmURL); urlErr != nil {
		return fmt.Errorf("Server not responding")
	}

	log.Debug("Using Telemetry URL: %q with TID: %q", telmURL, telmID)
	tl.URL = telmURL
	tl.TID = telmID
	tl.server = u.Hostname()

	// Set the policy
	Policy = telmPolicy

	return nil
}

// IsUsingPrivateIP return true if the current image is resolving
// the Telemetry server to a Private network IP address
func (tl *Telemetry) IsUsingPrivateIP() bool {
	inside := false

	if ips, err := net.LookupIP(tl.server); err == nil {
		// Create networks for all known Private Networks
		_, ipNetPriv10, _ := net.ParseCIDR("10.0.0.0/8")
		_, ipNetPriv172, _ := net.ParseCIDR("172.16.0.0/12")
		_, ipNetPriv192, _ := net.ParseCIDR("192.168.0.0/16")

		for _, ip := range ips {
			if ip.DefaultMask() == nil {
				log.Warning("PrivateIP: Ignoring non-IPv4 IP address: %s", ip)
				continue
			}

			in := ipNetPriv10.Contains(ip) ||
				ipNetPriv172.Contains(ip) ||
				ipNetPriv192.Contains(ip)
			log.Debug("PrivateIP: Found IP: %s, Private IP?: %s", ip, strconv.FormatBool(in))
			if in {
				inside = true
			}
		}
	} else {
		log.Warning("PrivateIP: Could not determine network location: %v", err)
	}

	return inside
}

// CreateTelemetryConf copies the contents of the log to the given filename
func (tl *Telemetry) CreateTelemetryConf(rootDir string) error {

	defConfFile := filepath.Join(rootDir, defaultTelemetryConf)
	// Make sure we can read the default Telemetry configuration file
	defConf, readErr := ioutil.ReadFile(defConfFile)
	if readErr != nil {
		return readErr
	}

	// Ensure the customer configuration file directory exists
	targetConfFile := filepath.Join(rootDir, customTelemetryConf)
	targetConfDir := filepath.Dir(targetConfFile)
	if err := utils.MkdirAll(targetConfDir); err != nil {
		return err
	}

	// Replace the server
	targetConf := serverExp.ReplaceAll(defConf, []byte("${1}"+tl.URL+"${3}"))
	// Replace the server
	targetConf = tidExp.ReplaceAll(targetConf, []byte("${1}"+tl.TID+"${3}"))

	// Write the new file
	writeErr := ioutil.WriteFile(targetConfFile, targetConf, 0644)
	if writeErr != nil {
		return writeErr
	}

	log.Debug("Created Telemetry server configuration file with URL %q and tag %q", tl.URL, tl.TID)

	return nil
}
