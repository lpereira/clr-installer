// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package tui

import (
	"github.com/VladimirMarkelov/clui"
)

// TelemetryPage is the Page implementation for the telemetry configuration page
type TelemetryPage struct {
	BasePage
}

const (
	telemetryHelp = `Enable Telemetry

Allow the Clear Linux OS for Intel Architecture to collect anonymous
reports to improve system stability? These reports only relate to
operating system details - no personally identifiable information is
collected.

See http://clearlinux.org/features/telemetry for more information.

Intel's privacy policy can be found at: http://www.intel.com/privacy.`
)

func newTelemetryPage(mi *Tui) (Page, error) {
	page := &TelemetryPage{
		BasePage: BasePage{
			// Tag this Page as required to be complete for the Install to proceed
			required: true,
		},
	}
	page.setupMenu(mi, TuiPageTelemetry, "Telemetry", BackButton|DoneButton, TuiPageMenu)

	lbl := clui.CreateLabel(page.content, 2, 11, telemetryHelp, Fixed)
	lbl.SetMultiline(true)

	page.backBtn.SetTitle("No, thanks")
	page.backBtn.SetSize(12, 1)

	page.doneBtn.SetTitle("Yes, enable telemetry!!")
	page.doneBtn.SetSize(25, 1)

	return page, nil
}

// DeActivate sets the model value and adjusts the "done" flag for this page
func (tp *TelemetryPage) DeActivate() {
	tp.getModel().EnableTelemetry(tp.action == ActionDoneButton)
	tp.SetDone(true)
}

// Activate activates the proper button depending on the current model value
// if telemetry is enabled in the data model then the done button will be active
// otherwise the back button will be activated.
func (tp *TelemetryPage) Activate() {
	if tp.getModel().IsTelemetryEnabled() {
		tp.activated = tp.doneBtn
	} else {
		tp.activated = tp.backBtn
	}
}
