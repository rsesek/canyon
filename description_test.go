// Copyright (c) 2013 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"testing"
)

func setDescription(s string) {
	*description = s
	descriptionTpl = nil
}

func TestValidateEmpty(t *testing.T) {
	setDescription("")
	if err := validateDescription(); err == nil {
		t.Error("Expected error for having an empty description, got nil")
	}
}

func TestFormatNoExtra(t *testing.T) {
	setDescription("Simple change for {{.SplitDirectory}}!\n\nBUG=1234")
	if err := validateDescription(); err != nil {
		t.Errorf("Unexpected error in validating description: %v", err)
		return
	}

	cl := &changeList{
		base: "chrome/browser",
		paths: []string{
			"chrome/browser/browser.cc",
			"chrome/browser/browser_window.h",
		},
	}
	actual := formatDescription(cl)
	expected := `Simple change for chrome/browser!

BUG=1234

===== Affected files: =====
chrome/browser/browser.cc
chrome/browser/browser_window.h

`
	testActualExpected(t, actual, expected)
}

func TestFormatWithExtra(t *testing.T) {
	setDescription("Don't have a cow, {{.SplitDirectory}}/")
	if err := validateDescription(); err != nil {
		t.Errorf("Unexpected error in validating description: %v", err)
		return
	}

	cl := &changeList{
		base: "base/process",
		paths: []string{
			"base/process/memory.h",
			"base/process/kill.h",
			"base/process/launch.h",
		},
		description: banner("HELLO") + "Tests are fun.\nAren't they?",
	}
	actual := formatDescription(cl)
	expected := `Don't have a cow, base/process/

===== Affected files: =====
base/process/memory.h
base/process/kill.h
base/process/launch.h


===== HELLO =====
Tests are fun.
Aren't they?
`
	testActualExpected(t, actual, expected)
}

func testActualExpected(t *testing.T, actual, expected string) {
	if actual != expected {
		t.Error("Actual formatted description does not match expected")
		t.Error("Expected:")
		t.Error(expected)
		t.Error("Actual:")
		t.Error(actual)
	}
}
