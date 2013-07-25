// Copyright (c) 2013 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"text/template"
)

var (
	description = flag.String("message", "", "The description template that will be used for the commit on each split branch. See the README for template data information.")

	descriptionTpl *template.Template
)

type descriptionParams struct {
	SplitDirectory string
	AffectedFiles  []string
	ExtraData      string
}

func validateDescription() (err error) {
	if *description == "" {
		return fmt.Errorf("Description is empty.")
	}
	*description += `

===== Affected files: =====
{{range $_, $f := .AffectedFiles}}{{.}}
{{end}}
{{if .ExtraData}}
{{.ExtraData}}
{{end}}`
	descriptionTpl, err = template.New("").Parse(*description)
	return
}

func formatDescription(cl *changeList) string {
	params := descriptionParams{
		SplitDirectory: cl.base,
		AffectedFiles:  cl.paths,
		ExtraData:      cl.description,
	}

	buf := new(bytes.Buffer)
	err := descriptionTpl.Execute(buf, params)
	if err != nil {
		log.Print("Error formatting description for %s: %v", cl.base, err)
		// Fataling out in the middle of branch creation is not good, so just return a template.
		return *description
	}

	return buf.String()
}

// banner creates a textual separator with a given format string.
func banner(m string, fmtArgs ...interface{}) string {
	return fmt.Sprintf("===== %s =====\n", fmt.Sprintf(m, fmtArgs...))
}
