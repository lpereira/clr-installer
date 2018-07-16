// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package network

import (
	"testing"
)

func TestGoodURL(t *testing.T) {

	if err := CheckURL("http://www.google.com"); err != nil {
		t.Fatal("Good HTTP URL failed")
	}

	if err := CheckURL("https://www.google.com"); err != nil {
		t.Fatal("Good HTTPS URL failed")
	}

	if err := CheckURL("https://cdn.download.clearlinux.org/update/"); err != nil {
		t.Fatal("Good Clear Linux HTTPS URL failed")
	}
}

func TestBadURL(t *testing.T) {

	if err := CheckURL("http://www.google.zonk"); err == nil {
		t.Fatal("Bad HTTP URL passed incorrectly")
	}

	if err := CheckURL("https://www.google.zonk"); err == nil {
		t.Fatal("Bad HTTPS URL passed incorrectly")
	}
}
