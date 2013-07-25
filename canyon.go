// Copyright (c) 2013 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

/*
canyon is a tool for making big splits. This assumes that you have a very large
changelist prepared in a single branch. canyon will then split up the large
change into multiple branches, by OWNERS file, and then will prepare a changelist
description.
*/

const kMaxDepth = 1

func main() {
	branch := strings.TrimSpace(gitOrDie("symbolic-ref", "--short", "HEAD"))

	fmt.Printf("Split changelist on branch %q into sub-changelists? [y/N] ", branch)
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	if buf[0] != 'y' {
		fmt.Println("Exiting")
		return
	}

	log.Print("Gathering changed files")
	files := strings.Split(gitOrDie("diff", "--name-only", "origin/master"), "\n")

	log.Print("Splitting changed files into groups for changelists")
	cs := newChangeSet(branch)
	for _, file := range files {
		if file == "" {
			continue
		}
		cs.splitByFile("OWNERS", file)
	}

	log.Print("Creating branches for splits")
	for _, cl := range cs.splits {
		splitBranch := cl.branchName(branch)
		log.Printf("Preparing branch %s", splitBranch)

		_, err := git("checkout", "-b", splitBranch, "origin/master")
		if err != nil {
			log.Printf("Failed to create new branch %q: %v", splitBranch, err)
			continue
		}

		_, err = git("checkout", branch, cl.base)
		if err != nil {
			log.Print("Failed to check out subdirectory from root branch")
			gitOrDie("reset", "--hard", "origin/master")
			continue
		}

		desc := fmt.Sprintf("Update include paths in %s for base/process changes.\n\nBUG=242290", cl.base)
		desc += "\n\n===== Affected paths: =====\n"
		for _, file := range cl.paths {
			desc += fmt.Sprintf("%s\n", file)
		}
		desc += "\n" + cl.description

		_, err = git("commit", "-a", "-m", desc)
		if err != nil {
			log.Print("Failed to create subchangelist")
			gitOrDie("reset", "--hard", "origin/master")
			continue
		}
	}

	git("checkout", branch)
}

// git runs the specified git commands and returns the output as a string,
// blocking to completion.
func git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(stdout), nil
}

// gitOrDie runs the git command and panics on failure.
func gitOrDie(args ...string) string {
	r, e := git(args...)
	if e != nil {
		panic(e.Error())
	}
	return r
}
