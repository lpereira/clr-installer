// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"fmt"

	"github.com/VladimirMarkelov/clui"
	"github.com/clearlinux/clr-installer/kernel"
)

// KernelPage is the Page implementation for the proxy configuration page
type KernelPage struct {
	BasePage
	kernels []*KernelRadio
	group   *clui.RadioGroup
}

// KernelRadio maps a map name and description with the actual checkbox
type KernelRadio struct {
	kernel *kernel.Kernel
	radio  *clui.Radio
}

// Activate marks selects the kernel radio based on the data model
func (kp *KernelPage) Activate() {
	model := kp.getModel()

	for _, curr := range kp.kernels {
		if !curr.kernel.Equals(model.Kernel) {
			continue
		}

		kp.group.SelectItem(curr.radio)
		break
	}
}

// GetConfigDefinition returns if the config was interactively defined by the user,
// was loaded from a config file or if the config is not set.
func (kp *KernelPage) GetConfigDefinition() int {
	k := kp.getModel().Kernel

	if k == nil {
		return ConfigNotDefined
	} else if k.IsUserDefined() {
		return ConfigDefinedByUser
	}

	return ConfigDefinedByConfig
}

func newKernelPage(mi *Tui) (Page, error) {
	page := &KernelPage{kernels: []*KernelRadio{}}

	kernels, err := kernel.LoadKernelList()
	if err != nil {
		return nil, err
	}

	for _, curr := range kernels {
		page.kernels = append(page.kernels, &KernelRadio{curr, nil})
	}

	page.setupMenu(mi, TuiPageKernel, "Kernel Selection", NoButtons)
	clui.CreateLabel(page.content, 2, 2, "Select desired kernel", Fixed)

	frm := clui.CreateFrame(page.content, AutoSize, AutoSize, BorderNone, Fixed)
	frm.SetPack(clui.Vertical)

	lblFrm := clui.CreateFrame(frm, AutoSize, AutoSize, BorderNone, Fixed)
	lblFrm.SetPack(clui.Vertical)
	lblFrm.SetPaddings(2, 0)

	page.group = clui.CreateRadioGroup()

	for _, curr := range page.kernels {
		lbl := fmt.Sprintf("%s: %s", curr.kernel.Name, curr.kernel.Desc)
		curr.radio = clui.CreateRadio(lblFrm, AutoSize, lbl, AutoSize)
		curr.radio.SetPack(clui.Horizontal)
		page.group.AddItem(curr.radio)
	}

	fldFrm := clui.CreateFrame(frm, 30, AutoSize, BorderNone, Fixed)
	fldFrm.SetPack(clui.Vertical)

	cancelBtn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Cancel", Fixed)
	cancelBtn.OnClick(func(ev clui.Event) {
		mi.gotoPage(TuiPageMenu, mi.currPage)
	})

	confirmBtn := CreateSimpleButton(page.cFrame, AutoSize, AutoSize, "Confirm", Fixed)
	confirmBtn.OnClick(func(ev clui.Event) {
		selected := page.group.Selected()
		page.getModel().Kernel = page.kernels[selected].kernel
		page.SetDone(true)
		mi.gotoPage(TuiPageMenu, mi.currPage)
	})

	return page, nil
}